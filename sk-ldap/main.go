package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"skas/sk-common/pkg/httpserver"
	"skas/sk-common/proto"
	"skas/sk-ldap/internal/handlers"
	"skas/sk-ldap/internal/ldapprovider"
)
import "skas/sk-ldap/internal/config"

func main() {
	if err := config.Setup(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to load configuration: %v\n", err)
		os.Exit(2)
	}
	config.Log.Info("sk-ldap start", "ldapServer", config.Conf.Ldap.Host)

	//config.Log.V(0).Info("Log V0")
	//config.Log.V(1).Info("Log V1")
	//config.Log.V(2).Info("Log V2")
	//config.Log.Error(errors.New("there is a problem"), "Test ERROR")

	name := fmt.Sprintf("ldap[%s]", config.Conf.Ldap.Host)
	s := &httpserver.Server{
		Name:   name,
		Log:    config.Log.WithName(fmt.Sprintf("%s", name)),
		Config: &config.Conf.Server,
	}
	s.Groom()
	provider, err := ldapprovider.New(&config.Conf.Ldap, config.Log, filepath.Dir(config.ConfigFile))
	if err != nil {
		config.Log.Error(err, "ldap config")
		os.Exit(3)
	}
	s.Router.Handle(proto.UserStatusUrlPath, &handlers.UserStatusHandler{
		BaseHandler: httpserver.BaseHandler{
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
