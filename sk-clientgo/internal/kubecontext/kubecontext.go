package kubecontext

import (
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"skas/sk-clientgo/internal/global"
	"sync"
)

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

var once sync.Once
var kubeContext string

func GetKubeContext() string {
	once.Do(func() {
		rawConfig := loadRawConfig(global.KubeconfigPath)
		kubeContext = rawConfig.CurrentContext
		if kubeContext == "" {
			kubeContext = "default"
		}
		global.Log.V(1).Info("GetKubeContext()", "kubeContext", kubeContext)
	})
	return kubeContext
}
