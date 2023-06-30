package user

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"skas/sk-clientgo/internal/global"
	"skas/sk-clientgo/internal/httpClient"
	"skas/sk-clientgo/internal/k8sclient"
	userdbv1alpha1 "skas/sk-common/k8sapis/userdb/v1alpha1"
)

func init() {
	addUserFlags(CreateCmd)
}

var CreateCmd = &cobra.Command{
	Use:   "create <user>",
	Short: "Create a new user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		skClient, err := httpClient.New()
		if err != nil {
			global.Log.Error(err, "error on http client init")
			os.Exit(10)
		}
		kc := k8sclient.New(userdbv1alpha1.SchemeBuilder, global.KubeconfigPath, userFlagsVars.namespace)
		// Lookup user
		user := &userdbv1alpha1.User{}
		err = kc.Client.Get(context.Background(), client.ObjectKey{
			Namespace: kc.Namespace,
			Name:      args[0],
		}, user)
		if client.IgnoreNotFound(err) != nil {
			global.Log.Error(err, "API server communication error")
			os.Exit(3)
		}
		if err != nil {
			// User not found. can crete
			user = &userdbv1alpha1.User{
				ObjectMeta: metav1.ObjectMeta{
					Name:      args[0],
					Namespace: kc.Namespace,
				},
			}
			populateUserFromFlags(skClient, &user.Spec)
			err = kc.Client.Create(context.Background(), user)
			if err != nil {
				global.Log.Error(err, "Kubernetes error while creating the user: %s", err.Error())
				os.Exit(3)
			}
			fmt.Printf("User '%s' created in namespace '%s'.\n", args[0], kc.Namespace)
			os.Exit(0)
		} else {
			fmt.Printf("ERROR: User '%s' allready exists in namespace '%s'. Unable to create\n", args[0], kc.Namespace)
			os.Exit(2)
		}
	},
}
