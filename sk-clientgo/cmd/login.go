package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"skas/sk-clientgo/internal/config"
	"skas/sk-clientgo/internal/log"
	"skas/sk-clientgo/internal/tokenbag"
)

var LoginCmd = &cobra.Command{
	Use:   "login [user, [password]]",
	Short: "Logout and get a new token",
	Args:  cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var login string
		var password string
		if len(args) >= 1 {
			login = args[0]
			if len(args) >= 2 {
				password = args[1]
			}
		}
		err := config.InitHttpClient()
		if err != nil {
			log.Log.Error(err, "error on InitHttpClient()")
			os.Exit(10)
		}
		tokenbag.DeleteTokenBag() // Logout first. Don't stay logged with old token if we are unable to login
		tokenBag := tokenbag.InteractiveLogin(login, password)
		if tokenBag == nil {
			os.Exit(3)
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "logged successfully..\n")
		}
	},
}

//
//
//func(c *cobra.Command) error {
//	c.mergePersistentFlags()
//	err := tmpl(c.OutOrStderr(), c.UsageTemplate(), c)
//	if err != nil {
//		c.PrintErrln(err)
//	}
//	return err
//}

//// UsageFunc returns either the function set by SetUsageFunc for this command
//// or a parent, or it returns a default usage function.
//func (c *Command) UsageFunc() (f func(*Command) error) {
//	if c.usageFunc != nil {
//		return c.usageFunc
//	}
//	if c.HasParent() {
//		return c.Parent().UsageFunc()
//	}
//	return func(c *Command) error {
//		c.mergePersistentFlags()
//		err := tmpl(c.OutOrStderr(), c.UsageTemplate(), c)
//		if err != nil {
//			c.PrintErrln(err)
//		}
//		return err
//	}
//}
