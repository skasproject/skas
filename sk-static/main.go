package main

import (
	"context"
	"fmt"
	"os"
	"skas/sk-common/pkg/httpserver"
	"skas/sk-common/proto"
	"skas/sk-static/internal/config"
	"skas/sk-static/internal/handlers"
)

func main() {
	if err := config.Setup(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to load configuration: %v\n", err)
		os.Exit(2)
	}

	config.Config.Log.Info("sk-static start", "nbUsers", len(config.Config.UserByLogin))

	//config.Config.Log.Error(errors.New("there is a problem"), "Test ERROR")
	//config.Config.Log.V(0).Info("Log V0")
	//config.Config.Log.V(1).Info("Log V1")
	//config.Config.Log.V(2).Info("Log V2")
	//config.Config.Log.V(-1).Info("Log V-1")
	//fmt.Printf("Users:\n%+v\n", config.Config.UserByLogin)

	s := &httpserver.Server{
		Name:     "static",
		Log:      config.Config.Log.WithName(fmt.Sprintf("%s http server", "static")),
		BindAddr: config.Config.BindAddr,
		NoSsl:    config.Config.NoSsl,
		CertDir:  config.Config.CertDir,
		CertName: config.Config.CertName,
		KeyName:  config.Config.KeyName,
	}
	s.Groom()
	s.Router.Handle(proto.UserStatusUrlPath, &handlers.UserStatusHandler{
		BaseHandler: httpserver.BaseHandler{
			Logger: s.Log,
		},
	}).Methods("GET")
	err := s.Start(context.Background())
	if err != nil {
		s.Log.Error(err, "Server")
		os.Exit(5)
	}
}
