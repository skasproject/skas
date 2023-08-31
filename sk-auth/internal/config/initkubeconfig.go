package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"skas/sk-common/proto/v1/proto"
	"strings"
)

// Try to fetch the cluster certificate in a 'well known" location

func fetchClusterCa() string {
	buf, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(buf)
}

func isHttps(url string) bool {
	return strings.HasPrefix(strings.ToLower(url), "https://")
}

// This function will try to fulfill some information by grabbing the current context
// https://stackoverflow.com/questions/38242062/how-to-get-kubernetes-cluster-name-from-k8s-api

func initKubeconfig(kc *proto.KubeconfigConfig) error {
	// Cluster section
	if kc.Cluster.ApiServerUrl == "" {
		return fmt.Errorf("cluster.apiServerUrl is a required parameter")
	}
	if isHttps(kc.Cluster.ApiServerUrl) && !kc.Cluster.InsecureSkipVerify {
		if kc.Cluster.RootCaData == "" {
			kc.Cluster.RootCaData = fetchClusterCa()
			if kc.Cluster.RootCaData == "" {
				return fmt.Errorf("cluster.rootCaData is a required parameter (Or set cluster.insecureSkipVerify)")
			}
		}
	}

	if kc.Context.Name == "" {
		kc.Context.Name = "skas@cluster.local"
	}
	return nil
}
