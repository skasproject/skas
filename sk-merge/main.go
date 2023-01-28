package main

import (
	"context"
	"fmt"
	"os"
	"skas/sk-common/pkg/httpserver"
	"skas/sk-common/pkg/httpserver/handlers"
	"skas/sk-common/proto"
	"skas/sk-merge/internal/config"
	"skas/sk-merge/internal/serverproviders"
)

func main() {
	if err := config.Setup(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to load configuration: %v\n", err)
		os.Exit(2)
	}
	config.Log.Info("sk-merge start", "version", config.Version, "logLevel", config.Conf.Log.Level)

	s := &httpserver.Server{
		Name:   "merge",
		Log:    config.Log.WithName(fmt.Sprintf("%s", "mergeServer")),
		Config: &config.Conf.Server,
	}
	s.Groom()
	provider, err := serverproviders.New(config.Log)
	if err != nil {
		config.Log.Error(err, "ldap config")
		os.Exit(3)
	}
	s.Router.Handle(proto.UserStatusUrlPath, &handlers.UserStatusHandler{
		BaseHandler: handlers.BaseHandler{
			Logger: s.Log,
		},
		Provider: provider,
	}).Methods("GET")
	err = s.Start(context.Background())
	if err != nil {
		s.Log.Error(err, "Error on Start()")
		os.Exit(5)
	}

}
