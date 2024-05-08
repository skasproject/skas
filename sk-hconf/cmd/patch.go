package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"path"
	"skas/sk-hconf/internal/global"
	"skas/sk-hconf/internal/readiness"
	"skas/sk-hconf/pkg/filepatcher"
	"skas/sk-hconf/pkg/texttemplate"
	"time"
)

var patchParams struct {
	nodeName           string
	remove             bool
	mark               bool
	force              bool
	patcherTemplate    string
	hookConfigTemplate string
}

func init() {
	PatchCmd.PersistentFlags().BoolVar(&patchParams.remove, "remove", false, "Remove webhook configuration")
	PatchCmd.PersistentFlags().BoolVar(&patchParams.force, "force", false, "Perform even if apiserver is down")
	PatchCmd.PersistentFlags().StringVar(&patchParams.nodeName, "nodeName", "", "Node Name")
	PatchCmd.PersistentFlags().BoolVar(&patchParams.mark, "mark", false, "Display dot on pod state change wait. Log if false")
	PatchCmd.PersistentFlags().StringVar(&patchParams.hookConfigTemplate, "hookConfigTemplate", "/templates/hookconfig.tmpl", "hookconfig file template")
	PatchCmd.PersistentFlags().StringVar(&patchParams.patcherTemplate, "patcherTemplate", "/templates/patcher.tmpl", "patcher file template")
	_ = PatchCmd.MarkPersistentFlagRequired("nodeName")
}

var PatchCmd = &cobra.Command{
	Use:   "patch",
	Short: "Patch an api server configuration",
	Run: func(cmd *cobra.Command, args []string) {

		global.Logger.Info("Auth webhook node configurator", "version", global.Version, "build", global.BuildTs, "logLevel", rootParams.logConfig.Level, "nodeName", patchParams.nodeName, "remove", patchParams.remove)

		// First, ensure the api server is ready
		probe, err := readiness.GetProbe(patchParams.nodeName)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to find readiness probe: %v\n", err)
			os.Exit(2)
		}
		err = probe.IsReady()
		if err != nil {
			if !patchParams.force {
				_, _ = fmt.Fprintf(os.Stderr, "API server pod is not ready. Will not patch (err:%v)\n", err)
				os.Exit(3)
			} else {
				global.Logger.Info("API server pod is NOT ready. Perform operation anyway")
			}
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
		if patchParams.remove {
			// In case of removal, when retrying, there will be no restart, as the kube-apiserver manifest will not change
			// So no down state is a 'normal' case
			_ = probe.WaitForDown(time.Second*30, patchParams.mark)
		} else {
			err = probe.WaitForDown(global.Config.TimeoutApiServer, patchParams.mark)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(3)
			}
		}
		err = probe.WaitForUp(global.Config.TimeoutApiServer, patchParams.mark)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(3)
		}
		fmt.Printf("\nSUCCESS!!\n")
	},
}

const hookConfig = "hookconfig.yaml"
const skasAuthCa = "skas_auth_ca.crt"

func configure() error {
	// Create skas folder
	if err := makeDirectoryIfNotExists(global.Config.SkasFolder); err != nil {
		return err
	}
	if err := makeDirectoryIfNotExists(global.Config.BackupFolder); err != nil {
		return err
	}
	if err := makeDirectoryIfNotExists(global.Config.TmpFolder); err != nil {
		return err
	}
	// And copy the hookConfig.yaml file
	model := map[string]interface{}{
		"Config": global.Config,
	}
	hookConfigTxt, err := texttemplate.NewAndRenderToTextFromFile(patchParams.hookConfigTemplate, model)
	if err != nil {
		return err
	}
	hc := path.Join(global.Config.SkasFolder, hookConfig)
	if err := os.WriteFile(hc, []byte(hookConfigTxt), 0600); err != nil {
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
	if err := os.WriteFile(caf, ca, 0600); err != nil {
		return err
	}
	if err := patchApiServerManifest(false); err != nil {
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
	err := patchApiServerManifest(true)
	if err != nil {
		return err
	}
	return os.RemoveAll(global.Config.SkasFolder)
}

const block1 = `- mountPath: /etc/kubernetes/skas
  name: skas-config`

const block2 = `- hostPath:
    path: /etc/kubernetes/skas
    type: ""
  name: skas-config`

func patchApiServerManifest(remove bool) error {

	model := map[string]interface{}{
		"Config": global.Config,
		"Values": map[string]interface{}{
			"remove":     remove,
			"nowRFC3339": time.Now().Format(time.RFC3339),
		},
	}
	patchOp, err := texttemplate.NewAndRenderToTextFromFile(patchParams.patcherTemplate, model)
	if err != nil {
		return err
	}

	//fmt.Printf("\n%s\n", patchOp)
	patchOperation := &filepatcher.PatchOperation{}
	err = yaml.UnmarshalStrict([]byte(patchOp), patchOperation)
	if err != nil {
		return err
	}

	err = patchOperation.Run()
	if err != nil {
		return err
	}
	return nil
}
