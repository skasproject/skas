package config

import (
	"github.com/go-logr/logr"
	"skas/sk-common/pkg/httpserver"
	"skas/sk-common/pkg/misc"
)

var (
	Conf Config
	Log  logr.Logger
)

type Config struct {
	Log       misc.LogConfig          `yaml:"log"`
	Server    httpserver.ServerConfig `yaml:"server"`
	Namespace string                  `yaml:"namespace"` // User database namepace
}
