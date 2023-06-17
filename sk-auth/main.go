package main

import (
	"context"
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
	"skas/sk-auth/internal/identitygetter"
	"skas/sk-auth/internal/tokenstore"
	"skas/sk-auth/internal/tokenstore/crd"
	"skas/sk-auth/internal/tokenstore/memory"
	"skas/sk-auth/k8sapis/session/v1alpha1"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/skclient"
	"skas/sk-common/pkg/skserver"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/pkg/skserver/protector"
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

	config.Log.Info("sk-auth start", "version", config.Version, "build", config.BuildTs, "logLevel", config.Conf.Log.Level, "tokenstore", config.Conf.Token.StorageType)

	var tokenStore tokenstore.TokenStore
	var mgr manager.Manager
	var runnableMgr runnable.AppManager
	// -----------------------------------------------------------------First step of setup

	if config.Conf.Token.StorageType == "memory" {
		tokenStore = memory.New(config.Conf.Token, config.Log.WithName("tokenstore"))
		runnableMgr = runnable.NewManager()
	} else {
		ctrl.SetLogger(config.Log.WithName("controller-runtime"))
		var err error
		mgr, err = ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
			Scheme:                 scheme,
			MetricsBindAddress:     config.Conf.MetricAddr,
			HealthProbeBindAddress: config.Conf.ProbeAddr,
			LeaderElection:         false,
			Logger:                 config.Log.WithName("manager"),
			Namespace:              config.Conf.Token.Namespace,
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
		tokenStore = crd.New(config.Conf.Token, mgr.GetClient(), config.Log.WithName("tokenstore"))
	}

	// --------------------------------------------------------------- http server setup

	provider, err := skclient.New(&config.Conf.Provider, "", "")
	if err != nil {
		config.Log.Error(err, "Error on client login creation")
	}
	identityGetter := identitygetter.New(provider)

	for idx, serverConfig := range config.Conf.Servers {
		server := skserver.New(fmt.Sprintf("server[%d]", idx), &config.Conf.Servers[idx].SkServerConfig, config.Log.WithName(fmt.Sprintf("authServer[%d]", idx)))

		// ---------------------------------------------------- Token service
		if !serverConfig.Services.Token.Disabled {
			prtct := protector.New(serverConfig.Services.Token.Protected, context.Background(), config.Log.WithName("sk-auth.token.protector"))
			hdlTc := &handlers.TokenCreateHandler{
				ClientManager:  clientauth.New(serverConfig.Services.Token.Clients, false),
				TokenStore:     tokenStore,
				IdentityGetter: identityGetter,
				Protector:      prtct,
			}
			server.AddHandler(proto.TokenCreateMeta, hdlTc)
			hdlTr := &handlers.TokenRenewHandler{
				ClientManager: clientauth.New(serverConfig.Services.Token.Clients, false),
				TokenStore:    tokenStore,
				Protector:     prtct,
			}
			server.AddHandler(proto.TokenRenewMeta, hdlTr)
		} else {
			config.Log.Info("'token' service disabled")
		}
		// ---------------------------------------------------- K8sAuth service
		if !serverConfig.Services.K8sAuth.Disabled {
			hdl := &handlers.TokenReviewHandler{
				TokenStore: tokenStore,
				Protector:  protector.New(serverConfig.Services.K8sAuth.Protected, context.Background(), config.Log.WithName("sk-auth.k8sAuth.protector")),
			}
			server.AddHandler(proto.TokenReviewMeta, hdl)
		} else {
			config.Log.Info("'tokenReview' service disabled")
		}
		// ---------------------------------------------------- PasswordChange service
		if !serverConfig.Services.PasswordChange.Disabled {
			hdl := &handlers.PasswordChangeHandler{
				ClientManager: clientauth.New(serverConfig.Services.PasswordChange.Clients, false),
				Provider:      provider,
				Protector:     protector.New(serverConfig.Services.PasswordChange.Protected, context.Background(), config.Log.WithName("sk-auth.passwordChange.protector")),
			}
			server.AddHandler(proto.PasswordChangeMeta, hdl)
		} else {
			config.Log.Info("'passwordChange' service disabled")
		}
		// ---------------------------------------------------- Kubeconfig service
		if !serverConfig.Services.Kubeconfig.Disabled {
			hdl := &handlers.KubeconfigHandler{
				ClientManager: clientauth.New(serverConfig.Services.Kubeconfig.Clients, false),
			}
			server.AddHandler(proto.KubeconfigMeta, hdl)
		} else {
			config.Log.Info("'kubeconfig' service disabled")
		}
		// ---------------------------------------------------- Login service
		if !serverConfig.Services.Login.Disabled {
			hdl := &handlers.LoginHandler{
				ClientManager:  clientauth.New(serverConfig.Services.Kubeconfig.Clients, false),
				IdentityGetter: identityGetter,
				Protector:      protector.New(serverConfig.Services.Login.Protected, context.Background(), config.Log.WithName("sk-auth.login.protector")),
			}
			server.AddHandler(proto.LoginMeta, hdl)
		} else {
			config.Log.Info("'login' service disabled")
		}
		// ---------------------------------------------------- Identity service
		if !serverConfig.Services.Identity.Disabled {
			prt := protector.New(serverConfig.Services.Identity.Protected, context.Background(), config.Log.WithName("sk-auth.identity.protector"))
			identityRequestValidator := &handlers.AdminHttpRequestValidator{
				TokenStore:     tokenStore,
				IdentityGetter: identityGetter,
				Protector:      prt,
			}

			hdl := &commonHandlers.IdentityHandler{
				IdentityGetter:       identityGetter,
				ClientManager:        clientauth.New(serverConfig.Services.Identity.Clients, false),
				HttpRequestValidator: identityRequestValidator,
				Protector:            prt,
			}
			server.AddHandler(proto.IdentityMeta, hdl)
		} else {
			config.Log.Info("'identity' service disabled")
		}
		if config.Conf.Token.StorageType == "memory" {
			runnableMgr.Add(server)
		} else {
			err = mgr.Add(server)
			if err != nil {
				config.Log.Error(err, "problem adding http server to the manager")
				os.Exit(1)
			}
		}
	}
	// ---------------------------------------------------------- End init and launch

	if config.Conf.Token.StorageType == "memory" {
		runnableMgr.Add(&tokenstore.Cleaner{
			Period:     60 * time.Second,
			TokenStore: tokenStore,
		})
		runnable.Run(runnableMgr.Build())
	} else {
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
