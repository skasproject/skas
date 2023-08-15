package user

import "github.com/spf13/cobra"

var userFlagsVars struct {
	namespace        string
	email            string
	commonName       string
	uid              int
	comment          string
	password         string
	passwordHash     string
	generatePassword bool
	inputPassword    bool
	state            string
}

func addUserFlags(c *cobra.Command) {
	c.PersistentFlags().StringVarP(&userFlagsVars.namespace, "namespace", "n", "", "User's DB namespace")
	c.PersistentFlags().StringVar(&userFlagsVars.email, "email", "", "User's email")
	c.PersistentFlags().StringVar(&userFlagsVars.commonName, "commonName", "", "User's common name")
	c.PersistentFlags().IntVar(&userFlagsVars.uid, "uid", 0, "User's UID")
	c.PersistentFlags().StringVar(&userFlagsVars.comment, "comment", "", "User's comment")
	c.PersistentFlags().StringVar(&userFlagsVars.password, "password", "", "User's password")
	c.PersistentFlags().StringVar(&userFlagsVars.passwordHash, "passwordHash", "", "User's password hash (Result of 'kubectl skas hash')")
	c.PersistentFlags().BoolVar(&userFlagsVars.generatePassword, "generatePassword", false, "Generate and display a password")
	c.PersistentFlags().BoolVar(&userFlagsVars.inputPassword, "inputPassword", false, "Interactive password request")
	c.PersistentFlags().StringVar(&userFlagsVars.state, "state", "enabled", "User's state (enabled|disabled)")
}
