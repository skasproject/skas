package config

import (
	"github.com/go-logr/logr"
	"skas/sk-common/clientmanager"
	"skas/sk-common/pkg/httpserver"
	"skas/sk-common/pkg/misc"
)

var (
	Conf Config
	Log  logr.Logger
)

type Config struct {
	Log        misc.LogConfig               `yaml:"log"`
	Server     httpserver.ServerConfig      `yaml:"server"`
	Clients    []clientmanager.ClientConfig `yaml:"clients"`
	Namespace  string                       `yaml:"namespace"` // User database namepace
	MetricAddr string                       `yaml:"metricAddr"`
	ProbeAddr  string                       `yaml:"probeAddr"`
}
