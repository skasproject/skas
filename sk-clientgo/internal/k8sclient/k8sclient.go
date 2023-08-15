package k8sclient

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
	"skas/sk-clientgo/internal/global"
)

type K8sClient struct {
	Client    client.Client
	Namespace string
}

// kubeconfigPath parameter should come from command line flags
// namespace parameter come from command line

func New(schemeBuilder *scheme.Builder, kubeconfigPath string, namespace string) *K8sClient {
	k8sClient := &K8sClient{
		Namespace: "",
	}
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfigPath // From the command line. Must take precedence
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		fmt.Printf("The kubeconfig cannot be loaded: %v\n", err)
		os.Exit(1)
	}
	crScheme := runtime.NewScheme()
	err = schemeBuilder.AddToScheme(crScheme)
	if err != nil {
		panic(err)
	}
	k8sClient.Client, err = runtimeclient.New(config, client.Options{
		Scheme: crScheme,
	})
	if err != nil {
		panic(err)
	}
	kubeconfigNamespace, overridden, err := kubeConfig.Namespace()
	if err != nil {
		panic(err)
	}
	global.Log.V(1).Info(fmt.Sprintf("Namespace overridden:%t", overridden))

	// Namespace precedence is commandLine > envVar > kubeconfigFile > "skas-system"
	if namespace != "" {
		k8sClient.Namespace = namespace
		global.Log.V(1).Info("Namespace from command line", "namespace", k8sClient.Namespace)
	} else {
		ns := os.Getenv("SKAS_NAMESPACE")
		if ns != "" {
			k8sClient.Namespace = ns
			global.Log.V(1).Info("Namespace from SKAS_NAMESPACE env var", "namespace", k8sClient.Namespace)
		} else {
			if kubeconfigNamespace != "" && kubeconfigNamespace != "default" {
				k8sClient.Namespace = kubeconfigNamespace
				global.Log.V(1).Info("Namespace from kubeconfig file", "namespace", k8sClient.Namespace)
			} else {
				k8sClient.Namespace = "skas-system"
				global.Log.V(1).Info("Namespace undefined. Set 'skas-system'", "namespace", k8sClient.Namespace)
			}
		}
	}
	return k8sClient
}
