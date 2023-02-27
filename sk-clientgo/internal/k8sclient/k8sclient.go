package k8sclient

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sClient struct {
	Client    client.Client
	Namespace string
}

// kubeconfigPath and namespace parameters should come from command line flags

func New(schemeBuilder runtime.SchemeBuilder, kubeconfigPath string, namespace string) (*K8sClient, error) {
	k8sClient := &K8sClient{
		Namespace: namespace,
	}
	if k8sClient.Namespace == "" {
		k8sClient.Namespace = os.Getenv("SKAS_NAMESPACE")
	}
	if k8sClient.Namespace == "" {
		k8sClient.Namespace = "skas-system"
	}
	if kubeconfigPath == "" {
		kubeconfigPath = os.Getenv("KUBECONFIG")
	}
	if kubeconfigPath == "" {
		kubeconfigPath = filepath.Join("~", ".kube", "config")
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
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
	return k8sClient, nil
}
