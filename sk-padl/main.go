package main

import (
	"fmt"
	"github.com/nmcclain/ldap"
	"os"
	cconfig "skas/sk-common/pkg/config"
	"skas/sk-common/pkg/skclient"
	"skas/sk-padl/internal/config"
	"skas/sk-padl/internal/handlers"
)

func main() {
	if err := config.Setup(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to load configuration: %v\n", err)
		os.Exit(2)
	}
	config.Log.Info("sk-padl start", "version", cconfig.Version, "build", cconfig.BuildTs, "logLevel", config.Conf.Log.Level, "roBindDn", config.Conf.RoBindDn, "usersBaseDn", config.Conf.UsersBaseDn, "groupsBaseDn", config.Conf.GroupsBaseDn)

	provider, err := skclient.New(&config.Conf.Provider, "", "")
	if err != nil {
		config.Log.Error(err, "Error on client login creation")
		os.Exit(3)
	}

	server := ldap.NewServer()
	handler := handlers.New(config.Log, provider)
	server.BindFunc("", handler)
	server.SearchFunc("", handler)
	server.CloseFunc("", handler)

	server.CompareFunc("", handler)
	server.AbandonFunc("", handler)
	server.ExtendedFunc("", handler)
	server.UnbindFunc("", handler)
	server.AddFunc("", handler)
	server.ModifyFunc("", handler)
	server.DeleteFunc("", handler)
	server.ModifyFunc("", handler)

	listenAddr := fmt.Sprintf("%s:%d", config.Conf.Interface, config.Conf.Port)
	if *config.Conf.Ssl {
		config.Log.V(0).Info("LDAPS server listening", "addr", listenAddr)
		if err := server.ListenAndServeTLS(listenAddr, config.CertPath, config.KeyPath); err != nil {
			config.Log.Error(err, "LDAPS Server init failed")
			os.Exit(1)
		}
	} else {
		config.Log.V(0).Info("LDAP server listening", "addr", listenAddr)
		if err := server.ListenAndServe(listenAddr); err != nil {
			config.Log.Error(err, "LDAP Server init failed")
			os.Exit(1)
		}
	}
}
