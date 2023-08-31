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

// NB: All RootCa will be cumulated

type ProviderConfig struct {
	Name                string `yaml:"name"`
	CredentialAuthority *bool  `yaml:"credentialAuthority"` // Is this ldap is authority for password checking
	GroupAuthority      *bool  `yaml:"groupAuthority"`      // Group will be fetched. Default true
	Critical            *bool  `yaml:"critical"`            // If true (default), a failure on this provider will leads 'invalid login'. Even if another provider grants access
	GroupPattern        string `yaml:"groupPattern"`        // Group pattern. Default "%s"
	UidOffset           int    `yaml:"uidOffset"`           // Will be added to the returned Uid. Default to 0
}

type MergeServerConfig struct {
	cconfig.SkServerConfig `yaml:",inline"`
	Services               struct {
		Identity       cconfig.ServiceConfig `yaml:"identity"`
		PasswordChange cconfig.ServiceConfig `yaml:"passwordChange"`
	} `yaml:"services"`
}

type Config struct {
	Log          misc.LogConfig              `yaml:"log"`
	Servers      []MergeServerConfig         `yaml:"servers"`
	Providers    []ProviderConfig            `yaml:"providers"`
	ProviderInfo map[string]*skclient.Config `yaml:"providerInfo"`
	// values added to above Providers
	RootCaPath string `yaml:"rootCaPath"` // Path to a trusted root CA file
	RootCaData string `yaml:"rootCaData"` // Base64 encoded PEM data containing root CA
}
