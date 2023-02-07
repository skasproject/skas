package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path"
	"skas/sk-clientgo/internal/config"
	"skas/sk-clientgo/internal/kubecontext"
	"skas/sk-clientgo/internal/log"
	"skas/sk-common/pkg/misc"
	"skas/sk-common/pkg/skhttp"
	"strings"
)

var RootCmd = &cobra.Command{
	Use:   "kubectl-skas",
	Short: "A kubectl plugin for Kubernetes authentification",
}

func init() {
	var logConfig misc.LogConfig
	var kubeconfig string
	var server skhttp.Config
	var reset bool

	// We must declare child in parent.
	// Performing RootCmd.AddCommand(...) in the child init() function will not works as there is chance the child package will not be loaded, as not imported by any package.
	RootCmd.AddCommand(authCmd)
	RootCmd.AddCommand(contextCmd)
	RootCmd.AddCommand(versionCmd)

	RootCmd.PersistentFlags().StringVar(&kubecontext.KubeContext, "Context", "", "Context")
	RootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Kubeconfig file path. Used to lookup Context")
	RootCmd.PersistentFlags().StringVar(&logConfig.Level, "logLevel", "DEBUG", "Log level")
	RootCmd.PersistentFlags().StringVar(&logConfig.Mode, "logMode", "dev", "Log mode: 'dev' or 'json'")

	RootCmd.PersistentFlags().BoolVar(&reset, "reset", false, "Reset configuration")

	RootCmd.PersistentFlags().StringVar(&server.RootCaPath, "rootCaPath", "", "Path to a trusted root CA file for client connection")
	RootCmd.PersistentFlags().StringVar(&server.RootCaData, "rootCaData", "", "Base64 encoded PEM data containing root CA for client connection")
	RootCmd.PersistentFlags().StringVar(&server.Url, "serverUrl", "", "Authentication server")
	RootCmd.PersistentFlags().StringVar(&server.ClientAuth.Id, "clientId", "", "Client ID for authentication server")
	RootCmd.PersistentFlags().StringVar(&server.ClientAuth.Secret, "clientSecret", "", "Client secret")
	RootCmd.PersistentFlags().BoolVar(&server.InsecureSkipVerify, "insecureSkipVerify", false, "Skip server certificate validation")

	kubecontext.KubeContext = strings.Trim(kubecontext.KubeContext, "\"")
	kubeconfig = strings.Trim(kubeconfig, "\"")
	logConfig.Level = strings.Trim(logConfig.Level, "\"")
	logConfig.Mode = strings.Trim(logConfig.Mode, "\"")
	server.RootCaPath = strings.Trim(server.RootCaPath, "\"")
	server.RootCaData = strings.Trim(server.RootCaData, "\"")
	server.Url = strings.Trim(server.Url, "\"")
	server.ClientAuth.Id = strings.Trim(server.ClientAuth.Id, "\"")
	server.ClientAuth.Secret = strings.Trim(server.ClientAuth.Secret, "\"")

	RootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		var err error
		log.Log, err = misc.HandleLog(&logConfig)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to load logging configuration: %v\n", err)
			os.Exit(2)
		}
		kubecontext.Initialize(kubeconfig)

		if server.RootCaPath != "" {
			if !path.IsAbs(server.RootCaPath) {
				cwd, err := os.Getwd()
				if err != nil {
					panic(err)
				}
				server.RootCaPath = path.Join(cwd, server.RootCaPath)
			}
		}
		if cmd != contextCmd && cmd != versionCmd && cmd != contextListCmd {
			config.Load()
			if config.Conf == nil || reset {
				config.Conf = &config.Config{
					Config: server,
				}
				checkConfig(config.Conf)
				config.Save()
			} else {
				dirtyConfig := false
				if server.Url != "" && server.Url != config.Conf.Url {
					config.Conf.Url = server.Url
					dirtyConfig = true
				}
				if server.RootCaPath != "" && server.RootCaPath != config.Conf.RootCaPath {
					config.Conf.RootCaPath = server.RootCaPath
					dirtyConfig = true
				}
				if server.RootCaData != "" && server.RootCaData != config.Conf.RootCaData {
					config.Conf.RootCaData = server.RootCaData
					dirtyConfig = true
				}
				if server.ClientAuth.Id != "" && server.ClientAuth.Id != config.Conf.ClientAuth.Id {
					config.Conf.ClientAuth.Id = server.ClientAuth.Id
					dirtyConfig = true
				}
				if server.ClientAuth.Secret != "" && server.ClientAuth.Secret != config.Conf.ClientAuth.Secret {
					config.Conf.ClientAuth.Secret = server.ClientAuth.Secret
					dirtyConfig = true
				}
				if server.InsecureSkipVerify != config.Conf.InsecureSkipVerify {
					config.Conf.InsecureSkipVerify = server.InsecureSkipVerify
				}
				checkConfig(config.Conf)
				if dirtyConfig {
					config.Save()
				}
			}
		}
	}
}

func checkConfig(conf *config.Config) {
	if conf.Url == "" {
		_, _ = fmt.Fprintf(os.Stderr, "\nERROR: Missing 'serverUrl' parameter on initial call\n\n")
		os.Exit(2)
	}
	if strings.HasPrefix(strings.ToLower(conf.Url), "https") && !conf.InsecureSkipVerify {
		if conf.RootCaPath == "" && conf.RootCaData == "" {
			_, _ = fmt.Fprintf(os.Stderr, "\nERROR: Missing 'rootCaPath' or 'rootCaData' parameter on initial call\n\n")
			os.Exit(2)
		}
	}
	if conf.ClientAuth.Id == "" || conf.ClientAuth.Secret == "" {
		_, _ = fmt.Fprintf(os.Stderr, "\nERROR: Missing 'clientId' and/or 'clientSecret' parameters on initial call\n\n")
		os.Exit(2)
	}
}

var debug = true

func Execute() {
	defer func() {
		if !debug {
			if r := recover(); r != nil {
				fmt.Printf("ERROR:%v\n", r)
				os.Exit(1)
			}
		}
	}()
	if err := RootCmd.Execute(); err != nil {
		//fmt.Println(err)
		os.Exit(2)
	}
}
