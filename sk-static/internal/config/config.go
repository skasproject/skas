package config

import (
	"github.com/go-logr/logr"
)

var Config struct {
	BindAddr    string
	UserByLogin map[string]StaticUser
	Log         logr.Logger
	NoSsl       bool
	CertDir     string
	CertName    string
	KeyName     string
}

type StaticUser struct {
	Login        string   `yaml:"login"`
	Uid          int64    `yaml:"uid"`
	CommonNames  []string `yaml:"commonNames"`
	Emails       []string `yaml:"emails"`
	Groups       []string `yaml:"groups"`
	PasswordHash string   `yaml:"passwordHash"`
}

type StaticUsers struct {
	Users []StaticUser `yaml:"users"`
}
