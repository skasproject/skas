package k8sclient

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

type K8sClient struct {
	Client    client.Client
	Namespace string
}

// kubeconfigPath and namespace parameters should come from command line flags

func New(schemeBuilder *scheme.Builder, kubeconfigPath string, namespace string) *K8sClient {
	k8sClient := &K8sClient{
		Namespace: namespace,
	}
	if k8sClient.Namespace == "" {
		k8sClient.Namespace = os.Getenv("SKAS_NAMESPACE")
	}
	if k8sClient.Namespace == "" {
		k8sClient.Namespace = "skas-system"
	}
	//if kubeconfigPath == "" {
	//	kubeconfigPath = os.Getenv("KUBECONFIG")
	//}
	//if kubeconfigPath == "" {
	//	kubeconfigPath = filepath.Join("~", ".kube", "config")
	//}
	//config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	//if err != nil {
	//	fmt.Printf("The kubeconfig cannot be loaded: %v\n", err)
	//	os.Exit(1)
	//}
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
	return k8sClient
}
