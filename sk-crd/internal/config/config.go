package config

import (
	"github.com/go-logr/logr"
	"skas/sk-common/pkg/httpserver"
	"skas/sk-common/pkg/misc"
	"skas/sk-common/proto/v1/proto"
)

var (
	Conf Config
	Log  logr.Logger
)

type Config struct {
	Log        misc.LogConfig          `yaml:"log"`
	Server     httpserver.ServerConfig `yaml:"server"`
	Clients    []proto.ClientAuth      `yaml:"clients"`
	Namespace  string                  `yaml:"namespace"` // User database namespace
	MetricAddr string                  `yaml:"metricAddr"`
	ProbeAddr  string                  `yaml:"probeAddr"`
}
