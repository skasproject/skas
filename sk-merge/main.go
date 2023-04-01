package main

import (
	"context"
	"fmt"
	"os"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/skserver"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
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
	config.Log.Info("sk-merge start", "version", config.Version, "logLevel", config.Conf.Log.Level)

	server := skserver.New("merge", &config.Conf.Server, config.Log.WithName(fmt.Sprintf("%s", "mergeServer")))

	providerChain, err := providerchain.New(config.Log.WithName("providerChain"))
	if err != nil {
		config.Log.Error(err, "Error on clientProviderChain creation")
		os.Exit(6)
	}
	// --------------------- Identity handler
	if !config.Conf.Services.Identity.Disabled {
		hdl := &commonHandlers.IdentityHandler{
			IdentityGetter: identitygetter.New(providerChain, config.Log),
			ClientManager:  clientauth.New(config.Conf.Services.Identity.Clients, true),
		}
		server.AddHandler(proto.IdentityMeta, hdl)
	} else {
		config.Log.Info("userIdentity service disabled")
	}
	// --------------------- PasswordChange handler
	if !config.Conf.Services.PasswordChange.Disabled {
		hdl := &handlers.PasswordChangeHandler{
			Chain:         providerChain,
			ClientManager: clientauth.New(config.Conf.Services.PasswordChange.Clients, true),
		}
		server.AddHandler(proto.PasswordChangeMeta, hdl)
	} else {
		config.Log.Info("passwordChange service disabled")
	}

	err = server.Start(context.Background())
	if err != nil {
		server.GetLog().Error(err, "Error on Start()")
		os.Exit(5)
	}
}
