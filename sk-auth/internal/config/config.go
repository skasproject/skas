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

type Token struct {
	InactivityTimeout *time.Duration `yaml:"inactivityTimeout"` // After this period without token validation, the session expire
	SessionMaxTTL     *time.Duration `yaml:"sessionMaxTTL"`     // After this period, the session expire, in all case.
	ClientTokenTTL    *time.Duration `yaml:"clientTokenTTL"`    // This is intended for the client CLI, for token caching
	StorageType       string         `yaml:"storageType"`       // 'memory' or 'crd'
	Namespace         string         `yaml:"namespace"`         // When tokenStorage==crd, the namespace to store Tokens.
	LastHitStep       int            `yaml:"lastHitStep"`       // When tokenStorage==crd, the max difference between reality and what is stored in API Server. In per mille of InactivityTimeout. Aim is to avoid API server overloading
}

type PasswordStrength struct {
	ForbidCommon bool `yaml:"forbidCommon"`
	MinimumScore int  `yaml:"minimumScore"`
}

type AuthServerConfig struct {
	cconfig.SkServerConfig `yaml:",inline"`
	Services               struct {
		Token            cconfig.ServiceConfig `yaml:"token"`
		Identity         cconfig.ServiceConfig `yaml:"identity"`
		Login            cconfig.ServiceConfig `yaml:"login"`
		K8sAuth          cconfig.ServiceConfig `yaml:"k8sAuth"`
		PasswordChange   cconfig.ServiceConfig `yaml:"passwordChange"`
		Kubeconfig       cconfig.ServiceConfig `yaml:"kubeconfig"`
		PasswordStrength cconfig.ServiceConfig `yaml:"passwordStrength"`
	} `yaml:"services"`
}

type Config struct {
	Log              misc.LogConfig         `yaml:"log"`
	Servers          []AuthServerConfig     `yaml:"servers"`
	Token            Token                  `yaml:"token"`
	Kubeconfig       proto.KubeconfigConfig `yaml:"kubeconfig"`
	Provider         skclient.Config        `yaml:"provider"`
	AdminGroup       string                 `yaml:"adminGroup"`
	MetricAddr       string                 `yaml:"metricAddr"`
	ProbeAddr        string                 `yaml:"probeAddr"`
	PasswordStrength PasswordStrength       `yaml:"passwordStrength"`
}
