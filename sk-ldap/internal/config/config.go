package config

import (
	"github.com/go-logr/logr"
	cconfig "skas/sk-common/pkg/config"
	"skas/sk-common/pkg/misc"
	"skas/sk-ldap/internal/identitygetter"
)

// Exported vars

var (
	Conf Config
	Log  logr.Logger
	File string
)

type Config struct {
	Log     misc.LogConfig          `yaml:"log"`
	Server  cconfig.SkServerConfig  `yaml:"server"`
	Clients []cconfig.ServiceClient `yaml:"clients"`
	Ldap    identitygetter.Config   `yaml:"ldap"`
}
