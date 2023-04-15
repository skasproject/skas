package main

import (
	"context"
	"fmt"
	"github.com/pior/runnable"
	"os"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/skserver"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-static/internal/config"
	"skas/sk-static/internal/identitygetter"
	"skas/sk-static/internal/users"
	"skas/sk-static/pkg/filewatcher"
)

func main() {
	if err := config.Setup(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to load configuration: %v\n", err)
		os.Exit(2)
	}

	usersWatcher, err := filewatcher.New(config.Conf.UsersFile, users.Parse, config.Log.WithName("usersFile watcher"))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to load users file: %v\n", err)
		os.Exit(2)
	}
	content := usersWatcher.GetContent().(*users.Content)
	config.Log.Info("sk-static start", "version", config.Version, "nbUsers", len(content.UserByLogin), "nbrGroupBindings", content.GroupBindingCount, "logLevel", config.Conf.Log.Level)

	go func() {
		if err := usersWatcher.Run(context.Background()); err != nil {
			config.Log.Error(err, "users file watcher error")
		}
	}()

	runnableMgr := runnable.NewManager()
	identityGetter := identitygetter.New(usersWatcher, config.Log.WithName("staticProvider"))

	for idx, serverConfig := range config.Conf.Servers {
		server := skserver.New(fmt.Sprintf("server[%d]", idx), &config.Conf.Servers[idx].SkServerConfig, config.Log.WithName(fmt.Sprintf("staticServer[%d]", idx)))
		if !serverConfig.Services.Identity.Disabled {
			// --------------------- Identity handler
			hdl := &commonHandlers.IdentityHandler{
				IdentityGetter: identityGetter,
				ClientManager:  clientauth.New(serverConfig.Services.Identity.Clients, serverConfig.Interface != "127.0.0.1"),
			}
			server.AddHandler(proto.IdentityMeta, hdl)
		} else {
			server.GetLog().Info("'identity' service disabled")
		}
		runnableMgr.Add(server)
	}
	runnable.Run(runnableMgr.Build())
}
