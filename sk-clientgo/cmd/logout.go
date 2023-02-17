package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"skas/sk-clientgo/internal/tokenbag"
)

var LogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear local token",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		tokenbag.DeleteTokenBag()
		fmt.Printf("Bye!\n")
	},
}
