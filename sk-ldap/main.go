package main

import (
	"context"
	"fmt"
	"github.com/pior/runnable"
	"os"
	"path/filepath"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/skserver"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/pkg/skserver/protector"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-ldap/internal/config"
	"skas/sk-ldap/internal/identitygetter"
)

func main() {
	if err := config.Setup(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to load configuration: %v\n", err)
		os.Exit(2)
	}
	config.Log.Info("sk-ldap start", "ldapServer", config.Conf.Ldap.Host, "version", config.Version, "build", config.BuildTs, "logLevel", config.Conf.Log.Level)

	runnableMgr := runnable.NewManager()

	for idx, serverConfig := range config.Conf.Servers {
		server := skserver.New(fmt.Sprintf("server[%d]", idx), &config.Conf.Servers[idx].SkServerConfig, config.Log.WithName(fmt.Sprintf("ldapServer[%d]", idx)))
		// --------------------- Identity handler
		if !serverConfig.Services.Identity.Disabled {
			// We re-instantiate one identityGetter per process, as not sure ldap client is thread safe (https://github.com/go-ldap/ldap/issues/130)
			identityGetter, err := identitygetter.New(&config.Conf.Ldap, config.Log, filepath.Dir(config.File))
			if err != nil {
				config.Log.Error(err, "ldap config")
				os.Exit(3)
			}
			hdl := &commonHandlers.IdentityHandler{
				IdentityGetter: identityGetter,
				ClientManager:  clientauth.New(serverConfig.Services.Identity.Clients, serverConfig.Interface != "127.0.0.1"),
				Protector:      protector.New(serverConfig.Services.Identity.Protected, context.Background(), config.Log.WithName("sk-ldap.identity.protector")),
			}
			server.AddHandler(proto.IdentityMeta, hdl)
		} else {
			server.GetLog().Info("'identity' service disabled")
		}
		runnableMgr.Add(server)
	}
	runnable.Run(runnableMgr.Build())
}
