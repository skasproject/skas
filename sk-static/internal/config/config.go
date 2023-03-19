package config

import (
	"github.com/go-logr/logr"
	cconfig "skas/sk-common/pkg/config"
	"skas/sk-common/pkg/misc"
)

var (
	Conf              Config
	Log               logr.Logger
	UserByLogin       map[string]StaticUser
	GroupsByUser      map[string][]string
	GroupBindingCount int
)

type Config struct {
	Log     misc.LogConfig          `yaml:"log"`
	Server  cconfig.SkServerConfig  `yaml:"server"`
	Clients []cconfig.ServiceClient `yaml:"clients"`
}

type StaticUser struct {
	Login        string   `yaml:"login"`
	Uid          *int     `yaml:"uid,omitempty"`
	CommonNames  []string `yaml:"commonNames"`
	Emails       []string `yaml:"emails"`
	PasswordHash string   `yaml:"passwordHash"`
	Disabled     *bool    `yaml:"disabled,omitempty"`
}

type StaticGroupBinding struct {
	User  string `yaml:"user"`
	Group string `yaml:"group"`
}

// This is the format of the users file

type StaticUsers struct {
	Users         []StaticUser         `yaml:"users"`
	GroupBindings []StaticGroupBinding `yaml:"groupBindings"`
}
