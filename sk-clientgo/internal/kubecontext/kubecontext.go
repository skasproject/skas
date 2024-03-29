package kubecontext

import (
	"k8s.io/client-go/tools/clientcmd"
	"skas/sk-clientgo/internal/global"
	"sync"
)

// From https://pkg.go.dev/k8s.io/client-go@v0.26.1/tools/clientcmd

var once sync.Once
var kubeContext string
var kubeconfigFile string

func GetKubeContext() ( /*kubeconfigFile*/ string /*kubecontext*/, string) {
	once.Do(func() {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		loadingRules.ExplicitPath = global.KubeconfigPath // From the command line. Must take precedence
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		kubeconfigFile = kubeConfig.ConfigAccess().GetDefaultFilename()
		rawConfig, err := kubeConfig.RawConfig()
		if err != nil {
			panic(err)
		}
		kubeContext = rawConfig.CurrentContext
		if kubeContext == "" {
			kubeContext = "default"
		}
		global.Log.V(1).Info("GetKubeContext()", "kubeContext", kubeContext, "kubeconfigFile", kubeconfigFile)
	})
	return kubeconfigFile, kubeContext
}
