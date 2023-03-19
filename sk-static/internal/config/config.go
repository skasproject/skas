package config

import (
	"github.com/go-logr/logr"
	"skas/sk-common/pkg/misc"
	"skas/sk-common/pkg/skserver"
	"skas/sk-common/proto/v1/proto"
)

var (
	Conf              Config
	Log               logr.Logger
	UserByLogin       map[string]StaticUser
	GroupsByUser      map[string][]string
	GroupBindingCount int
)

type Config struct {
	Log     misc.LogConfig        `yaml:"log"`
	Server  skserver.ServerConfig `yaml:"server"`
	Clients []proto.ClientAuth    `yaml:"clients"`
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
