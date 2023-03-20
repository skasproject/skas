package main

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	kubeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	userdbv1alpha1 "skas/sk-common/k8sapis/userdb/v1alpha1"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/skserver"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-crd/internal/config"
	"skas/sk-crd/internal/crdidentityprovider"
	"skas/sk-crd/internal/handlers"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(userdbv1alpha1.AddToScheme(scheme))
}

func main() {
	if err := config.Setup(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to load configuration: %v\n", err)
		os.Exit(2)
	}

	config.Log.Info("sk-crd start", "userDbNamespace", config.Conf.Namespace, "version", config.Version, "logLevel", config.Conf.Log.Level)

	//config.Config.Log.V(0).Info("Log V0")
	//config.Config.Log.V(1).Info("Log V1")
	//config.Config.Log.V(2).Info("Log V2")
	//config.Config.Log.Error(errors.New("there is a problem"), "Test ERROR")
	//fmt.Printf("Users:\n%+v\n", config.Config.UserByLogin)

	ctrl.SetLogger(config.Log.WithName("controller-runtime"))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     config.Conf.MetricAddr,
		HealthProbeBindAddress: config.Conf.ProbeAddr,
		LeaderElection:         false,
		Logger:                 config.Log.WithName("manager"),
		Namespace:              config.Conf.Namespace,
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

	server := skserver.New("crdServer", &config.Conf.Server, config.Log.WithName("crdServer"))

	hdlUi := &commonHandlers.UserIdentityHandler{
		Provider:      crdidentityprovider.New(mgr.GetClient(), config.Conf.Namespace, config.Log.WithName("crdprovider")),
		ClientManager: clientauth.New(config.Conf.Clients, true),
	}
	server.AddHandler(proto.UserIdentityMeta, hdlUi)

	hdlCp := &handlers.PasswordChangeHandler{
		ClientManager: clientauth.New(config.Conf.Clients, true),
		KubeClient:    mgr.GetClient(),
		Namespace:     config.Conf.Namespace,
	}
	server.AddHandler(proto.PasswordChangeMeta, hdlCp)

	err = mgr.Add(server)
	if err != nil {
		config.Log.Error(err, "problem adding http server to the manager")
		os.Exit(1)
	}

	err = mgr.GetFieldIndexer().IndexField(context.TODO(), &userdbv1alpha1.GroupBinding{}, "userkey", func(rawObj kubeclient.Object) []string {
		ugb := rawObj.(*userdbv1alpha1.GroupBinding)
		return []string{ugb.Spec.User}
	})
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	config.Log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		config.Log.Error(err, "problem running manager")
		os.Exit(1)
	}

}
