package main

import (
	"fmt"
	"github.com/pior/runnable"
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
	"skas/sk-common/pkg/skclient"
	"skas/sk-common/pkg/skserver"
	"skas/sk-common/proto/v1/proto"
	"time"
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

	// -----------------------------------------------------------------First step of setup

	if config.Conf.TokenConfig.StorageType == "memory" {
		tokenStore = memory.New(config.Conf.TokenConfig, config.Log.WithName("tokenstore"))
	} else {
		ctrl.SetLogger(config.Log.WithName("controller-runtime"))
		var err error
		mgr, err = ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
			Scheme:                 scheme,
			MetricsBindAddress:     config.Conf.MetricAddr,
			HealthProbeBindAddress: config.Conf.ProbeAddr,
			LeaderElection:         false,
			Logger:                 config.Log.WithName("manager"),
			Namespace:              config.Conf.TokenConfig.Namespace,
		})
		time.Sleep(time.Second)
		if err != nil {
			config.Log.Error(err, "unable to initialize manager")
			os.Exit(2)
		}

		if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
			config.Log.Error(err, "unable to set up health check")
			os.Exit(3)
		}
		if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
			config.Log.Error(err, "unable to set up ready check")
			os.Exit(4)
		}
		tokenStore = crd.New(config.Conf.TokenConfig, mgr.GetClient(), config.Log.WithName("tokenstore"))
	}

	// --------------------------------------------------------------- http server setup

	server := skserver.New("auth", &config.Conf.Server, config.Log.WithName("authServer"))

	provider, err := skclient.New(&config.Conf.Provider, "", "")
	if err != nil {
		config.Log.Error(err, "Error on client login creation")
	}
	// ---------------------------------------------------- Token service
	if !config.Conf.Services.Token.Disabled {
		hdlTc := &handlers.TokenCreateHandler{
			ClientManager: clientauth.New(config.Conf.Services.Token.Clients, false),
			TokenStore:    tokenStore,
			Provider:      provider,
		}
		server.AddHandler(proto.TokenCreateMeta, hdlTc)
		hdlTr := &handlers.TokenRenewHandler{
			ClientManager: clientauth.New(config.Conf.Services.Token.Clients, false),
			TokenStore:    tokenStore,
		}
		server.AddHandler(proto.TokenRenewMeta, hdlTr)
	} else {
		config.Log.Info("token service disabled")
	}
	// ---------------------------------------------------- K8sAuth service
	if !config.Conf.Services.K8sAuth.Disabled {
		hdl := &handlers.TokenReviewHandler{
			TokenStore: tokenStore,
		}
		server.AddHandler(proto.TokenReviewMeta, hdl)
	} else {
		config.Log.Info("tokenReview service disabled")
	}
	// ---------------------------------------------------- Describe service
	if !config.Conf.Services.Describe.Disabled {
		hdl := &handlers.UserDescribeHandler{
			ClientManager: clientauth.New(config.Conf.Services.Describe.Clients, false),
			TokenStore:    tokenStore,
			Provider:      provider,
		}
		server.AddHandler(proto.UserDescribeMeta, hdl)
	} else {
		config.Log.Info("userDescribe service disabled")
	}
	// ---------------------------------------------------- PasswordChange service
	if !config.Conf.Services.PasswordChange.Disabled {
		hdl := &handlers.PasswordChangeHandler{
			ClientManager: clientauth.New(config.Conf.Services.PasswordChange.Clients, false),
			Provider:      provider,
		}
		server.AddHandler(proto.PasswordChangeMeta, hdl)
	} else {
		config.Log.Info("passwordChange service disabled")
	}
	// ---------------------------------------------------- Kubeconfig service
	if !config.Conf.Services.Kubeconfig.Disabled {
		hdl := &handlers.KubeconfigHandler{
			ClientManager: clientauth.New(config.Conf.Services.Kubeconfig.Clients, false),
		}
		server.AddHandler(proto.KubeconfigMeta, hdl)
	} else {
		config.Log.Info("kubeconfig service disabled")
	}

	// ---------------------------------------------------------- End init and launch

	if config.Conf.TokenConfig.StorageType == "memory" {
		runnableMgr := runnable.NewManager()
		runnableMgr.Add(server)
		runnableMgr.Add(&tokenstore.Cleaner{
			Period:     60 * time.Second,
			TokenStore: tokenStore,
		})
		runnable.Run(runnableMgr.Build())
		//err := s.Start(context.Background())
		//if err != nil {
		//	s.Log.Error(err, "Error on Start()")
		//	os.Exit(5)
		//}
	} else {
		err = mgr.Add(server)
		if err != nil {
			config.Log.Error(err, "problem adding http server to the manager")
			os.Exit(1)
		}
		err := mgr.Add(&tokenstore.Cleaner{
			Period:     60 * time.Second,
			TokenStore: tokenStore,
		})
		if err != nil {
			config.Log.Error(err, "problem adding cleaner to the manager")
			os.Exit(1)
		}
		config.Log.Info("starting manager")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			config.Log.Error(err, "problem running manager")
			os.Exit(1)
		}
	}
}
