package proto

import (
	"fmt"
	"io"
)

var KubeconfigMeta = &RequestMeta{
	Method:  "GET",
	UrlPath: "/v1/kubeconfig",
}

type KubeconfigRequest struct {
	ClientAuth ClientAuth `json:"clientAuth"`
}

// This structure is also used in configuration. So the yaml

type KubeconfigConfig struct {
	Cluster struct {
		ApiServerUrl       string `yaml:"apiServerUrl" json:"apiServerUrl"`
		RootCaData         string `yaml:"rootCaData" json:"rootCaData"`
		InsecureSkipVerify bool   `yaml:"insecureSkipVerify" json:"insecureSkipVerify"`
	} `yaml:"cluster" json:"cluster"`
	User struct {
		AuthServerUrl      string     `yaml:"authServerUrl" json:"authServerUrl"` // Typically, the ingress of the sk-auth servic
		RootCaData         string     `yaml:"rootCaData" json:"rootCaData"`
		RootCaPath         string     `yaml:"rootCaPath,omitempty" json:"rootCaPath,omitempty"` // This can be used in config. It will be read and set in RootCaData, as it is not relevant in client context
		InsecureSkipVerify bool       `yaml:"insecureSkipVerify" json:"insecureSkipVerify"`
		ClientAuth         ClientAuth `yaml:"clientAuth" json:"clientAuth"`
	} `yaml:"user" json:"user"`
	Context struct {
		Name      string `yaml:"name" json:"name"` // Cluster will be named <context>-cluster and user <context>-user
		Namespace string `yaml:"namespace" json:"namespace"`
	}
}

type KubeconfigResponse struct {
	KubeconfigConfig
}

// -----------------------------------------------------

var _ RequestPayload = &KubeconfigRequest{}
var _ ResponsePayload = &KubeconfigResponse{}

func (k *KubeconfigRequest) String() string {
	return fmt.Sprintf("KubeconfigRequest (client.id:%s", k.ClientAuth.Id)
}
func (k *KubeconfigRequest) ToJson() ([]byte, error) {
	return toJson(k)
}
func (k *KubeconfigRequest) FromJson(r io.Reader) error {
	return fromJson(r, k)
}

func (k *KubeconfigResponse) FromJson(r io.Reader) error {
	return fromJson(r, k)
}
