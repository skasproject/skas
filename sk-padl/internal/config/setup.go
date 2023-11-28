package config

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
	"os"
	"path/filepath"
	"regexp"
	"skas/sk-common/pkg/config"
	"skas/sk-common/pkg/misc"
)

// Exported vars

var (
	Conf                      Config
	Log                       logr.Logger
	CertPath                  string
	KeyPath                   string
	UidFromUserFilterRegexes  []*regexp.Regexp
	UidFromGroupFilterRegexes []*regexp.Regexp
	UidFromDnRegexes          []*regexp.Regexp
)

var yes = true

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
		fmt.Printf("%s\n", config.Version)
		os.Exit(0)
	}

	// ------------------------------------ Load config file

	file, err := misc.LoadConfig(configFile, &Conf)
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

	// ------------------------------------- Handle Default values and performs some basic checks
	if Conf.UsersBaseDn == "" {
		Conf.UsersBaseDn = "ou=users,dc=skasproject,dc=com"
	}
	if Conf.GroupsBaseDn == "" {
		Conf.GroupsBaseDn = "ou=groups,dc=skasproject,dc=com"
	}
	if Conf.RoBindDn == "" {
		Conf.RoBindDn = "cn=readonly,dc=system,dc=skasproject,dc=com"
	}
	if Conf.RoBindPassword == "" {
		return fmt.Errorf("'roBindPassword' must be defined in configuration")
	}
	if Conf.Interface == "" {
		Conf.Interface = "0.0.0.0"
	}
	if Conf.Ssl == nil {
		Conf.Ssl = &yes
	}
	if Conf.Port == 0 {
		if *Conf.Ssl {
			Conf.Port = 636
		} else {
			Conf.Port = 389
		}
	}
	if *Conf.Ssl {
		if Conf.CertName == "" {
			Conf.CertName = "tls.crt"
		}
		if Conf.KeyName == "" {
			Conf.KeyName = "tls.key"
		}
		if Conf.CertDir == "" {
			return fmt.Errorf("'certDir' must be defined as 'ssl' is activated")
		}
		Conf.CertDir = misc.AdjustPath(filepath.Dir(file), Conf.CertDir)
		CertPath = filepath.Join(Conf.CertDir, Conf.CertName)
		err = misc.EnsureReadable(CertPath)
		if err != nil {
			return err
		}
		KeyPath = filepath.Join(Conf.CertDir, Conf.KeyName)
		err = misc.EnsureReadable(KeyPath)
		if err != nil {
			return err
		}
	}
	if Conf.UidFromUserFilterRegexes == nil || len(Conf.UidFromUserFilterRegexes) == 0 {
		Conf.UidFromUserFilterRegexes = []string{
			`^\(uid=(\w+)\)$`,
			`^\(\&\(objectClass=inetOrgPerson\)\(uid=(\w+)\)\)$`,
		}
	}
	if Conf.UidFromGroupFilterRegexes == nil || len(Conf.UidFromGroupFilterRegexes) == 0 {
		Conf.UidFromGroupFilterRegexes = []string{
			`^\(memberUid=(\w+)\)$`,
			`^\(\&\(objectClass=groupOfUniqueNames\)\(memberUid=(\w+)\)\)$`,
			`^\(\&\(objectClass=groupOfNames\)\(memberUid=(\w+)\)\)$`,
		}
	}
	if Conf.UidFromDnRegexes == nil || len(Conf.UidFromDnRegexes) == 0 {
		Conf.UidFromDnRegexes = []string{
			`^uid=(\w+),.*$`,
		}
	}

	UidFromUserFilterRegexes = make([]*regexp.Regexp, 0, 10)
	for idx, expr := range Conf.UidFromUserFilterRegexes {
		re, err := regexp.Compile(expr)
		if err != nil {
			return fmt.Errorf("unable to compile UidFromUserFilterRegexes[%d]: %w", idx, err)
		}
		UidFromUserFilterRegexes = append(UidFromUserFilterRegexes, re)
	}

	UidFromGroupFilterRegexes = make([]*regexp.Regexp, 0, 10)
	for idx, expr := range Conf.UidFromGroupFilterRegexes {
		re, err := regexp.Compile(expr)
		if err != nil {
			return fmt.Errorf("unable to compile UidFromGroupFilterRegexes[%d]: %w", idx, err)
		}
		UidFromGroupFilterRegexes = append(UidFromGroupFilterRegexes, re)
	}

	UidFromDnRegexes = make([]*regexp.Regexp, 0, 10)
	for idx, expr := range Conf.UidFromDnRegexes {
		re, err := regexp.Compile(expr)
		if err != nil {
			return fmt.Errorf("unable to compile UidFromDnRegexes[%d]: %w", idx, err)
		}
		UidFromDnRegexes = append(UidFromDnRegexes, re)
	}

	return nil
}
