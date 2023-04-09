package user

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"skas/sk-clientgo/httpClient"
	"skas/sk-clientgo/internal/global"
	"skas/sk-clientgo/internal/tokenbag"
	"skas/sk-clientgo/internal/utils"
	"skas/sk-common/pkg/skclient"
	"skas/sk-common/proto/v1/proto"
	"strings"
	"text/tabwriter"
)

var password string
var explain bool
var inputPassword bool

func init() {
	DescribeCmd.PersistentFlags().BoolVar(&explain, "explain", false, "Describe by provider")
	DescribeCmd.PersistentFlags().StringVar(&password, "password", "", "User's password")
	DescribeCmd.PersistentFlags().BoolVar(&inputPassword, "inputPassword", false, "Interactive password request")

}

var DescribeCmd = &cobra.Command{
	Use:   "describe <user>",
	Short: "Describe a user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := httpClient.New()
		if err != nil {
			global.Log.Error(err, "error on InitHttpClient()")
			os.Exit(10)
		}
		tokenBag := tokenbag.Retrieve(client)
		if tokenBag == nil {
			tokenBag = tokenbag.InteractiveLogin(client, "", "")
		}
		if tokenBag == nil {
			os.Exit(3)
		}
		if inputPassword {
			password = utils.InputPassword(fmt.Sprintf("Password for user '%s':", args[0]))
		}
		uer := &proto.IdentityRequest{
			ClientAuth: client.GetClientAuth(),
			Login:      args[0],
			Password:   password,
			Detailed:   true,
		}
		resp := &proto.IdentityResponse{}
		err = client.Do(proto.IdentityMeta, uer, resp, &skclient.HttpAuth{Token: tokenBag.Token})
		if err != nil {
			_, ok := err.(*skclient.UnauthorizedError)
			if ok {
				_, _ = fmt.Fprintf(os.Stderr, "Unauthorized!\n")
				os.Exit(2)
			}
			global.Log.Error(err, "error on UserDescribeRequest()")
			os.Exit(4)
		}
		//fmt.Printf("%v", resp)

		tw := new(tabwriter.Writer)
		tw.Init(os.Stdout, 2, 4, 3, ' ', 0)
		twb := newTabwriterBuffer()
		twb.add("USER", "%s", resp.Login)
		twb.add("STATUS", "%s", resp.Status)
		twb.add("UID", "%d", resp.Uid)
		twb.add("GROUPS", "%s", strings.Join(resp.Groups, ","))
		twb.add("EMAILS", "%s", strings.Join(resp.Emails, ","))
		twb.add("COMMON NAMES", "%s", strings.Join(resp.CommonNames, ","))
		twb.add("AUTH", "%s", resp.Authority)
		twb.endOfLine(tw)
		_ = tw.Flush()
		if explain {
			fmt.Printf("\nDetail:\n")
			tw := new(tabwriter.Writer)
			tw.Init(os.Stdout, 2, 4, 3, ' ', 0)
			twb := newTabwriterBuffer()
			for _, item := range resp.Details {
				twb.add("PROVIDER", "%s", item.Provider.Name)
				twb.add("STATUS", "%s", item.Status)
				twb.add("UID", "%d", item.Uid)
				twb.add("GROUPS", "%s", strings.Join(item.Groups, ","))
				twb.add("EMAILS", "%s", strings.Join(item.Emails, ","))
				twb.add("COMMON NAMES", "%s", strings.Join(item.CommonNames, ","))
				twb.endOfLine(tw)
			}
			_ = tw.Flush()
		}
	},
}

type tabwriterBuffer struct {
	head      string
	tags      string
	values    []interface{}
	firstLine bool
	firstCell bool
}

func newTabwriterBuffer() *tabwriterBuffer {
	t := &tabwriterBuffer{
		head:   "",
		tags:   "",
		values: make([]interface{}, 0, 20),
	}
	t.firstLine = true
	t.firstCell = true
	return t
}

func (t *tabwriterBuffer) add(title string, tag string, value interface{}) {
	if t.firstLine {
		if !t.firstCell {
			t.head += "\t"
			t.tags += "\t"
		}
		t.head += title
		t.tags += tag
	}
	t.values = append(t.values, value)
	t.firstCell = false
}

func (t *tabwriterBuffer) endOfLine(tw *tabwriter.Writer) {
	if t.firstLine {
		t.head += "\n"
		t.tags += "\n"
		_, _ = fmt.Fprintf(tw, t.head)
	}
	_, _ = fmt.Fprintf(tw, t.tags, t.values...)
	t.firstLine = false
	t.firstCell = true
	t.values = make([]interface{}, 0, 20)
}
