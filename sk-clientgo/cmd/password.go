package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"skas/sk-clientgo/httpClient"
	"skas/sk-clientgo/internal/global"
	"skas/sk-clientgo/internal/tokenbag"
	"skas/sk-clientgo/internal/utils"
	"skas/sk-common/proto/v1/proto"
)

var oldPassword string
var newPassword string

func init() {
	PasswordCmd.PersistentFlags().StringVar(&oldPassword, "oldPassword", "", "Old password")
	PasswordCmd.PersistentFlags().StringVar(&newPassword, "newPassword", "", "New password")
}

var PasswordCmd = &cobra.Command{
	Use:   "password",
	Short: "Change current password",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := httpClient.New()
		if err != nil {
			global.Log.Error(err, "error on http client init")
			os.Exit(10)
		}
		tokenBag := tokenbag.Retrieve(client)
		if tokenBag == nil {
			fmt.Printf("You must be logged to change your password!\n")
			os.Exit(2)
		}
		fmt.Printf("Will change password for user '%s'\n", tokenBag.User.Login)
		if oldPassword == "" {
			oldPassword = utils.InputPassword("Old password:")
		}
		if newPassword == "" {
			newPassword = utils.InputPassword("New password:")
			newPassword2 := utils.InputPassword("Confirm new password:")
			if newPassword != newPassword2 {
				fmt.Printf("New passwords did not match!\n")
				os.Exit(2)
			}
		}

		newPasswordHash := utils.Hash(newPassword)
		passwordChangeRequest := &proto.PasswordChangeRequest{
			ClientAuth:      client.GetClientAuth(),
			Provider:        tokenBag.Authority,
			Login:           tokenBag.User.Login,
			OldPassword:     oldPassword,
			NewPasswordHash: newPasswordHash,
		}
		resp := proto.PasswordChangeResponse{}
		err = client.Do(proto.PasswordChangeMeta, passwordChangeRequest, &resp, nil)
		if err != nil {
			global.Log.Error(err, "error on PasswordChangeRequest")
			os.Exit(4)
		}
		switch resp.Status {
		case proto.PasswordChanged:
			fmt.Printf("Password has been changed sucessfully.\n")
		case proto.InvalidOldPassword:
			fmt.Printf("ERROR: Invalid old password\n")
		case proto.InvalidNewPassword:
			fmt.Printf("ERROR: Invalid new password\n")
		case proto.Unsupported:
			fmt.Printf("SORRY: Your password can't be changed by this tool.\n")
		case proto.UnknownProvider:
			fmt.Printf("ERROR: Internal system error (Unknown provider '%s')\n", passwordChangeRequest.Provider)
		case proto.UserNotFound:
			fmt.Printf("ERROR:Internal system error (Unknown user '%s')\n", passwordChangeRequest.Login)
		default:
			fmt.Printf("Internal system error (Unknown status '%s')\n", resp.Status)
		}
	},
}
