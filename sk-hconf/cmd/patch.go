package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var PatchCmd = &cobra.Command{
	Use:   "patch",
	Short: "Patch an api server configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Allo!!\n")
	},
}
