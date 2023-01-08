package config

import (
	"fmt"
	"github.com/bombsimon/logrusr/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"strings"
)

var logLevelByString = map[string]logrus.Level{
	"PANIC": logrus.PanicLevel,
	"FATAL": logrus.FatalLevel,
	"ERROR": logrus.ErrorLevel,
	"WARN":  logrus.WarnLevel,
	"INFO":  logrus.InfoLevel,
	"DEBUG": logrus.DebugLevel,
	"TRACE": logrus.TraceLevel,
}

func Setup() error {
	var version bool
	var userFile string
	var logLevel string
	var logMode string

	pflag.StringVar(&userFile, "userFile", "users.yaml", "Users file")
	pflag.StringVar(&Config.BindAddr, "bindAddr", "127.0.0.1:7001", "The address to listen on")
	pflag.BoolVar(&version, "version", false, "Display version info")
	pflag.StringVar(&logLevel, "logLevel", "INFO", "Log level (PANIC|FATAL|ERROR|WARN|INFO|DEBUG|TRACE)")
	pflag.StringVar(&logMode, "logMode", "json", "Log mode: 'dev' or 'json'")
	pflag.BoolVar(&Config.NoSsl, "noSsl", false, "Server in plain text")
	pflag.StringVar(&Config.CertDir, "certDir", "", "TLS Certificate directory")
	pflag.StringVar(&Config.CertName, "certName", "tls.crt", "TLS Certificate name")
	pflag.StringVar(&Config.KeyName, "keyName", "tls.key", "TLS Certificate key name")

	pflag.Parse()

	// ------------------------------------ Version display
	if version {
		fmt.Printf("%s\n", Version)
		os.Exit(0)
	}

	// -----------------------------------Handle logging  stuff
	if logMode != "dev" && logMode != "json" {
		_, _ = fmt.Fprintf(os.Stderr, "\nERROR: Invalid logMode value: %s. Must be one of 'dev' or 'json'\n", logMode)
		os.Exit(2)
	}
	llevel, ok := logLevelByString[strings.ToUpper(logLevel)]
	if !ok {
		_, _ = fmt.Fprintf(os.Stderr, "\n%s is an invalid value for logLevel\n", logLevel)
		os.Exit(2)
	}

	logrusLog := logrus.New()
	logrusLog.SetLevel(llevel)
	if logMode == "json" {
		logrusLog.SetFormatter(&logrus.JSONFormatter{})
	}
	Config.Log = logrusr.New(logrusLog)

	// ------------------------------------------- Check CertDir
	if !Config.NoSsl && Config.CertDir == "" {
		_, _ = fmt.Fprintf(os.Stderr, "\nERROR: --certDir is not defined while --noSsl is not set\n")
		os.Exit(3)
	}

	// --------------------------------------- Load users file
	return loadUsers(userFile)
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
	Config.UserByLogin = make(map[string]StaticUser)
	for idx, _ := range staticUsers.Users {
		Config.UserByLogin[staticUsers.Users[idx].Login] = staticUsers.Users[idx]
	}
	return err

}
