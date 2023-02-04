package main

import (
	"context"
	"fmt"
	"os"
	"skas/sk-auth/internal/config"
	"skas/sk-auth/internal/handlers"
	"skas/sk-auth/internal/tokenstore"
	"skas/sk-auth/internal/tokenstore/memory"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/httpserver"
	basehandlers "skas/sk-common/pkg/httpserver/handlers"
	"skas/sk-common/pkg/skhttp"
	"skas/sk-common/proto/v1/proto"
)

func main() {
	if err := config.Setup(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to load configuration: %v\n", err)
		os.Exit(2)
	}

	config.Log.Info("sk-auth start", "version", config.Version, "logLevel", config.Conf.Log.Level, "tokenstore", config.Conf.TokenConfig.StorageType)

	s := &httpserver.Server{
		Name:   "static",
		Log:    config.Log.WithName("authServer"),
		Config: &config.Conf.Server,
	}
	s.Groom()

	var tokenStore tokenstore.TokenStore
	if config.Conf.TokenConfig.StorageType == "memory" {
		tokenStore = memory.New(config.Conf.TokenConfig, config.Log.WithName("tokenstore"))
	} else {
		panic("Crd tokenstore not yet implemented")
	}

	loginClient, err := skhttp.New(&config.Conf.LoginProvider, "", "")
	if err != nil {
		config.Log.Error(err, "Error on client login creation")
	}

	if config.Conf.Services.Token.Enabled {
		s.Router.Handle(proto.TokenRequestUrlPath, &handlers.TokenRequestHandler{
			BaseHandler: basehandlers.BaseHandler{
				Logger: s.Log,
			},
			ClientManager: clientauth.New(config.Conf.Services.Token.Clients),
			TokenStore:    tokenStore,
			LoginClient:   loginClient,
		}).Methods("GET")
	}

	if config.Conf.TokenConfig.StorageType == "memory" {
		err := s.Start(context.Background())
		if err != nil {
			s.Log.Error(err, "Error on Start()")
			os.Exit(5)
		}
	} else {
		panic("Crd tokenstore not yet implemented")
	}

}
