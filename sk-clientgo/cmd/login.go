package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"skas/sk-clientgo/httpClient"
	"skas/sk-clientgo/internal/global"
	"skas/sk-clientgo/internal/tokenbag"
)

var LoginCmd = &cobra.Command{
	Use:   "login [user, [password]]",
	Short: "Logout and get a new token",
	Args:  cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var login string
		var password string
		if len(args) >= 1 {
			login = args[0]
			if len(args) >= 2 {
				password = args[1]
			}
		}
		client, err := httpClient.New()
		if err != nil {
			global.Log.Error(err, "Error on http client init")
			os.Exit(6)
		}
		tokenbag.DeleteTokenBag() // Logout first. Don't stay logged with old token if we are unable to login
		tokenBag := tokenbag.InteractiveLogin(client, login, password)
		if tokenBag == nil {
			os.Exit(3)
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "logged successfully..\n")
		}
	},
}
