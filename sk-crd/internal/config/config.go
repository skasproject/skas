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

type Config struct {
	Log        misc.LogConfig          `yaml:"log"`
	Server     cconfig.SkServerConfig  `yaml:"server"`
	Clients    []cconfig.ServiceClient `yaml:"clients"`
	Namespace  string                  `yaml:"namespace"` // User database namespace
	MetricAddr string                  `yaml:"metricAddr"`
	ProbeAddr  string                  `yaml:"probeAddr"`
}
