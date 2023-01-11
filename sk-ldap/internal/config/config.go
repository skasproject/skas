package config

import (
	"github.com/go-logr/logr"
	"skas/sk-common/pkg/httpserver"
	"skas/sk-common/pkg/misc"
	"skas/sk-ldap/internal/ldapprovider"
)

// Exported vars

var (
	Conf       Config
	Log        logr.Logger
	ConfigFile string
)

type Config struct {
	Log    misc.LogConfig          `yaml:"log"`
	Server httpserver.ServerConfig `yaml:"server"`
	Ldap   ldapprovider.Config     `yaml:"ldap"`
}

// NB: These values are strongly inspired from dex configuration (https://github.com/dexidp/dex)
