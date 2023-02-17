package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"skas/sk-common/proto/v1/proto"
)

// Try to fetch the cluster certificate in a 'well known" location

func fetchClusterCa() string {
	buf, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(buf)
}

func fetchAuthCa(path string) string {
	p := path
	if p == "" {
		p = "/tmp/cert/server/ca.crt"
	}
	buf, err := os.ReadFile(p)
	if err != nil {
		if path != "" {
			// We have provided a path, but it is invalid. Display a message
			Log.Error(err, "Error while loading 'auth' server CA file", "path", path)
		}
		return ""
	}
	return base64.StdEncoding.EncodeToString(buf)
}

// This function will try to fulfill some information by grabbing the current context
// https://stackoverflow.com/questions/38242062/how-to-get-kubernetes-cluster-name-from-k8s-api

func initKubeconfig(kc *proto.KubeconfigConfig) error {
	// Cluster section
	if kc.Cluster.ApiServerUrl == "" {
		return fmt.Errorf("cluster.apiServerUrl is a required parameter")
	}
	if !kc.Cluster.InsecureSkipVerify {

		if kc.Cluster.RootCaData == "" {
			kc.Cluster.RootCaData = fetchClusterCa()
			if kc.Cluster.RootCaData == "" {
				return fmt.Errorf("cluster.rootCaData is a required parameter (Or set cluster.insecureSkipTLSVerify)")
			}
		}
	}
	// User section
	if kc.User.AuthServerUrl == "" {
		return fmt.Errorf("user.authServerUrl is a required parameter")
	}
	if !kc.User.InsecureSkipVerify {
		if kc.User.RootCaData == "" {
			kc.User.RootCaData = fetchAuthCa(kc.User.RootCaPath)
			if kc.User.RootCaData == "" {
				return fmt.Errorf("user.rootCaData is a required parameter (Or set user.insecureSkipVerify)")
			}
			kc.User.RootCaPath = "" // We don't want client to see this (As the path is inside pod server
		}
	}

	if kc.ContextName == "" {
		kc.ContextName = "skas"
	}
	return nil
}
