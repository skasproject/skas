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
	var metricAddr string
	var probeAddr string

	var inactivityTimeout string
	var sessionMaxTTL string
	var clientTokenTTL string
	var tokenStorage string
	var tokenNamespace string
	var lastHitStep int

	pflag.StringVar(&configFile, "configFile", "config.yaml", "Configuration file")
	pflag.BoolVar(&version, "version", false, "Display version info")
	pflag.StringVar(&logLevel, "logLevel", "INFO", "Log level (PANIC|FATAL|ERROR|WARN|INFO|DEBUG|TRACE)")
	pflag.StringVar(&logMode, "logMode", "json", "Log mode: 'dev' or 'json'")
	pflag.StringVar(&bindAddr, "bindAddr", "127.0.0.1:7014", "Server bind address <host>:<port>")
	pflag.StringVar(&metricAddr, "metricAddr", ":8080", "Metrics bind address (\"0\" to disable)")
	pflag.StringVar(&probeAddr, "probeAddr", ":8181", "Probe bind address (\"0\" to disable)\"")

	pflag.StringVar(&inactivityTimeout, "inactivityTimeout", "30m", "Session inactivity time out")
	pflag.StringVar(&sessionMaxTTL, "sessionMaxTTL", "24h", "Session max TTL")
	pflag.StringVar(&clientTokenTTL, "clientTokenTTL", "30s", "Client local token TTL")
	pflag.StringVar(&tokenStorage, "tokenStorageType", "memory", "Tokens storage mode: 'memory' or 'crd'")
	pflag.StringVar(&tokenNamespace, "tokenNamespace", "skas-system", "Tokens storage namespace when tokenStorage==crd")
	pflag.IntVar(&lastHitStep, "lastHitStep", 3, "Delay to store lastHit in CRD, when tokenStorage==crd. In % of inactivityTimeout")

	pflag.CommandLine.SortFlags = false
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
	misc.AdjustConfigString(pflag.CommandLine, &Conf.MetricAddr, "metricAddr")
	misc.AdjustConfigString(pflag.CommandLine, &Conf.ProbeAddr, "probeAddr")

	misc.AdjustConfigDuration(pflag.CommandLine, &Conf.TokenConfig.InactivityTimeout, "inactivityTimeout")
	misc.AdjustConfigDuration(pflag.CommandLine, &Conf.TokenConfig.SessionMaxTTL, "sessionMaxTTL")
	misc.AdjustConfigDuration(pflag.CommandLine, &Conf.TokenConfig.ClientTokenTTL, "clientTokenTTL")
	misc.AdjustConfigString(pflag.CommandLine, &Conf.TokenConfig.StorageType, "tokenStorageType")
	misc.AdjustConfigString(pflag.CommandLine, &Conf.TokenConfig.Namespace, "tokenNamespace")
	misc.AdjustConfigInt(pflag.CommandLine, &Conf.TokenConfig.LastHitStep, "lastHitStep")

	// -----------------------------------Handle logging  stuff
	Log, err = misc.HandleLog(&Conf.Log)
	if err != nil {
		return err
	}

	// Handle some defaults

	return nil
}
