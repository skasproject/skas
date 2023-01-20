package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"skas/sk-common/pkg/misc"
)

func Setup() error {
	var configFile string
	var version bool
	var userFile string
	var logLevel string
	var logMode string

	pflag.StringVar(&configFile, "configFile", "config.yaml", "Configuration file")
	pflag.StringVar(&userFile, "userFile", "users.yaml", "Users file")
	pflag.BoolVar(&version, "version", false, "Display version info")
	pflag.StringVar(&logLevel, "logLevel", "INFO", "Log level (PANIC|FATAL|ERROR|WARN|INFO|DEBUG|TRACE)")
	pflag.StringVar(&logMode, "logMode", "json", "Log mode: 'dev' or 'json'")

	pflag.Parse()

	// ------------------------------------ Version display
	if version {
		fmt.Printf("%s\n", Version)
		os.Exit(0)
	}

	// ------------------------------------ Load config file
	fn, err := filepath.Abs(configFile)
	if err != nil {
		return err
	}
	file, err := os.Open(fn)
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(file)
	decoder.SetStrict(true)
	if err = decoder.Decode(&Conf); err != nil {
		return fmt.Errorf("file '%s': %w", configFile, err)
	}

	misc.AdjustConfigString(pflag.CommandLine, &Conf.Log.Mode, "logMode")
	misc.AdjustConfigString(pflag.CommandLine, &Conf.Log.Level, "logLevel")

	// -----------------------------------Handle logging  stuff
	Log, err = misc.HandleLog(&Conf.Log)
	if err != nil {
		return err
	}
	// ------------------------------------------- Set some default
	if Conf.Server.BindAddr == "" {
		Conf.Server.BindAddr = "127.0.0.1:7010"
	}

	// --------------------------------------- Load users file
	if err = loadUsers(userFile); err != nil {
		return fmt.Errorf("file '%s': %w", userFile, err)
	}
	return nil
}

func loadUsers(fileName string) error {
	fn, err := filepath.Abs(fileName)
	if err != nil {
		return err
	}
	file, err := os.Open(fn)
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(file)
	decoder.SetStrict(true)
	staticUsers := StaticUsers{}
	if err = decoder.Decode(&staticUsers); err != nil {
		return err
	}
	UserByLogin = make(map[string]StaticUser)
	for idx, _ := range staticUsers.Users {
		UserByLogin[staticUsers.Users[idx].Login] = staticUsers.Users[idx]
	}
	GroupsByUser = make(map[string][]string)
	for _, gb := range staticUsers.GroupBindings {
		u := gb.User
		g := gb.Group
		groups, ok := GroupsByUser[u]
		if ok {
			GroupsByUser[u] = append(groups, g)
		} else {
			GroupsByUser[u] = []string{g}
		}
	}
	return nil
}
