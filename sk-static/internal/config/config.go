package config

import (
	"github.com/go-logr/logr"
	cconfig "skas/sk-common/pkg/config"
	"skas/sk-common/pkg/misc"
)

//var UserByLogin       map[string]StaticUser
//var GroupsByUser      map[string][]string

var (
	Conf              Config
	Log               logr.Logger
	GroupBindingCount int
)

type StaticServerConfig struct {
	cconfig.SkServerConfig `yaml:",inline"`
	Services               struct {
		Identity cconfig.ServiceConfig `yaml:"identity"`
	} `yaml:"services"`
}

type Config struct {
	Log       misc.LogConfig       `yaml:"log"`
	Servers   []StaticServerConfig `yaml:"servers"`
	UsersFile string               `yaml:"usersFile"`
}
