package config

import (
	"github.com/go-logr/logr"
	cconfig "skas/sk-common/pkg/config"
	"skas/sk-common/pkg/misc"
	"skas/sk-common/pkg/skclient"
)

// Exported vars

var (
	Conf Config
	Log  logr.Logger
)

// NB: All RootCA will be cumulated

type ClientProviderConfig struct {
	Name                string          `yaml:"name"`
	HttpClient          skclient.Config `yaml:"httpClient"`
	Enabled             *bool           `yaml:"enabled"`             // Allow to disable a provider
	CredentialAuthority *bool           `yaml:"credentialAuthority"` // Is this ldap is authority for password checking
	GroupAuthority      *bool           `yaml:"groupAuthority"`      // Group will be fetched. Default true
	Critical            *bool           `yaml:"critical"`            // If true (default), a failure on this provider will leads 'invalid login'. Even if another provider grants access
	GroupPattern        string          `yaml:"groupPattern"`        // Group pattern. Default "%s"
	UidOffset           int             `yaml:"uidOffset"`           // Will be added to the returned Uid. Default to 0
}

type Config struct {
	Log       misc.LogConfig         `yaml:"log"`
	Server    cconfig.SkServerConfig `yaml:"server"`
	Providers []ClientProviderConfig `yaml:"providers"`
	// values added to above Providers
	RootCaPath string `yaml:"rootCaPath"` // Path to a trusted root CA file
	RootCaData string `yaml:"rootCaData"` // Base64 encoded PEM data containing root CA
	Services   struct {
		Login          cconfig.ServiceConfig `yaml:"login"`
		UserIdentity   cconfig.ServiceConfig `yaml:"userIdentity"`
		UserDescribe   cconfig.ServiceConfig `yaml:"userDescribe"`
		PasswordChange cconfig.ServiceConfig `yaml:"passwordChange"`
	} `yaml:"services"`
}
