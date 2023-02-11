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
	config.Log.Info("Login service", "enabled", config.Conf.Services.Login.Enabled)
	config.Log.Info("UserStatus service", "enabled", config.Conf.Services.UserStatus.Enabled)
	config.Log.Info("UserDescribe service", "enabled", config.Conf.Services.UserDescribe.Enabled)

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
	if config.Conf.Services.UserDescribe.Enabled {
		s.Router.Handle(proto.UserDescribeMeta.UrlPath, handlers.UserDescribeHandler{
			BaseHandler: commonHandlers.BaseHandler{
				Logger: s.Log.WithName("userDescribe handler"),
			},
			Chain:         providerChain,
			ClientManager: clientauth.New(config.Conf.Services.UserDescribe.Clients),
		}).Methods(proto.UserDescribeMeta.Method)
	}
	// --------------------- Login handler
	if config.Conf.Services.Login.Enabled {
		s.Router.Handle(proto.LoginMeta.UrlPath, handlers.LoginHandler{
			BaseHandler: commonHandlers.BaseHandler{
				Logger: s.Log.WithName("login handler"),
			},
			Chain:         providerChain,
			ClientManager: clientauth.New(config.Conf.Services.Login.Clients),
		}).Methods(proto.LoginMeta.Method)
	}
	// --------------------- UserStatus handler
	if config.Conf.Services.UserStatus.Enabled {
		statusServerProvider, err := serverproviders.NewStatusServerProvider(providerChain, config.Log)
		if err != nil {
			config.Log.Error(err, "Error on statusServerProvider creation")
			os.Exit(3)
		}
		s.Router.Handle(proto.UserStatusMeta.UrlPath, &commonHandlers.UserStatusHandler{
			BaseHandler: commonHandlers.BaseHandler{
				Logger: s.Log.WithName("userStatus handler"),
			},
			Provider:      statusServerProvider,
			ClientManager: clientauth.New(config.Conf.Services.UserStatus.Clients),
		}).Methods(proto.UserStatusMeta.Method)
	}

	err = s.Start(context.Background())
	if err != nil {
		s.Log.Error(err, "Error on Start()")
		os.Exit(5)
	}
}
