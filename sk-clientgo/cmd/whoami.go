package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"skas/sk-clientgo/internal/config"
	"skas/sk-clientgo/internal/log"
	"skas/sk-clientgo/internal/tokenbag"
	"skas/sk-common/pkg/misc"
	"strings"
	"text/tabwriter"
)

var all bool

func init() {
	WhoamiCmd.PersistentFlags().BoolVar(&all, "all", false, "Add 'technical' informations")
}

var WhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display current logged user, if any",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		err := config.InitHttpClient()
		if err != nil {
			log.Log.Error(err, "error on InitHttpClient()")
			os.Exit(10)
		}
		tokenBag := tokenbag.Retrieve()
		if tokenBag != nil {
			tw := new(tabwriter.Writer)
			tw.Init(os.Stdout, 2, 4, 3, ' ', 0)
			if all {
				_, _ = fmt.Fprintf(tw, "USER\tID\tGROUPS\tAUTH.\tTOKEN")
				_, _ = fmt.Fprintf(tw, "\n%s\t%d\t%s\t%s\t%s", tokenBag.User.Login, tokenBag.User.Uid, strings.Join(tokenBag.User.Groups, ","), tokenBag.Authority, misc.ShortenString(tokenBag.Token))
			} else {
				_, _ = fmt.Fprintf(tw, "USER\tID\tGROUPS")
				_, _ = fmt.Fprintf(tw, "\n%s\t%d\t%s", tokenBag.User.Login, tokenBag.User.Uid, strings.Join(tokenBag.User.Groups, ","))
			}
			_, _ = fmt.Fprintf(tw, "\n")
			_ = tw.Flush()
		} else {
			fmt.Printf("Nobody! (Not logged)\n")
			os.Exit(3)
		}
	},
}
