package config

import (
	"github.com/go-logr/logr"
	"skas/sk-common/pkg/client"
	"skas/sk-common/pkg/httpserver"
	"skas/sk-common/pkg/misc"
)

var (
	Conf Config
	Log  logr.Logger
)

type Config struct {
	Log        misc.LogConfig          `yaml:"log"`
	Server     httpserver.ServerConfig `yaml:"server"`
	Clients    []client.Config         `yaml:"clients"`
	Namespace  string                  `yaml:"namespace"` // User database namepace
	MetricAddr string                  `yaml:"metricAddr"`
	ProbeAddr  string                  `yaml:"probeAddr"`
}
