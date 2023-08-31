package user

import "github.com/spf13/cobra"

var bindFlagsVars struct {
	strict      bool
	namespace   string
	bindingName string
}

func addBindFlags(c *cobra.Command) {
	c.PersistentFlags().BoolVar(&bindFlagsVars.strict, "strict", false, "Generate error if binding already exists")
	c.PersistentFlags().StringVarP(&bindFlagsVars.namespace, "namespace", "n", "", "User's DB namespace")
	c.PersistentFlags().StringVar(&bindFlagsVars.bindingName, "bindingName", "", "K8s binding Name. Default to <user>:<group>")
}
