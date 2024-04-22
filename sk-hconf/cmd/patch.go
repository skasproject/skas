package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path"
	"skas/sk-hconf/internal/global"
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
		if patchParams.remove {
			err = unConfigure()
		} else {
			err = configure()
		}
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error while modify configuration: %v\n", err)
			os.Exit(3)
		}

		// And wait for a restart cycle
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

const hookConfig = "hookConfig.yaml"
const skasAuthCa = "skas_auth_ca.crt"

func configure() error {
	// Create skas folder
	err := makeDirectoryIfNotExists(global.Config.SkasFolder)
	if err != nil {
		return err
	}
	// And copy the hookConfig.yaml file
	hc := path.Join(global.Config.SkasFolder, hookConfig)
	err = os.WriteFile(hc, []byte(global.Config.HookConfigContent), 0600)
	if err != nil {
		return err
	}
	// And now the sk-auth CA certificate
	secret, err := global.ClientSet.CoreV1().Secrets(global.Config.CertificateAuthority.Secret.Namespace).Get(context.Background(), global.Config.CertificateAuthority.Secret.Name, v1.GetOptions{})
	if err != nil {
		return err
	}
	ca, ok := secret.Data[global.Config.CertificateAuthority.KeyInData]
	if !ok {
		return fmt.Errorf("unable to find data[%s] in secret %s:%s", global.Config.CertificateAuthority.KeyInData, global.Config.CertificateAuthority.Secret.Namespace, global.Config.CertificateAuthority.Secret.Name)
	}
	caf := path.Join(global.Config.SkasFolder, skasAuthCa)
	err = os.WriteFile(caf, ca, 0600)
	if err != nil {
		return err
	}
	return nil
}

func makeDirectoryIfNotExists(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.Mkdir(path, os.ModeDir|0755)
		} else {
			return err
		}
	}
	return nil
}

func unConfigure() error {
	return os.RemoveAll(global.Config.SkasFolder)
}
