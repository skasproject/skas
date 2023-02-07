package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:    "auth",
	Short:  "To be used as client-go exec plugin",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("AUTH")
		//common.InitHttpConnection()
		//tokenBag := common.RetrieveTokenBag()
		//if tokenBag == nil {
		//	tokenBag = common.DoLoginSilently("", "")
		//}
		//ec := ExecCredential{
		//	ApiVersion: "client.authentication.k8s.io/v1beta1",
		//	Kind:       "ExecCredential",
		//}
		//if tokenBag == nil {
		//	// No tokenBag
		//} else {
		//	ec.Status.Token = tokenBag.Token
		//}
		//err := json.NewEncoder(os.Stdout).Encode(ec)
		//if err != nil {
		//	panic(err)
		//}
	},
}

type ExecCredential struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Status     struct {
		Token string `json:"token"`
	} `json:"status"`
}
