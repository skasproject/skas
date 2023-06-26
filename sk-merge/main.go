package main

import (
	"context"
	"fmt"
	"github.com/pior/runnable"
	"os"
	"skas/sk-common/pkg/clientauth"
	cconfig "skas/sk-common/pkg/config"
	"skas/sk-common/pkg/skserver"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/pkg/skserver/protector"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-merge/internal/config"
	"skas/sk-merge/internal/handlers"
	"skas/sk-merge/internal/identitygetter"
	"skas/sk-merge/internal/providerchain"
)

func main() {
	if err := config.Setup(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to load configuration: %v\n", err)
		os.Exit(2)
	}
	config.Log.Info("sk-merge start", "version", cconfig.Version, "build", cconfig.BuildTs, "logLevel", config.Conf.Log.Level)

	// providerChain is based on http.client, which is thread safe. So, can be shared by all servers
	providerChain, err := providerchain.New(config.Log.WithName("providerChain"))
	if err != nil {
		config.Log.Error(err, "Error on clientProviderChain creation")
		os.Exit(6)
	}
	// identityGetter is based on http.client, which is thread safe. So, can be shared by all servers
	identityGetter := identitygetter.New(providerChain, config.Log)
	runnableMgr := runnable.NewManager()

	for idx, serverConfig := range config.Conf.Servers {
		server := skserver.New(fmt.Sprintf("server[%d]", idx), &config.Conf.Servers[idx].SkServerConfig, config.Log.WithName(fmt.Sprintf("mergeServer[%d]", idx)))
		// --------------------- Identity handler
		if !serverConfig.Services.Identity.Disabled {
			hdl := &commonHandlers.IdentityHandler{
				IdentityGetter: identityGetter,
				ClientManager:  clientauth.New(serverConfig.Services.Identity.Clients, serverConfig.Interface != "127.0.0.1"),
				Protector:      protector.New(serverConfig.Services.Identity.Protected, context.Background(), config.Log.WithName("sk-merge.identity.protector")),
			}
			server.AddHandler(proto.IdentityMeta, hdl)
		} else {
			server.GetLog().Info("'identity' service disabled")
		}
		// --------------------- PasswordChange handler
		if !serverConfig.Services.PasswordChange.Disabled {
			hdl := &handlers.PasswordChangeHandler{
				Chain:         providerChain,
				ClientManager: clientauth.New(serverConfig.Services.PasswordChange.Clients, serverConfig.Interface != "127.0.0.1"),
				Protector:     protector.New(serverConfig.Services.PasswordChange.Protected, context.Background(), config.Log.WithName("sk-merge.passwordChange.protector")),
			}
			server.AddHandler(proto.PasswordChangeMeta, hdl)
		} else {
			server.GetLog().Info("'passwordChange' service disabled")
		}
		runnableMgr.Add(server)
	}
	runnable.Run(runnableMgr.Build())

}
