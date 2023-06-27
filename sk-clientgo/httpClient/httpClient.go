package httpClient

import (
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"os"
	osuser "os/user"
	"path"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"skas/sk-clientgo/internal/kubecontext"
	"skas/sk-clientgo/internal/loadsave"
	"skas/sk-common/pkg/misc"
	"skas/sk-common/pkg/skclient"
	"strconv"
)

var flags struct {
	server             skclient.Config
	reset              bool
	insecureSkipVerify string
}

func AddFlags(c *cobra.Command) {
	c.PersistentFlags().BoolVar(&flags.reset, "reset", false, "Reset configuration")
	c.PersistentFlags().StringVar(&flags.server.RootCaPath, "authRootCaPath", "", "Path to a trusted root CA file for client connection to skas auth server")
	c.PersistentFlags().StringVar(&flags.server.RootCaData, "authRootCaData", "", "Base64 encoded PEM data containing root CA for client connection to skas auth server")
	c.PersistentFlags().StringVar(&flags.server.Url, "authServerUrl", "", "Authentication server")
	c.PersistentFlags().StringVar(&flags.server.ClientAuth.Id, "clientId", "", "Client ID for authentication server")
	c.PersistentFlags().StringVar(&flags.server.ClientAuth.Secret, "clientSecret", "", "Client secret")
	c.PersistentFlags().StringVar(&flags.insecureSkipVerify, "authInsecureSkipVerify", "", "Skip skas auth server certificate validation ('true' or 'false')")
}

func groomFlags() {
	var err error
	if flags.insecureSkipVerify != "" {
		flags.server.InsecureSkipVerify, err = strconv.ParseBool(flags.insecureSkipVerify)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "authInsecureSkipVerify is not a valid boolean value")
			os.Exit(2)
		}
	}
	if flags.server.RootCaPath != "" {
		if !path.IsAbs(flags.server.RootCaPath) {
			cwd, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			flags.server.RootCaPath = path.Join(cwd, flags.server.RootCaPath)
		}
	}
}

func New() (skclient.SkClient, error) {
	groomFlags()
	return skclient.New(loadUpdateConfig(kubecontext.GetKubeContext()), "", "")
}

func NewForInit(serverUrl string) (skclient.SkClient, error) {
	groomFlags()
	conf := &flags.server
	if serverUrl != "" {
		if conf.Url != "" {
			_, _ = fmt.Fprintf(os.Stderr, "--authServerUrl should not be set on the 'init' command\n")
			os.Exit(2)
		}
		conf.Url = serverUrl
	}
	checkConfig(conf)
	return skclient.New(conf, "", "")
}

func loadUpdateConfig(kubeconfigFile string, kubeContext string) *skclient.Config {
	conf := loadConfig(kubeconfigFile, kubeContext)
	if conf == nil {
		conf = &flags.server
		checkConfig(conf)
		err := saveConfig(kubeconfigFile, kubeContext, conf)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%s", err.Error())
			os.Exit(3)
		}
	} else {
		dirtyConfig := false
		if (flags.server.Url != "" || flags.reset) && flags.server.Url != conf.Url {
			conf.Url = flags.server.Url
			dirtyConfig = true
		}
		if (flags.server.RootCaPath != "" || flags.reset) && flags.server.RootCaPath != conf.RootCaPath {
			conf.RootCaPath = flags.server.RootCaPath
			dirtyConfig = true
		}
		if (flags.server.RootCaData != "" || flags.reset) && flags.server.RootCaData != conf.RootCaData {
			conf.RootCaData = flags.server.RootCaData
			dirtyConfig = true
		}
		if (flags.server.ClientAuth.Id != "" || flags.reset) && flags.server.ClientAuth.Id != conf.ClientAuth.Id {
			conf.ClientAuth.Id = flags.server.ClientAuth.Id
			dirtyConfig = true
		}
		if (flags.server.ClientAuth.Secret != "" || flags.reset) && flags.server.ClientAuth.Secret != conf.ClientAuth.Secret {
			conf.ClientAuth.Secret = flags.server.ClientAuth.Secret
			dirtyConfig = true
		}
		if (flags.insecureSkipVerify != "" || flags.reset) && flags.server.InsecureSkipVerify != conf.InsecureSkipVerify {
			conf.InsecureSkipVerify = flags.server.InsecureSkipVerify
			dirtyConfig = true
		}
		checkConfig(conf)
		if dirtyConfig {
			err := saveConfig(kubeconfigFile, kubeContext, conf)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "%s", err.Error())
				os.Exit(3)
			}

		}
	}
	return conf
}

func checkConfig(conf *skclient.Config) {
	if conf.Url == "" {
		_, _ = fmt.Fprintf(os.Stderr, "\nERROR: Missing 'authServerUrl' parameter\n\n")
		os.Exit(2)
	}
	// We may use a certificate recognized by the system
	//if strings.HasPrefix(strings.ToLower(conf.Url), "https") && !conf.InsecureSkipVerify {
	//	if conf.RootCaPath == "" && conf.RootCaData == "" {
	//		_, _ = fmt.Fprintf(os.Stderr, "\nERROR: Missing 'rootCaPath' or 'rootCaData' parameter\n\n")
	//		os.Exit(2)
	//	}
	//}
	// Client.id/secret can be "" if server accept such config
	//if conf.ClientAuth.Id == "" || conf.ClientAuth.Secret == "" {
	//	_, _ = fmt.Fprintf(os.Stderr, "\nERROR: Missing 'clientId' and/or 'clientSecret' parameters on initial call\n\n")
	//	os.Exit(2)
	//}
}

func loadConfig(kubeconfigFile string, kubeContext string) *skclient.Config {
	conf := &skclient.Config{}
	configPath := buildPath(kubeconfigFile, kubeContext)
	if loadsave.LoadStuff(configPath, func(decoder *yaml.Decoder) error {
		return decoder.Decode(conf)
	}) {
		log.Log.V(1).Info("LoadConfig()", "path", configPath, "server", conf.Url, "rootCaPath", conf.RootCaPath, "rootCaData", misc.ShortenString(conf.RootCaData), "clientId", conf.ClientAuth.Id, "clientSecret", "*****", "insecureSkipVerify", conf.InsecureSkipVerify)
		return conf
	} else {
		log.Log.V(1).Info("LoadConfig() -> nil", "configPath", configPath)
		return nil
	}
}

func saveConfig(kubeconfigFile string, kubeContext string, conf *skclient.Config) error {
	configPath := buildPath(kubeconfigFile, kubeContext)
	log.Log.V(1).Info("SaveConfig()", "configPath", configPath, "server", conf.Url, "rootCaPath", conf.RootCaPath, "rootCaData", misc.ShortenString(conf.RootCaData), "clientId", conf.ClientAuth.Id, "clientSecret", "*****", "insecureSkipVerify", conf.InsecureSkipVerify)
	err := loadsave.SaveStuff(configPath, func(encoder *yaml.Encoder) error {
		return encoder.Encode(conf)
	})
	if err != nil {
		return fmt.Errorf("error while saving configuration in '%s': %w", configPath, err)
	}
	return nil
}

func buildPath(kubeconfigFile string, context string) string {
	usr, err := osuser.Current()
	if err != nil {
		panic(err)
	}
	return path.Join(usr.HomeDir, ".kube/cache/skas", kubeconfigFile, context, "config.json")
}
