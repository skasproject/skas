package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"skas/sk-clientgo/internal/config"
	"skas/sk-clientgo/internal/log"
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
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		err := config.InitHttpClient()
		if err != nil {
			log.Log.Error(err, "error on InitHttpClient()")
			os.Exit(10)
		}
		tokenbag.DeleteTokenBag() // Logout first. Don't stay logged with old token if we are unable to login
		tokenBag := tokenbag.InteractiveLogin(login, password)
		if tokenBag == nil {
			os.Exit(3)
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "logged successfully..\n")
		}
	},
}
