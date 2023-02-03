package main

import (
	"context"
	"fmt"
	"os"
	"skas/sk-auth/internal/config"
	"skas/sk-auth/k8sapi/session/v1alpha1"
	"skas/sk-common/pkg/httpserver"
)

func main() {
	if err := config.Setup(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to load configuration: %v\n", err)
		os.Exit(2)
	}

	config.Log.Info("sk-auth start", "version", config.Version, "logLevel", config.Conf.Log.Level)

	s := &httpserver.Server{
		Name:   "static",
		Log:    config.Log.WithName("authServer"),
		Config: &config.Conf.Server,
	}
	s.Groom()

	xx := &v1alpha1.Token{}
	fmt.Println(xx.Spec.Client)

	err := s.Start(context.Background())
	if err != nil {
		s.Log.Error(err, "Error on Start()")
		os.Exit(5)
	}

}
