package user

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"skas/sk-clientgo/internal/global"
	"skas/sk-clientgo/internal/k8sclient"
	userdbv1alpha1 "skas/sk-common/k8sapis/userdb/v1alpha1"
	"strings"
)

func init() {
	addBindFlags(BindCmd)
}

var BindCmd = &cobra.Command{
	Use:   "bind <user> <group>",
	Short: "Include a user to a group",
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
			// Binding not found. can create
			binding := &userdbv1alpha1.GroupBinding{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: kc.Namespace,
					Name:      bindingName,
				},
				Spec: userdbv1alpha1.GroupBindingSpec{
					User:  args[0],
					Group: args[1],
				},
			}
			err = kc.Client.Create(context.Background(), binding)
			if err != nil {
				global.Log.Error(err, "Kubernetes error while creating the groupBinding: %s", err.Error())
				os.Exit(3)
			}
			fmt.Printf("GroupBinding '%s' created in namespace '%s'.\n", bindingName, kc.Namespace)
			os.Exit(0)

		} else {
			// Binding found.
			if bindFlagsVars.strict {
				fmt.Printf("ERROR: GroupBinding '%s' already exists in namespace '%s'.\n", bindingName, kc.Namespace)
				os.Exit(1)
			} else {
				// Check binding is a coherent one
				if binding.Spec.User != args[0] {
					fmt.Printf("ERROR: Uncoherent binding '%s' found in namespace '%s': User is '%s' instead of '%s'\n", bindingName, kc.Namespace, binding.Spec.User, args[0])
					os.Exit(5)
				}
				if binding.Spec.Group != args[1] {
					fmt.Printf("ERROR: Uncoherent binding '%s' found in namespace '%s': Group is '%s' instead of '%s'\n", bindingName, kc.Namespace, binding.Spec.Group, args[1])
					os.Exit(6)
				}
				fmt.Printf("GroupBinding '%s' already exists in namespace '%s'.\n", bindingName, kc.Namespace)
				os.Exit(0)
			}
		}
	},
}

func buildBindingName(bindingName, user, group string) string {
	if bindingName == "" {
		bindingName = fmt.Sprintf("%s.%s", user, group)
		bindingName = strings.Replace(bindingName, ":", ".", -1)
		bindingName = strings.Replace(bindingName, "_", ".", -1)
	}
	return bindingName
}
