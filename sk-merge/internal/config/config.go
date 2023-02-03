package config

import (
	"github.com/go-logr/logr"
	"skas/sk-common/pkg/client"
	"skas/sk-common/pkg/httpclient"
	"skas/sk-common/pkg/httpserver"
	"skas/sk-common/pkg/misc"
)

// Exported vars

var (
	Conf Config
	Log  logr.Logger
)

// NB: All RootCA will be cumulated

type ClientProviderConfig struct {
	Name                string            `yaml:"name"`
	HttpClientConfig    httpclient.Config `yaml:"httpClient"`
	Enabled             *bool             `yaml:"enabled"`             // Allow to disable a provider
	CredentialAuthority *bool             `yaml:"credentialAuthority"` // Is this ldap is authority for password checking
	GroupAuthority      *bool             `yaml:"groupAuthority"`      // Group will be fetched. Default true
	Critical            *bool             `yaml:"critical"`            // If true (default), a failure on this provider will leads 'invalid login'. Even if another provider grants access
	GroupPattern        string            `yaml:"groupPattern"`        // Group pattern. Default "%s"
	UidOffset           int64             `yaml:"uidOffset"`           // Will be added to the returned Uid. Default to 0
	Client              client.Config     `yaml:"client"`
}

type ServiceConfig struct {
	Enabled bool            `yaml:"enabled"`
	Clients []client.Config `yaml:"clients"`
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
