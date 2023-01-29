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
	var logLevel string
	var logMode string
	var bindAddr string

	pflag.StringVar(&configFile, "configFile", "config.yaml", "Configuration file")
	pflag.BoolVar(&version, "version", false, "Display version info")
	pflag.StringVar(&logLevel, "logLevel", "INFO", "Log level (PANIC|FATAL|ERROR|WARN|INFO|DEBUG|TRACE)")
	pflag.StringVar(&logMode, "logMode", "json", "Log mode: 'dev' or 'json'")
	pflag.StringVar(&bindAddr, "bindAddr", "127.0.0.1:7013", "Server bind address <host>:<port>")

	pflag.Parse()

	// ------------------------------------ Version display
	if version {
		fmt.Printf("%s\n", Version)
		os.Exit(0)
	}

	// ------------------------------------ Load config file
	var err error
	configFile, err = filepath.Abs(configFile)
	if err != nil {
		return err
	}
	file, err := os.Open(configFile)
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(file)
	decoder.SetStrict(true)
	if err = decoder.Decode(&Conf); err != nil {
		return err
	}

	misc.AdjustConfigString(pflag.CommandLine, &Conf.Log.Mode, "logMode")
	misc.AdjustConfigString(pflag.CommandLine, &Conf.Log.Level, "logLevel")
	misc.AdjustConfigString(pflag.CommandLine, &Conf.Server.BindAddr, "bindAddr")

	// ----------------------------------- Adjust path from config file path
	base := filepath.Dir(configFile)
	Conf.RootCaPath = misc.AdjustPath(base, Conf.RootCaPath)
	for idx, _ := range Conf.Providers {
		Conf.Providers[idx].HttpClientConfig.RootCaPath = misc.AdjustPath(base, Conf.Providers[idx].HttpClientConfig.RootCaPath)
	}
	// ----------------------------------- Handle logging  stuff
	Log, err = misc.HandleLog(&Conf.Log)
	if err != nil {
		return err
	}
	for idx, _ := range Conf.Providers {
		Conf.Providers[idx].Init()
	}
	return nil
}

func (c *ClientProviderConfig) Init() {
	// Set default values
	yes := true
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
