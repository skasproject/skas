package config

import (
	"github.com/go-logr/logr"
	"skas/sk-common/pkg/httpserver"
	"skas/sk-common/pkg/misc"
)

var (
	Conf        Config
	Log         logr.Logger
	UserByLogin map[string]StaticUser
)

type Config struct {
	Log    misc.LogConfig          `yaml:"log"`
	Server httpserver.ServerConfig `yaml:"server"`
}

type StaticUser struct {
	Login        string   `yaml:"login"`
	Uid          int64    `yaml:"uid"`
	CommonNames  []string `yaml:"commonNames"`
	Emails       []string `yaml:"emails"`
	Groups       []string `yaml:"groups"`
	PasswordHash string   `yaml:"passwordHash"`
	Disabled     *bool    `yaml:"disabled, omitempty"`
}

// This is the format of the users file

type StaticUsers struct {
	Users []StaticUser `yaml:"users"`
}
