package main

import (
	"context"
	"fmt"
	"os"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/skserver"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-merge/internal/clientproviderchain"
	"skas/sk-merge/internal/config"
	"skas/sk-merge/internal/handlers"
	"skas/sk-merge/internal/serverproviders"
)

func main() {
	if err := config.Setup(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to load configuration: %v\n", err)
		os.Exit(2)
	}
	config.Log.Info("sk-merge start", "version", config.Version, "logLevel", config.Conf.Log.Level)

	server := skserver.New("merge", &config.Conf.Server, config.Log.WithName(fmt.Sprintf("%s", "mergeServer")))

	providerChain, err := clientproviderchain.New(config.Log.WithName("providerChain"))
	if err != nil {
		config.Log.Error(err, "Error on clientProviderChain creation")
		os.Exit(6)
	}
	if providerChain.GetLength() == 0 {
		config.Log.Error(fmt.Errorf("no client provider defined"), "No client provider defined")
		os.Exit(7)
	}
	// --------------------- UserDescribe handler
	if !config.Conf.Services.UserDescribe.Disabled {
		hdl := &handlers.UserDescribeHandler{
			Chain:         providerChain,
			ClientManager: clientauth.New(config.Conf.Services.UserDescribe.Clients, true),
		}
		server.AddHandler(proto.UserDescribeMeta, hdl)
	} else {
		config.Log.Info("UserDescribe service disabled")
	}
	// --------------------- Login handler
	if !config.Conf.Services.Login.Disabled {
		hdl := &handlers.LoginHandler{
			Chain:         providerChain,
			ClientManager: clientauth.New(config.Conf.Services.Login.Clients, true),
		}
		server.AddHandler(proto.LoginMeta, hdl)
	} else {
		config.Log.Info("Login service disabled")
	}
	// --------------------- UserIdentity handler
	if !config.Conf.Services.UserIdentity.Disabled {
		identityServerProvider, err := serverproviders.NewIdentityServerProvider(providerChain, config.Log)
		if err != nil {
			config.Log.Error(err, "Error on identityServerProvider creation")
			os.Exit(3)
		}
		hdl := &commonHandlers.UserIdentityHandler{
			Provider:      identityServerProvider,
			ClientManager: clientauth.New(config.Conf.Services.UserIdentity.Clients, true),
		}
		server.AddHandler(proto.UserIdentityMeta, hdl)
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
