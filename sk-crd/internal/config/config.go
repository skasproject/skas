package config

import (
	"github.com/go-logr/logr"
	cconfig "skas/sk-common/pkg/config"
	"skas/sk-common/pkg/misc"
)

var (
	Conf Config
	Log  logr.Logger
)

type CrdServerConfig struct {
	cconfig.SkServerConfig `yaml:",inline"`
	Services               struct {
		Identity       cconfig.ServiceConfig `yaml:"identity"`
		PasswordChange cconfig.ServiceConfig `yaml:"passwordChange"`
	} `yaml:"services"`
}

type Config struct {
	Log        misc.LogConfig    `yaml:"log"`
	Servers    []CrdServerConfig `yaml:"servers"`
	Namespace  string            `yaml:"namespace"` // User database namespace
	MetricAddr string            `yaml:"metricAddr"`
	ProbeAddr  string            `yaml:"probeAddr"`
}
