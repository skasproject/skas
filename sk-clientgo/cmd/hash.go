package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"skas/sk-clientgo/internal/utils"
)

var HashCmd = &cobra.Command{
	Use:   "hash [password]",
	Short: "Provided password hash, for use in config file",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		passwd := ""
		if len(args) >= 1 {
			passwd = args[0]
		}
		if passwd == "" {
			passwd = utils.InputPassword("Password:")
			passwd2 := utils.InputPassword("Confirm password:")
			if passwd != passwd2 {
				fmt.Printf("Passwords did not match!\n")
				return
			}
		}
		hash := utils.Hash(passwd)
		fmt.Printf("%s\n", string(hash))
	},
}
