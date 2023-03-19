package config

import (
	"github.com/go-logr/logr"
	"skas/sk-common/pkg/misc"
	"skas/sk-common/pkg/skserver"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-ldap/internal/serverprovider"
)

// Exported vars

var (
	Conf       Config
	Log        logr.Logger
	ConfigFile string
)

type Config struct {
	Log     misc.LogConfig        `yaml:"log"`
	Server  skserver.ServerConfig `yaml:"server"`
	Clients []proto.ClientAuth    `yaml:"clients"`
	Ldap    serverprovider.Config `yaml:"ldap"`
}
