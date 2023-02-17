package kubecontext

import (
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"skas/sk-clientgo/internal/log"
)

// Exposed global variables

var KubeContext string

// From https://pkg.go.dev/k8s.io/client-go@v0.26.1/tools/clientcmd

func loadRawConfig(kubeconfig string) clientcmdapi.Config {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfig // From the command line. Must take precedence
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		panic(err)
	}
	return rawConfig
}

func Initialize(kubeconfig string) {
	if KubeContext == "" {
		rawConfig := loadRawConfig(kubeconfig)
		KubeContext = rawConfig.CurrentContext
		if KubeContext == "" {
			KubeContext = "default"
		}
	}
	log.Log.V(1).Info("kubeContext.Initialize()", "kubeContext", KubeContext)
}
