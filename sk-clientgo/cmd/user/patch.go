package user

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"skas/sk-clientgo/internal/global"
	"skas/sk-clientgo/internal/httpClient"
	"skas/sk-clientgo/internal/k8sclient"
	"skas/sk-clientgo/internal/utils"
	userdbv1alpha1 "skas/sk-common/k8sapis/userdb/v1alpha1"
	"skas/sk-common/pkg/skclient"
	"strings"
)

var patchFlagsVars struct {
	create bool
}

func init() {
	PatchCmd.PersistentFlags().BoolVar(&patchFlagsVars.create, "create", false, "Create if not exists")
	addUserFlags(PatchCmd)
}

var PatchCmd = &cobra.Command{
	Use:   "patch <user>",
	Short: "Patch an existing user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		skClient, err := httpClient.New()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
			//global.Log.Error(err, "error on http client init")
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
			// User not found
			if patchFlagsVars.create {
				fmt.Printf("User '%s' not found. Will create it\n", args[0])
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
				fmt.Printf("ERROR: User '%s' not found in namespace '%s'.\nSet '--create' flag to allow creation.\n", args[0], kc.Namespace)
				os.Exit(1)
			}
		} else {
			// User found
			global.Log.V(1).Info("User found", "user", user.Name, "commonNames", user.Spec.CommonNames, "uid", user.Spec.Uid)
			populateUserFromFlags(skClient, &user.Spec)
			err = kc.Client.Update(context.Background(), user)
			if err != nil {
				global.Log.Error(err, "Kubernetes error while creating the user: %s", err.Error())
				os.Exit(3)
			}
			fmt.Printf("User '%s' updated in namespace '%s'.\n", args[0], kc.Namespace)
			os.Exit(0)
		}
	},
}

var False = false
var True = true

func populateUserFromFlags(client skclient.SkClient, user *userdbv1alpha1.UserSpec) {
	if userFlagsVars.commonName != "" {
		if user.CommonNames == nil {
			user.CommonNames = make([]string, 1)
		}
		user.CommonNames[0] = userFlagsVars.commonName
	}
	if userFlagsVars.email != "" {
		if user.Emails == nil {
			user.Emails = make([]string, 1)
		}
		user.Emails[0] = userFlagsVars.email
	}
	if userFlagsVars.comment != "" {
		user.Comment = userFlagsVars.comment
	}
	if userFlagsVars.uid != 0 {
		user.Uid = &userFlagsVars.uid
	}
	if userFlagsVars.state != "" {
		st := strings.ToLower(strings.TrimSpace(userFlagsVars.state))
		if st == "enabled" {
			user.Disabled = &False
		} else if st == "disabled" {
			user.Disabled = &True
		} else {
			fmt.Printf("Invalide '--state' value: '%s'. Must be 'enabled' or 'disabled'\n", userFlagsVars.state)
			os.Exit(1)
		}
	}
	hash := handlePasswordHash(client)

	if hash != "" {
		user.PasswordHash = hash
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func handlePasswordHash(client skclient.SkClient) string {
	if userFlagsVars.passwordHash != "" {
		// WARNING: In this case, passwordStrength is not checked.
		global.Log.Info("WARNING: When providing password by its hash, there is no password strength test. Ensure password is not too simple.")
		return userFlagsVars.passwordHash
	}
	if userFlagsVars.password != "" {
		return testAndHashPassword(client, userFlagsVars.password)
	}
	if userFlagsVars.generatePassword {
		b := make([]byte, 12)
		for i := range b {
			b[i] = letterBytes[rand.Intn(len(letterBytes))]
		}
		p := string(b)
		fmt.Printf("The following password has been generated: '%s'\n(Save it as it will not be accessible anymore).\n", p)
		return utils.Hash(p)
	}
	if userFlagsVars.inputPassword {
		for count := 3; count > 0; count++ {
			passwd := utils.InputPassword("Password:")
			passwd2 := utils.InputPassword("Confirm password:")
			if passwd != passwd2 {
				fmt.Printf("Passwords did not match! Retry\n")
			} else {
				return testAndHashPassword(client, passwd)
			}
		}
		fmt.Printf("Too many retry. Aborting\n")
		os.Exit(1)
	}
	return ""
}

func testAndHashPassword(client skclient.SkClient, password string) string {
	if !utils.PasswordCheck(client, password, nil, nil) {
		fmt.Printf("Unsatisfactory password strength!\n")
		os.Exit(2)
	}
	return utils.Hash(password)
}
