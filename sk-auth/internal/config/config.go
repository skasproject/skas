package config

import (
	"github.com/go-logr/logr"
	cconfig "skas/sk-common/pkg/config"
	"skas/sk-common/pkg/misc"
	"skas/sk-common/pkg/skclient"
	"skas/sk-common/proto/v1/proto"
	"time"
)

var (
	Conf Config
	Log  logr.Logger
)

type TokenConfig struct {
	InactivityTimeout *time.Duration `yaml:"inactivityTimeout"` // After this period without token validation, the session expire
	SessionMaxTTL     *time.Duration `yaml:"sessionMaxTTL"`     // After this period, the session expire, in all case.
	ClientTokenTTL    *time.Duration `yaml:"clientTokenTTL"`    // This is intended for the client CLI, for token caching
	StorageType       string         `yaml:"storageType"`       // 'memory' or 'crd'
	Namespace         string         `yaml:"namespace"`         // When tokenStorage==crd, the namespace to store Tokens.
	LastHitStep       int            `yaml:"lastHitStep"`       // When tokenStorage==crd, the max difference between reality and what is stored in API Server. In per mille of InactivityTimeout. Aim is to avoid API server overloading
}

type Config struct {
	Log         misc.LogConfig         `yaml:"log"`
	Server      cconfig.SkServerConfig `yaml:"server"`
	TokenConfig TokenConfig            `yaml:"tokenConfig"`
	Kubeconfig  proto.KubeconfigConfig `yaml:"kubeconfig"`
	Provider    skclient.Config        `yaml:"provider"`
	AdminGroup  string                 `yaml:"adminGroup"`
	Services    struct {
		Token          cconfig.ServiceConfig `yaml:"token"`
		Describe       cconfig.ServiceConfig `yaml:"describe"`
		K8sAuth        cconfig.ServiceConfig `yaml:"k8sAuth"`
		PasswordChange cconfig.ServiceConfig `yaml:"passwordChange"`
		Kubeconfig     cconfig.ServiceConfig `yaml:"kubeconfig"`
	} `yaml:"services"`
	MetricAddr string `yaml:"metricAddr"`
	ProbeAddr  string `yaml:"probeAddr"`
}
