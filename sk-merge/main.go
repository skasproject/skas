package main

import (
	"context"
	"fmt"
	"os"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/httpserver"
	commonHandlers "skas/sk-common/pkg/httpserver/handlers"
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
	config.Log.Info("Login service", "enabled", !config.Conf.Services.Login.Disabled)
	config.Log.Info("UserIdentity service", "enabled", !config.Conf.Services.UserIdentity.Disabled)
	config.Log.Info("UserDescribe service", "enabled", !config.Conf.Services.UserDescribe.Disabled)
	config.Log.Info("PasswordChange service", "enabled", !config.Conf.Services.PasswordChange.Disabled)

	s := &httpserver.Server{
		Name:   "merge",
		Log:    config.Log.WithName(fmt.Sprintf("%s", "mergeServer")),
		Config: &config.Conf.Server,
	}
	s.Groom()

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
		s.Router.Handle(proto.UserDescribeMeta.UrlPath, handlers.UserDescribeHandler{
			BaseHandler: commonHandlers.BaseHandler{
				Logger: s.Log.WithName("userDescribe handler"),
			},
			Chain:         providerChain,
			ClientManager: clientauth.New(config.Conf.Services.UserDescribe.Clients, true),
		}).Methods(proto.UserDescribeMeta.Method)
	}
	// --------------------- Login handler
	if !config.Conf.Services.Login.Disabled {
		s.Router.Handle(proto.LoginMeta.UrlPath, handlers.LoginHandler{
			BaseHandler: commonHandlers.BaseHandler{
				Logger: s.Log.WithName("login handler"),
			},
			Chain:         providerChain,
			ClientManager: clientauth.New(config.Conf.Services.Login.Clients, true),
		}).Methods(proto.LoginMeta.Method)
	}
	// --------------------- UserIdentity handler
	if !config.Conf.Services.UserIdentity.Disabled {
		identityServerProvider, err := serverproviders.NewIdentityServerProvider(providerChain, config.Log)
		if err != nil {
			config.Log.Error(err, "Error on identityServerProvider creation")
			os.Exit(3)
		}
		s.Router.Handle(proto.UserIdentityMeta.UrlPath, &commonHandlers.UserIdentityHandler{
			BaseHandler: commonHandlers.BaseHandler{
				Logger: s.Log.WithName("userIdentity handler"),
			},
			Provider:      identityServerProvider,
			ClientManager: clientauth.New(config.Conf.Services.UserIdentity.Clients, true),
		}).Methods(proto.UserIdentityMeta.Method)
	}
	// --------------------- PasswordChange handler
	if !config.Conf.Services.PasswordChange.Disabled {
		s.Router.Handle(proto.PasswordChangeMeta.UrlPath, handlers.PasswordChangeHandler{
			BaseHandler: commonHandlers.BaseHandler{
				Logger: s.Log.WithName("passwordChange handler"),
			},
			Chain:         providerChain,
			ClientManager: clientauth.New(config.Conf.Services.PasswordChange.Clients, true),
		}).Methods(proto.PasswordChangeMeta.Method)
	}

	err = s.Start(context.Background())
	if err != nil {
		s.Log.Error(err, "Error on Start()")
		os.Exit(5)
	}
}
