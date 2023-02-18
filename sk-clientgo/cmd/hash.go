package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strings"
	"syscall"
)

var HashCmd = &cobra.Command{
	Use:   "hash",
	Short: "Provided password hash, for use in config file",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		passwd := ""
		if len(args) >= 1 {
			passwd = args[0]
		}
		if passwd == "" {
			passwd = inputPassword("Password:")
			passwd2 := inputPassword("Confirm password:")
			if passwd != passwd2 {
				fmt.Printf("Passwords did not match!\n")
				return
			}
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", string(hash))
	},
}

func inputPassword(prompt string) string {
	_, err := fmt.Fprint(os.Stdout, prompt)
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
