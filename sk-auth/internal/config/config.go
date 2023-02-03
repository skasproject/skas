package config

import (
	"github.com/go-logr/logr"
	"skas/sk-common/pkg/client"
	"skas/sk-common/pkg/httpclient"
	"skas/sk-common/pkg/httpserver"
	"skas/sk-common/pkg/misc"
	"time"
)

var (
	Conf Config
	Log  logr.Logger
)

type TokenConfig struct {
	InactivityTimeout *time.Duration `yaml:"inactivityTimeout"` // After this period without token validation, the session expire
	SessionMaxTTL     *time.Duration `yaml:"sessionMaxTTL"`     // After this period, the session expire, in all case.
	ClientTokenTTL    *time.Duration `yaml:"clientTokenTTL"`    // This is intended for the client CLI, for token caching
	StorageType       string         `yaml:"storageType"`       // 'memory' or 'crd'
	Namespace         string         `yaml:"namespace"`         // When tokenStorage==crd, the namespace to store Tokens.
	LastHitStep       int            `yaml:"lastHitStep"`       // When tokenStorage==crd, the max difference between reality and what is stored in API Server. In per mille of InactivityTimeout. Aim is to avoid API server overloading
}

type Config struct {
	Log           misc.LogConfig          `yaml:"log"`
	Server        httpserver.ServerConfig `yaml:"server"`
	Clients       []client.Config         `yaml:"clients"`
	TokenConfig   TokenConfig             `yaml:"tokenConfig"`
	LoginProvider struct {
		HttpClientConfig httpclient.Config `yaml:"httpClient"`
		Client           client.Config     `yaml:"client"`
	}
}
