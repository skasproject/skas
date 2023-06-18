package utils

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"skas/sk-clientgo/internal/global"
	"skas/sk-common/pkg/skclient"
	"skas/sk-common/proto/v1/proto"
	"strings"
	"syscall"
)

// Return password hash

func Hash(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hash)
}

func InputPassword(prompt string) string {
	_, err := fmt.Fprint(os.Stderr, prompt)
	if err != nil {
		panic(err)
	}
	bytePassword, err2 := terminal.ReadPassword(int(syscall.Stdin))
	if err2 != nil {
		panic(err2)
	}
	_, _ = fmt.Fprintf(os.Stderr, "\n")
	return strings.TrimSpace(string(bytePassword))
}

// PasswordCheck test if password is acceptable, based on server validation rules
func PasswordCheck(client skclient.SkClient, password string, user *proto.User, extUserInput []string) bool {
	userInputs := make([]string, 0, 5)
	if user != nil {
		userInputs = append(userInputs, user.Login)
		userInputs = append(userInputs, user.Emails...)
		userInputs = append(userInputs, user.Groups...)
		userInputs = append(userInputs, user.CommonNames...)
	}
	if extUserInput != nil {
		userInputs = append(userInputs, extUserInput...)
	}
	passwordStrengthRequest := &proto.PasswordStrengthRequest{
		ClientAuth: client.GetClientAuth(),
		Password:   password,
		UserInputs: userInputs,
	}
	response := proto.PasswordStrengthResponse{}
	err := client.Do(proto.PasswordStrengthMeta, passwordStrengthRequest, &response, nil)
	if err != nil {
		global.Log.Error(err, "error on PasswordChangeRequest")
		os.Exit(4)
	}
	global.Log.V(1).Info("PasswordStrengthResponse", "acceptable", response.Acceptable, "score", response.Score, "isCommon", response.IsCommon)
	return response.Acceptable
}
