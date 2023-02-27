package user

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	addFlags(CreateCmd)
}

var CreateCmd = &cobra.Command{
	Use:   "create <user>",
	Short: "Create a new user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Create user....")
	},
}
