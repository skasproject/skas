package config

import (
	cconfig "skas/sk-common/pkg/config"
	"skas/sk-common/pkg/misc"
	"skas/sk-common/pkg/skclient"
)

type Config struct {
	cconfig.SkServerConfig    `yaml:",inline"`
	Log                       misc.LogConfig  `yaml:"log"`
	UsersBaseDn               string          `yaml:"usersBaseDn"`  // Default: "ou=users,dc=skasproject,dc=com"
	GroupsBaseDn              string          `yaml:"groupsBaseDn"` // Default: "ou=groups,dc=skasproject,dc=com"
	RoBindDn                  string          `yaml:"roBindDn"`     // Default: "cm=readonly,dc=system,dc=skasproject,dc=com"
	RoBindPassword            string          `yaml:"roBindPassword"`
	UidFromUserFilterRegexes  []string        `yaml:"uidFromUserFilterRegexes"`
	UidFromGroupFilterRegexes []string        `yaml:"uidFromGroupFilterRegexes"`
	UidFromDnRegexes          []string        `yaml:"uidFromDnRegexes"`
	Provider                  skclient.Config `yaml:"provider"`
}
