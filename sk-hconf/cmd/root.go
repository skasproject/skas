package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"skas/sk-hconf/internal/global"
	"skas/sk-hconf/internal/misc"
	"skas/sk-hconf/pkg/k8sapi"
)

var rootParams struct {
	logConfig  misc.LogConfig
	kubeconfig string
	configFile string
}

func init() {
	RootCmd.AddCommand(PatchCmd)
	RootCmd.PersistentFlags().StringVar(&rootParams.logConfig.Level, "logLevel", "INFO", "Log level")
	RootCmd.PersistentFlags().StringVar(&rootParams.logConfig.Mode, "logMode", "dev", "Log mode: 'dev' or 'json'")
	RootCmd.PersistentFlags().StringVar(&rootParams.kubeconfig, "kubeconfig", "", "kubeconfig file path. Override default configuration.")
	RootCmd.PersistentFlags().StringVar(&rootParams.configFile, "configFile", "config.yaml", "Configuration file")

}

var RootCmd = &cobra.Command{
	Use:   "hconf",
	Short: "Authentication Webhook configurator",
	Long:  "A tool to configure the SKAS authentication webhook in the K8S API server",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error
		global.Logger, err = misc.HandleLog(&rootParams.logConfig)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to set logging configuration: %v\n", err)
			os.Exit(2)
		}
		err = misc.LoadYaml(rootParams.configFile, &global.Config)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to load configuration: %v\n", err)
			os.Exit(2)
		}
		global.KubeClient, err = k8sapi.GetKubeClient(rootParams.kubeconfig)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to initialize kubernetes client: %v\n", err)
			os.Exit(1)
		}

	},
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
