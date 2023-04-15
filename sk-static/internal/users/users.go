package users

import (
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

// -----------------------------------------------------
// This is the format of the users file

type StaticUsers struct {
	Users         []StaticUser         `yaml:"users"`
	GroupBindings []StaticGroupBinding `yaml:"groupBindings"`
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

// --------------------------------------------------------

// This is the object returned by the parser

type Content struct {
	UserByLogin       map[string]StaticUser
	GroupsByUser      map[string][]string
	GroupBindingCount int
}

func Parse(fileName string) (interface{}, error) {
	fn, err := filepath.Abs(fileName)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	decoder := yaml.NewDecoder(file)
	decoder.SetStrict(true)
	staticUsers := StaticUsers{}
	if err = decoder.Decode(&staticUsers); err != nil {
		return nil, err
	}
	content := &Content{
		UserByLogin:       make(map[string]StaticUser),
		GroupsByUser:      make(map[string][]string),
		GroupBindingCount: len(staticUsers.GroupBindings),
	}
	for idx, _ := range staticUsers.Users {
		content.UserByLogin[staticUsers.Users[idx].Login] = staticUsers.Users[idx]
	}
	for _, gb := range staticUsers.GroupBindings {
		u := gb.User
		g := gb.Group
		groups, ok := content.GroupsByUser[u]
		if ok {
			content.GroupsByUser[u] = append(groups, g)
		} else {
			content.GroupsByUser[u] = []string{g}
		}
	}
	return content, nil
}
