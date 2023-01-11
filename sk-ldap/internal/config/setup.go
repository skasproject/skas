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

	pflag.StringVar(&configFile, "configFile", "config.yaml", "Configuration file")
	pflag.BoolVar(&version, "version", false, "Display version info")
	pflag.StringVar(&logLevel, "logLevel", "", "Log level (PANIC|FATAL|ERROR|WARN|INFO|DEBUG|TRACE)")
	pflag.StringVar(&logMode, "logMode", "", "Log mode: 'dev' or 'json'")

	pflag.Parse()

	// ------------------------------------ Version display
	if version {
		fmt.Printf("%s\n", Version)
		os.Exit(0)
	}

	// ------------------------------------ Load config file
	var err error
	ConfigFile, err = filepath.Abs(configFile)
	if err != nil {
		return err
	}
	file, err := os.Open(ConfigFile)
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(file)
	decoder.SetStrict(true)
	if err = decoder.Decode(&Conf); err != nil {
		return err
	}

	// -----------------------------------Handle logging  stuff
	Log, err = misc.HandleLog(&Conf.Log, logLevel, logMode)
	if err != nil {
		return err
	}
	// ------------------------------------------- Set some default
	if Conf.Server.BindAddr == "" {
		Conf.Server.BindAddr = "127.0.0.1:7011"
	}
	return nil
}
