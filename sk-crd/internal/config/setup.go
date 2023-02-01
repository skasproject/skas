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
	var bindAddr string
	var namespace string
	var metricAddr string
	var probeAddr string

	pflag.StringVar(&configFile, "configFile", "config.yaml", "Configuration file")
	pflag.BoolVar(&version, "version", false, "Display version info")
	pflag.StringVar(&logLevel, "logLevel", "INFO", "Log level (PANIC|FATAL|ERROR|WARN|INFO|DEBUG|TRACE)")
	pflag.StringVar(&logMode, "logMode", "json", "Log mode: 'dev' or 'json'")
	pflag.StringVar(&bindAddr, "bindAddr", "127.0.0.1:7012", "Server bind address <host>:<port>")
	pflag.StringVar(&namespace, "namespace", "skas-userdb", "Namespace hosting user definition")
	pflag.StringVar(&metricAddr, "metricAddr", ":8080", "Metrics bind address (\"0\" to disable)")
	pflag.StringVar(&probeAddr, "probeAddr", ":8181", "Probe bind address (\"0\" to disable)\"")

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
	misc.AdjustConfigString(pflag.CommandLine, &Conf.Server.BindAddr, "bindAddr")
	misc.AdjustConfigString(pflag.CommandLine, &Conf.Namespace, "namespace")
	misc.AdjustConfigString(pflag.CommandLine, &Conf.MetricAddr, "metricAddr")
	misc.AdjustConfigString(pflag.CommandLine, &Conf.ProbeAddr, "probeAddr")

	// -----------------------------------Handle logging  stuff
	Log, err = misc.HandleLog(&Conf.Log)
	if err != nil {
		return err
	}

	return nil
}
