package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	contextCmd.AddCommand(contextListCmd)
}

var contextCmd = &cobra.Command{
	Use:     "context",
	Short:   "Manage contexts",
	Aliases: []string{"contexts"},
}

var contextListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Display local Context configuration",
	Aliases: []string{"contexts"},
	Run: func(cmd *cobra.Command, args []string) {
		panic("Not yet implemented")
		//
		//currentContext := common.Context
		//contexts := internal.ListContext()
		//tw := new(tabwriter.Writer)
		//tw.Init(os.Stdout, 2, 4, 3, ' ', 0)
		//_, _ = fmt.Fprintf(tw, " \tCONTEXT\tSERVER\tCA")
		//for _, ctx := range contexts {
		//	var mark string
		//	if ctx == currentContext {
		//		mark = "*"
		//	} else {
		//		mark = ""
		//	}
		//	myConfig := internal.LoadConfig(ctx)
		//	_, _ = fmt.Fprintf(tw, "\n%s\t%s\t%s\t%s", mark, ctx, myConfig.Server, myConfig.RootCaFile)
		//}
		//_, _ = fmt.Fprintf(tw, "\n")
		//_ = tw.Flush()
		////fmt.Printf("Contexts:%v\n", contexts)
	},
}
