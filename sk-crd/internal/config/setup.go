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
	var metricAddr string
	var probeAddr string

	pflag.StringVar(&configFile, "configFile", "config.yaml", "Configuration file")
	pflag.BoolVar(&version, "version", false, "Display version info")
	pflag.StringVar(&logLevel, "logLevel", "", "Log level (PANIC|FATAL|ERROR|WARN|INFO|DEBUG|TRACE)")
	pflag.StringVar(&logMode, "logMode", "", "Log mode: 'dev' or 'json'")
	pflag.StringVar(&metricAddr, "metricAddr", "", "TCP address that the controller should bind to for serving prometheus metrics. '0' ti disable")
	pflag.StringVar(&probeAddr, "probeAddr", "", "TCP address that the controller should bind to for serving health probes. '0' to disable")

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

	// -----------------------------------Handle logging  stuff
	Log, err = misc.HandleLog(&Conf.Log, logLevel, logMode)
	if err != nil {
		return err
	}
	// ----------------------------------- Handle manager service
	if metricAddr != "" {
		Conf.MetricAddr = metricAddr
	}
	if probeAddr != "" {
		Conf.ProbeAddr = probeAddr
	}

	// ------------------------------------------- Set some default
	if Conf.Server.BindAddr == "" {
		Conf.Server.BindAddr = "127.0.0.1:7012"
	}
	if Conf.Namespace == "" {
		Conf.Namespace = "skas-userdb"
	}
	if Conf.MetricAddr == "" {
		Conf.MetricAddr = ":8080"
	}
	if Conf.ProbeAddr == "" {
		Conf.ProbeAddr = ":8181"
	}
	return nil
}
