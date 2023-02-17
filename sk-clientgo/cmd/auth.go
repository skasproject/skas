package cmd

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"os"
	"skas/sk-clientgo/internal/config"
	"skas/sk-clientgo/internal/log"
	"skas/sk-clientgo/internal/tokenbag"
)

// This is intended to be used as client-go exc plugin. It communicates by a json printed on stdout.
// So, not other print to stdout should be performed. Use stderr to display messages to the user

var authCmd = &cobra.Command{
	Use:    "auth",
	Short:  "To be used as client-go exec plugin",
	Hidden: true,
	Args:   cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		err := config.InitHttpClient()
		if err != nil {
			log.Log.Error(err, "error on InitHttpClient()")
			os.Exit(10)
		}
		tokenBag := tokenbag.Retrieve()
		if tokenBag == nil {
			tokenBag = tokenbag.InteractiveLogin("", "")
		}
		ec := ExecCredential{
			ApiVersion: "client.authentication.k8s.io/v1beta1",
			Kind:       "ExecCredential",
		}
		if tokenBag == nil {
			// No tokenBag
		} else {
			ec.Status.Token = tokenBag.Token
		}
		err = json.NewEncoder(os.Stdout).Encode(ec)
		if err != nil {
			panic(err)
		}
	},
}

type ExecCredential struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Status     struct {
		Token string `json:"token"`
	} `json:"status"`
}
