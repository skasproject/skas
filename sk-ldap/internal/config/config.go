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

type LdapServerConfig struct {
	cconfig.SkServerConfig `yaml:",inline"`
	Services               struct {
		Identity cconfig.ServiceConfig `yaml:"identity"`
	} `yaml:"services"`
}

type Config struct {
	Log     misc.LogConfig        `yaml:"log"`
	Servers []LdapServerConfig    `yaml:"servers"`
	Ldap    identitygetter.Config `yaml:"ldap"`
}
