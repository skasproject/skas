package user

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"skas/sk-clientgo/internal/global"
	"skas/sk-clientgo/internal/k8sclient"
	userdbv1alpha1 "skas/sk-common/k8sapis/userdb/v1alpha1"
)

func init() {
	addBindFlags(UnbindCmd)
}

var UnbindCmd = &cobra.Command{
	Use:   "unbind <user> <group>",
	Short: "Remove a user from a group",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		kc := k8sclient.New(userdbv1alpha1.SchemeBuilder, global.KubeconfigPath, bindFlagsVars.namespace)
		bindingName := buildBindingName(bindFlagsVars.bindingName, args[0], args[1])
		// Lookup binding
		binding := &userdbv1alpha1.GroupBinding{}
		err := kc.Client.Get(context.Background(), client.ObjectKey{
			Namespace: kc.Namespace,
			Name:      bindingName,
		}, binding)
		if client.IgnoreNotFound(err) != nil {
			global.Log.Error(err, "API server communication error")
			os.Exit(3)
		}
		if err != nil {
			// Binding not found.
			if bindFlagsVars.strict {
				fmt.Printf("ERROR: GroupBinding '%s' not found in namespace '%s'.\n", bindingName, kc.Namespace)
				os.Exit(3)
			} else {
				fmt.Printf("GroupBinding '%s' not found in namespace '%s'.\n", bindingName, kc.Namespace)
				fmt.Printf("")
				os.Exit(0)
			}

		} else {
			// Binding found. Must delete
			err := kc.Client.Delete(context.Background(), binding)
			if err != nil {
				global.Log.Error(err, "API server communication error")
				os.Exit(3)
			}
			fmt.Printf("GroupBinding '%s' in namespace '%s' has been deleted.\n", bindingName, kc.Namespace)
		}
	},
}
