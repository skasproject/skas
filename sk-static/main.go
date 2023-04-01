package main

import (
	"context"
	"fmt"
	"os"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/skserver"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-static/internal/config"
	"skas/sk-static/internal/identitygetter"
)

func main() {
	if err := config.Setup(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to load configuration: %v\n", err)
		os.Exit(2)
	}

	config.Log.Info("sk-static start", "version", config.Version, "nbUsers", len(config.UserByLogin), "nbrGroupBindings", config.GroupBindingCount, "logLevel", config.Conf.Log.Level)

	server := skserver.New("staticServer", &config.Conf.Server, config.Log.WithName("staticServer"))

	hdl := &commonHandlers.IdentityHandler{
		IdentityGetter: identitygetter.New(config.Log.WithName("staticProvider")),
		ClientManager:  clientauth.New(config.Conf.Clients, true),
	}
	server.AddHandler(proto.IdentityMeta, hdl)
	err := server.Start(context.Background())
	if err != nil {
		server.GetLog().Error(err, "Error on Start()")
		os.Exit(5)
	}
}
