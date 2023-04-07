package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"
	"skas/sk-common/pkg/misc"
)

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

	var err error
	File, err = misc.LoadConfig(configFile, &Conf)
	if err != nil {
		return err
	}

	misc.AdjustConfigString(pflag.CommandLine, &Conf.Log.Mode, "logMode")
	misc.AdjustConfigString(pflag.CommandLine, &Conf.Log.Level, "logLevel")

	// -----------------------------------Handle logging  stuff
	Log, err = misc.HandleLog(&Conf.Log)
	if err != nil {
		return err
	}

	// ------------------------------------- Handle servers config
	// If the server list is empty, a first and only one is added.
	if Conf.Servers == nil || len(Conf.Servers) == 0 {
		Conf.Servers = []LdapServerConfig{LdapServerConfig{}}
	}
	for idx, _ := range Conf.Servers {
		Conf.Servers[idx].Default(7011 + (100 * idx))
	}

	return nil
}
