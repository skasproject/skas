package main

import (
	"context"
	"fmt"
	"github.com/pior/runnable"
	"os"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/datawatcher"
	"skas/sk-common/pkg/datawatcher/cmwatcher"
	"skas/sk-common/pkg/datawatcher/filewatcher"
	"skas/sk-common/pkg/skserver"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-static/internal/config"
	"skas/sk-static/internal/identitygetter"
	"skas/sk-static/internal/users"
)

func main() {
	if err := config.Setup(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to load configuration: %v\n", err)
		os.Exit(2)
	}
	runnableMgr := runnable.NewManager()

	var watcher datawatcher.DataWatcher
	var err error
	if config.Conf.UsersFile != "" {
		watcher, err = filewatcher.New(config.Conf.UsersFile, users.Parse, config.Log.WithName("fileWatcher"))
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to start users file watcher: %v\n", err)
			os.Exit(2)
		}
	} else if config.Conf.UsersConfigMap != "" {
		watcher, err = cmwatcher.New(context.Background(), config.Conf.UsersConfigMap, "users.yaml", users.Parse, config.Log.WithName("cmWatcher"), config.Conf.CmLocation.Namespace, config.Conf.CmLocation.Kubeconfig)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to start users configMap watcher: %v\n", err)
			os.Exit(2)
		}
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Netheir usersFile or usersConfifMap was defined \n")
		os.Exit(2)
	}
	runnableMgr.Add(watcher)
	content := watcher.Get().(*users.Content)
	config.Log.Info("sk-static start", "version", config.Version, "nbUsers", len(content.UserByLogin), "nbrGroupBindings", content.GroupBindingCount, "logLevel", config.Conf.Log.Level)

	identityGetter := identitygetter.New(watcher, config.Log.WithName("staticProvider"))

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
