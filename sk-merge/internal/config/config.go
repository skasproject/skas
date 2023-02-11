package config

import (
	"github.com/go-logr/logr"
	"skas/sk-common/pkg/httpserver"
	"skas/sk-common/pkg/misc"
	"skas/sk-common/pkg/skhttp"
	"skas/sk-common/proto/v1/proto"
)

// Exported vars

var (
	Conf Config
	Log  logr.Logger
)

// NB: All RootCA will be cumulated

type ClientProviderConfig struct {
	Name                string        `yaml:"name"`
	HttpClient          skhttp.Config `yaml:"httpClient"`
	Enabled             *bool         `yaml:"enabled"`             // Allow to disable a provider
	CredentialAuthority *bool         `yaml:"credentialAuthority"` // Is this ldap is authority for password checking
	GroupAuthority      *bool         `yaml:"groupAuthority"`      // Group will be fetched. Default true
	Critical            *bool         `yaml:"critical"`            // If true (default), a failure on this provider will leads 'invalid login'. Even if another provider grants access
	GroupPattern        string        `yaml:"groupPattern"`        // Group pattern. Default "%s"
	UidOffset           int64         `yaml:"uidOffset"`           // Will be added to the returned Uid. Default to 0
}

type ServiceConfig struct {
	Disabled bool               `yaml:"disabled"`
	Clients  []proto.ClientAuth `yaml:"clients"`
}

type Config struct {
	Log       misc.LogConfig          `yaml:"log"`
	Server    httpserver.ServerConfig `yaml:"server"`
	Providers []ClientProviderConfig  `yaml:"providers"`
	// values added to above Providers
	RootCaPath string `yaml:"rootCaPath"` // Path to a trusted root CA file
	RootCaData string `yaml:"rootCaData"` // Base64 encoded PEM data containing root CA
	Services   struct {
		Login        ServiceConfig `yaml:"login"`
		UserStatus   ServiceConfig `yaml:"userStatus"`
		UserDescribe ServiceConfig `yaml:"userDescribe"`
	} `yaml:"services"`
}
