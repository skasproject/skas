package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/skserver"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-ldap/internal/config"
	"skas/sk-ldap/internal/serverprovider"
)

func main() {
	if err := config.Setup(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to load configuration: %v\n", err)
		os.Exit(2)
	}
	config.Log.Info("sk-ldap start", "ldapServer", config.Conf.Ldap.Host, "version", config.Version, "logLevel", config.Conf.Log.Level)

	server := skserver.New("ldapServer", &config.Conf.Server, config.Log.WithName("ldapServer"))

	provider, err := serverprovider.New(&config.Conf.Ldap, config.Log, filepath.Dir(config.File))
	if err != nil {
		config.Log.Error(err, "ldap config")
		os.Exit(3)
	}
	hdl := &commonHandlers.UserIdentityHandler{
		Provider:      provider,
		ClientManager: clientauth.New(config.Conf.Clients, true),
	}
	server.AddHandler(proto.UserIdentityMeta, hdl)

	err = server.Start(context.Background())
	if err != nil {
		server.GetLog().Error(err, "Error on Start()")
		os.Exit(5)
	}
}
