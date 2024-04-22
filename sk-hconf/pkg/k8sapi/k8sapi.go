package k8sapi

import (
	"fmt"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	Scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(Scheme))
	utilruntime.Must(sourcev1.AddToScheme(Scheme))
}

//
//func GetRestConfig() (*rest.Config, misc.LoggableError) {
//	config, err := rest.InClusterConfig()
//	if err != nil {
//		// fallback to kubeconfig
//		home, err := os.UserHomeDir()
//		if err != nil {
//			return nil, misc.NewLoggableError(err, "unable to locate home directory")
//		}
//		kubeconfig := filepath.Join(home, ".kube", "config")
//		if envVar := os.Getenv("KUBECONFIG"); len(envVar) > 0 {
//			kubeconfig = envVar
//		}
//		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
//		if err != nil {
//			return nil, misc.NewLoggableError(err, "unable to build kubernetes config")
//		}
//	}
//	return config, nil
//}

func BuildRestConfig(kubeconfig string) (*rest.Config, error) {
	var restConfig *rest.Config = nil
	var err error
	if kubeconfig == "" {
		if envvar := os.Getenv("KUBECONFIG"); len(envvar) > 0 {
			kubeconfig = envvar
		}
	}
	if kubeconfig == "" {
		restConfig, err = rest.InClusterConfig()
	}
	if restConfig == nil {
		home := homedir.HomeDir()
		if kubeconfig == "" && home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	return restConfig, nil
}

func GetKubeClientFromConfig(config *rest.Config) (client.Client, error) {
	kubeClient, err := client.New(config, client.Options{
		Scheme: Scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to build kubernetes client: %w", err)
	}
	return kubeClient, nil
}

func GetKubeClient(kubeconfig string) (client.Client, error) {
	config, err := BuildRestConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	return GetKubeClientFromConfig(config)
}

func GetClientSet(kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := BuildRestConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}
