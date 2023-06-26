package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"skas/sk-common/pkg/config"
)

var versionVarFlags struct {
	extended bool
}

func init() {
	versionCmd.PersistentFlags().BoolVar(&versionVarFlags.extended, "extended", false, "Add build number")
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "display skas client version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if versionVarFlags.extended {
			fmt.Printf("%s.%s\n", config.Version, config.BuildTs)
		} else {
			fmt.Printf("%s\n", config.Version)
		}
	},
}
