package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"
	"skas/sk-common/pkg/config"
	"skas/sk-common/pkg/misc"
)

func Setup() error {
	var configFile string
	var version bool
	var logLevel string
	var logMode string
	var adminGroups string
	var metricAddr string
	var probeAddr string

	var inactivityTimeout string
	var sessionMaxTTL string
	var clientTokenTTL string
	var tokenStorage string
	var namespace string
	var lastHitStep int

	pflag.StringVar(&configFile, "configFile", "config.yaml", "Configuration file")
	pflag.BoolVar(&version, "version", false, "Display version info")
	pflag.StringVar(&logLevel, "logLevel", "INFO", "Log level (PANIC|FATAL|ERROR|WARN|INFO|DEBUG|TRACE)")
	pflag.StringVar(&logMode, "logMode", "json", "Log mode: 'dev' or 'json'")
	pflag.StringVar(&adminGroups, "adminGroups", "skas-admin", "SKAS administrator groups (Allow user describe)")
	pflag.StringVar(&metricAddr, "metricAddr", ":8080", "Metrics bind address (\"0\" to disable)")
	pflag.StringVar(&probeAddr, "probeAddr", ":8181", "Probe bind address (\"0\" to disable)\"")

	pflag.StringVar(&inactivityTimeout, "inactivityTimeout", "30m", "Session inactivity time out")
	pflag.StringVar(&sessionMaxTTL, "sessionMaxTTL", "24h", "Session max TTL")
	pflag.StringVar(&clientTokenTTL, "clientTokenTTL", "30s", "Client local token TTL")
	pflag.StringVar(&tokenStorage, "tokenStorageType", "memory", "Tokens storage mode: 'memory' or 'crd'")
	pflag.StringVar(&namespace, "namespace", "skas-system", "Tokens storage namespace when tokenStorage==crd")
	pflag.IntVar(&lastHitStep, "lastHitStep", 3, "Delay to store lastHit in CRD, when tokenStorage==crd. In % of inactivityTimeout")

	pflag.CommandLine.SortFlags = false
	pflag.Parse()

	// ------------------------------------ Version display
	if version {
		fmt.Printf("%s\n", config.Version)
		os.Exit(0)
	}

	// ------------------------------------ Load config file
	_, err := misc.LoadConfig(configFile, &Conf)
	if err != nil {
		return err
	}

	misc.AdjustConfigString(pflag.CommandLine, &Conf.Log.Mode, "logMode")
	misc.AdjustConfigString(pflag.CommandLine, &Conf.Log.Level, "logLevel")
	misc.AdjustConfigStringArray(pflag.CommandLine, &Conf.AdminGroups, "adminGroups")
	misc.AdjustConfigString(pflag.CommandLine, &Conf.MetricAddr, "metricAddr")
	misc.AdjustConfigString(pflag.CommandLine, &Conf.ProbeAddr, "probeAddr")

	misc.AdjustConfigDuration(pflag.CommandLine, &Conf.Token.InactivityTimeout, "inactivityTimeout")
	misc.AdjustConfigDuration(pflag.CommandLine, &Conf.Token.SessionMaxTTL, "sessionMaxTTL")
	misc.AdjustConfigDuration(pflag.CommandLine, &Conf.Token.ClientTokenTTL, "clientTokenTTL")
	misc.AdjustConfigString(pflag.CommandLine, &Conf.Token.StorageType, "tokenStorageType")
	misc.AdjustConfigString(pflag.CommandLine, &Conf.Namespace, "namespace")
	misc.AdjustConfigInt(pflag.CommandLine, &Conf.Token.LastHitStep, "lastHitStep")

	// -----------------------------------Handle logging  stuff
	Log, err = misc.HandleLog(&Conf.Log)
	if err != nil {
		return err
	}

	// ------------------------------------- Handle servers config
	if Conf.Servers == nil || len(Conf.Servers) == 0 {
		return fmt.Errorf("at least one 'server' must be defined")
	}
	serverWithKubeconfigCount := 0
	for idx, srv := range Conf.Servers {
		if srv.Interface == "" {
			return fmt.Errorf("server[%d]: 'interface' must be defined", idx)
		}
		if srv.Port == 0 {
			return fmt.Errorf("server[%d]: 'port' must be defined", idx)
		}
		if srv.Ssl == nil {
			return fmt.Errorf("server[%d]: 'ssl' must be set to 'true' or 'false'", idx)
		}
		if !srv.Services.Kubeconfig.Disabled {
			serverWithKubeconfigCount++
		}
		if srv.Services.Kubeconfig.Protected {
			return fmt.Errorf("server[%d]: 'kubeconfig service can't be protected", idx)
		}
		if srv.Services.PasswordStrength.Protected {
			return fmt.Errorf("server[%d]: 'passwordStrength service can't be protected", idx)
		}
	}
	if serverWithKubeconfigCount > 0 {
		err = initKubeconfig(&Conf.Kubeconfig)
		if err != nil {
			return fmt.Errorf("error in Kubeconfig section: %w", err)
		}
	}
	return nil
}
