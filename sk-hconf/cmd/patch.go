package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"skas/sk-hconf/internal/readiness"
	"time"
)

var patchParams struct {
	nodeName string
	remove   bool
	timeout  time.Duration
	mark     bool
}

func init() {
	PatchCmd.PersistentFlags().BoolVar(&patchParams.remove, "remove", false, "Remove webhook configuration")
	PatchCmd.PersistentFlags().StringVar(&patchParams.nodeName, "nodeName", "", "Node Name")
	PatchCmd.PersistentFlags().DurationVar(&patchParams.timeout, "timeout", time.Second*60, "Timeout on API server down or up")
	PatchCmd.PersistentFlags().BoolVar(&patchParams.mark, "mark", false, "Display dot on pod state change wait. Log if false")
	_ = PatchCmd.MarkPersistentFlagRequired("nodeName")
}

var PatchCmd = &cobra.Command{
	Use:   "patch",
	Short: "Patch an api server configuration",
	Run: func(cmd *cobra.Command, args []string) {
		// First, ensure the api server is ready
		probe, err := readiness.GetProbe(patchParams.nodeName)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to find readiness probe: %v\n", err)
			os.Exit(2)
		}
		err = probe.IsReady()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "API server pod is not ready. Will not patch (err:%v)\n", err)
			os.Exit(3)
		}
		// Patch
		err = probe.WaitForDown(patchParams.timeout, patchParams.mark)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(3)
		}
		err = probe.WaitForUp(patchParams.timeout, patchParams.mark)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(3)
		}
		fmt.Printf("\nSUCCESS!!\n")
	},
}
