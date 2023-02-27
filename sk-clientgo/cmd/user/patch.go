package user

import (
	"fmt"
	"github.com/spf13/cobra"
)

var patchFlagsVars struct {
	create bool
}

func init() {
	PatchCmd.PersistentFlags().BoolVar(&patchFlagsVars.create, "create", false, "Create if not exists")
	addFlags(PatchCmd)
}

var PatchCmd = &cobra.Command{
	Use:   "patch <user>",
	Short: "Patch an existing user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Patch user '%s'...", args[1])
		//kc, err := k8sclient.New(userdbv1alpha1.SchemeBuilder, )
	},
}
