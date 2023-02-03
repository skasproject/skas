package config

import (
	"github.com/go-logr/logr"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/httpserver"
	"skas/sk-common/pkg/misc"
	"skas/sk-ldap/internal/serverprovider"
)

// Exported vars

var (
	Conf       Config
	Log        logr.Logger
	ConfigFile string
)

type Config struct {
	Log     misc.LogConfig          `yaml:"log"`
	Server  httpserver.ServerConfig `yaml:"server"`
	Clients []clientauth.Config     `yaml:"clients"`
	Ldap    serverprovider.Config   `yaml:"ldap"`
}
