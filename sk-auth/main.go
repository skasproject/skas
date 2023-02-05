package main

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"skas/sk-auth/internal/config"
	"skas/sk-auth/internal/handlers"
	"skas/sk-auth/internal/tokenstore"
	"skas/sk-auth/internal/tokenstore/crd"
	"skas/sk-auth/internal/tokenstore/memory"
	"skas/sk-auth/k8sapis/session/v1alpha1"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/httpserver"
	basehandlers "skas/sk-common/pkg/httpserver/handlers"
	"skas/sk-common/pkg/skhttp"
	"skas/sk-common/proto/v1/proto"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
}

func main() {
	if err := config.Setup(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to load configuration: %v\n", err)
		os.Exit(2)
	}

	config.Log.Info("sk-auth start", "version", config.Version, "logLevel", config.Conf.Log.Level, "tokenstore", config.Conf.TokenConfig.StorageType)

	var tokenStore tokenstore.TokenStore
	var mgr manager.Manager

	if config.Conf.TokenConfig.StorageType == "memory" {
		tokenStore = memory.New(config.Conf.TokenConfig, config.Log.WithName("tokenstore"))
	} else {
		var err error
		mgr, err = ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
			Scheme:                 scheme,
			MetricsBindAddress:     config.Conf.MetricAddr,
			HealthProbeBindAddress: config.Conf.ProbeAddr,
			LeaderElection:         false,
			Logger:                 config.Log.WithName("manager"),
			Namespace:              config.Conf.TokenConfig.Namespace,
		})
		if err != nil {
			config.Log.Error(err, "unable to initialize manager")
			os.Exit(1)
		}

		if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
			config.Log.Error(err, "unable to set up health check")
			os.Exit(1)
		}
		if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
			config.Log.Error(err, "unable to set up ready check")
			os.Exit(1)
		}
		tokenStore = crd.New(config.Conf.TokenConfig, mgr.GetClient(), config.Log.WithName("tokenstore"))
	}

	s := &httpserver.Server{
		Name:   "auth",
		Log:    config.Log.WithName("authServer"),
		Config: &config.Conf.Server,
	}
	s.Groom()

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
		err = mgr.Add(s)
		if err != nil {
			config.Log.Error(err, "problem adding http server to the manager")
			os.Exit(1)
		}
		config.Log.Info("starting manager")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			config.Log.Error(err, "problem running manager")
			os.Exit(1)
		}
	}

}
