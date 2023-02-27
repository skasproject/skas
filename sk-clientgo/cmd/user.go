package cmd

import (
	"github.com/spf13/cobra"
)

var namespace string

func init() {
	UserCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "skas-system", "User's DB namespace")
}

var UserCmd = &cobra.Command{
	Use:   "user",
	Short: "Skas user management",
}
