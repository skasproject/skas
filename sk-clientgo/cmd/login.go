package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"skas/sk-clientgo/internal/config"
	"skas/sk-clientgo/internal/tokenbag"
)

var login string
var password string

func init() {
	LoginCmd.PersistentFlags().StringVar(&login, "user", "", "User name")
	LoginCmd.PersistentFlags().StringVar(&password, "password", "", "User password")
}

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Logout and get a new token",
	Run: func(cmd *cobra.Command, args []string) {
		config.InitHttpClient()
		tokenbag.DeleteTokenBag() // Logout first. Don't stay logged with old token if we are unable to login
		tokenBag := tokenbag.InteractiveLogin(login, password)
		if tokenBag == nil {
			os.Exit(3)
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "logged successfully..\n")
		}
	},
}
