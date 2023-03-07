package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	usercmd "skas/sk-clientgo/cmd/user"
	"skas/sk-clientgo/internal/global"
	"skas/sk-common/pkg/misc"
)

var RootCmd = &cobra.Command{
	Use:   "kubectl-skas",
	Short: "A kubectl plugin for Kubernetes authentication",
}

// Used in init.go
var kubeconfigPath string

func init() {
	var logConfig misc.LogConfig

	// We must declare child in parent.
	// Performing RootCmd.AddCommand(...) in the child init() function will not works as there is chance the child package will not be loaded, as not imported by any package.
	RootCmd.AddCommand(authCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(LoginCmd)
	RootCmd.AddCommand(LogoutCmd)
	RootCmd.AddCommand(WhoamiCmd)
	RootCmd.AddCommand(InitCmd)
	RootCmd.AddCommand(HashCmd)
	RootCmd.AddCommand(UserCmd)
	UserCmd.AddCommand(usercmd.CreateCmd)
	UserCmd.AddCommand(usercmd.PatchCmd)
	UserCmd.AddCommand(usercmd.BindCmd)
	UserCmd.AddCommand(usercmd.UnbindCmd)
	UserCmd.AddCommand(usercmd.DescribeCmd)

	RootCmd.PersistentFlags().StringVar(&logConfig.Level, "logLevel", "INFO", "Log level")
	RootCmd.PersistentFlags().StringVar(&logConfig.Mode, "logMode", "dev", "Log mode: 'dev' or 'json'")
	RootCmd.PersistentFlags().StringVar(&global.KubeconfigPath, "kubeconfig", "", "kubeconfig file path. Override default configuration.")

	RootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		var err error
		global.Log, err = misc.HandleLog(&logConfig)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to load logging configuration: %v\n", err)
			os.Exit(2)
		}
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
