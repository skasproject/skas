package main

import (
	"context"
	"fmt"
	"os"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/httpserver"
	"skas/sk-common/pkg/httpserver/handlers"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-static/internal/config"
	"skas/sk-static/internal/staticstatusprovider"
)

func main() {
	if err := config.Setup(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to load configuration: %v\n", err)
		os.Exit(2)
	}

	config.Log.Info("sk-static start", "version", config.Version, "nbUsers", len(config.UserByLogin), "nbrGroupBindings", config.GroupBindingCount, "logLevel", config.Conf.Log.Level)

	//config.Config.Log.V(0).Info("Log V0")
	//config.Config.Log.V(1).Info("Log V1")
	//config.Config.Log.V(2).Info("Log V2")
	//config.Config.Log.Error(errors.New("there is a problem"), "Test ERROR")
	//fmt.Printf("Users:\n%+v\n", config.Config.UserByLogin)

	s := &httpserver.Server{
		Name:   "static",
		Log:    config.Log.WithName("staticServer"),
		Config: &config.Conf.Server,
	}
	s.Groom()
	s.Router.Handle(proto.UserStatusMeta.UrlPath, &handlers.UserStatusHandler{
		BaseHandler: handlers.BaseHandler{
			Logger: s.Log,
		},
		Provider:      staticstatusprovider.New(config.Log.WithName("staticProvider")),
		ClientManager: clientauth.New(config.Conf.Clients),
	}).Methods(proto.UserStatusMeta.Method)
	err := s.Start(context.Background())
	if err != nil {
		s.Log.Error(err, "Error on Start()")
		os.Exit(5)
	}
}
