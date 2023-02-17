package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"skas/sk-clientgo/internal/config"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "display skas client version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(config.Version)
	},
}
