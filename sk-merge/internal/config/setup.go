package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"
	"path/filepath"
	"skas/sk-common/pkg/misc"
)

var yes = true
var no = false

func Setup() error {
	var configFile string
	var version bool
	var logLevel string
	var logMode string

	pflag.StringVar(&configFile, "configFile", "config.yaml", "Configuration file")
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
	_, err := misc.LoadConfig(configFile, &Conf)
	if err != nil {
		return err
	}
	misc.AdjustConfigString(pflag.CommandLine, &Conf.Log.Mode, "logMode")
	misc.AdjustConfigString(pflag.CommandLine, &Conf.Log.Level, "logLevel")

	// ----------------------------------- Adjust path from config file path
	base := filepath.Dir(configFile)
	Conf.RootCaPath = misc.AdjustPath(base, Conf.RootCaPath)
	for idx, _ := range Conf.Providers {
		Conf.Providers[idx].HttpClient.RootCaPath = misc.AdjustPath(base, Conf.Providers[idx].HttpClient.RootCaPath)
	}
	// ----------------------------------- Handle logging  stuff
	Log, err = misc.HandleLog(&Conf.Log)
	if err != nil {
		return err
	}
	// ------------------------------------ Handle providers config
	for idx, _ := range Conf.Providers {
		Conf.Providers[idx].Init()
	}
	// ------------------------------------- Handle servers config
	// If the server list is empty, a first and only one is added.
	if Conf.Servers == nil || len(Conf.Servers) == 0 {
		Conf.Servers = []MergeServerConfig{MergeServerConfig{}}
	}
	for idx, _ := range Conf.Servers {
		Conf.Servers[idx].Default(7013 + (idx * 100))
	}
	return nil
}

func (c *ProviderConfig) Init() {
	// Set default values
	if c.Enabled == nil {
		c.Enabled = &yes
	}
	if c.CredentialAuthority == nil {
		c.CredentialAuthority = &yes
	}
	if c.GroupAuthority == nil {
		c.GroupAuthority = &yes
	}
	if c.Critical == nil {
		c.Critical = &yes
	}
	if c.GroupPattern == "" {
		c.GroupPattern = "%s"
	}
}
