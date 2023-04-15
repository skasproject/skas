package config

import (
	"github.com/go-logr/logr"
	cconfig "skas/sk-common/pkg/config"
	"skas/sk-common/pkg/misc"
)

//var UserByLogin       map[string]StaticUser
//var GroupsByUser      map[string][]string

var (
	Conf Config
	Log  logr.Logger
)

type StaticServerConfig struct {
	cconfig.SkServerConfig `yaml:",inline"`
	Services               struct {
		Identity cconfig.ServiceConfig `yaml:"identity"`
	} `yaml:"services"`
}

type Config struct {
	Log            misc.LogConfig       `yaml:"log"`
	Servers        []StaticServerConfig `yaml:"servers"`
	UsersFile      string               `yaml:"usersFile"`      // Exclusive from UsersConfigMap
	UsersConfigMap string               `yaml:"usersConfigMap"` // Exclusive from UsersFile
	CmLocation     struct {             // Used only in out-of-cluster context
		Namespace  string `yaml:"namespace"`  // If empty lookup current namespace.
		Kubeconfig string `yaml:"kubeconfig"` // If empty, lookup in cluster config
	} `yaml:"cmLocation"`
}
