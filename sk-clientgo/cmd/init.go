package cmd

import (
	"encoding/base64"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"os"
	"skas/sk-clientgo/internal/global"
	"skas/sk-clientgo/internal/httpClient"
	"skas/sk-common/pkg/skclient"
	"skas/sk-common/proto/v1/proto"
	"strconv"
	"strings"
)

var contextNameOverride string
var apiServerUrlOverride string
var namespaceOverride string
var command string
var noContextSwitch bool
var force bool
var authRootCaPath string
var clientId string
var clientSecret string
var authInsecureSkipVerify bool

func init() {
	InitCmd.PersistentFlags().StringVar(&contextNameOverride, "contextNameOverride", "", "Override context name. (Will be used as base for cluster and user name)")
	InitCmd.PersistentFlags().StringVar(&apiServerUrlOverride, "apiServerUrlOverride", "", "Override K8s API server URL")
	InitCmd.PersistentFlags().StringVar(&namespaceOverride, "namespaceOverride", "", "Override namespace")
	InitCmd.PersistentFlags().StringVar(&command, "command", "kubectl-sk", "The skas kubectl plugin executable")
	InitCmd.PersistentFlags().BoolVar(&noContextSwitch, "noContextSwitch", false, "Do not set default context to the newly create one.")
	InitCmd.PersistentFlags().BoolVar(&force, "force", false, "Override any already existing context")

	InitCmd.PersistentFlags().StringVar(&authRootCaPath, "authRootCaPath", "", "Path to a trusted root CA file for client connection to skas auth server")
	InitCmd.PersistentFlags().StringVar(&clientId, "clientId", "", "Client ID for authentication server")
	InitCmd.PersistentFlags().StringVar(&clientSecret, "clientSecret", "", "Client secret")
	InitCmd.PersistentFlags().BoolVar(&authInsecureSkipVerify, "authInsecureSkipVerify", false, "Skip skas auth server certificate validation")

}

var InitCmd = &cobra.Command{
	Use:   "init <configuration_url>",
	Short: "Add a new context in Kubeconfig file for skas access",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Build client config from parameters
		url := args[0]
		if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
			url = "https://" + url
		}
		skConfig := &skclient.Config{
			Url:                url,
			InsecureSkipVerify: authInsecureSkipVerify,
		}
		skConfig.ClientAuth.Id = clientId
		skConfig.ClientAuth.Secret = clientSecret
		if authRootCaPath != "" && !authInsecureSkipVerify {
			rootCABytes, err := os.ReadFile(authRootCaPath)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "ERROR: Unable to read CA file: %s\n", err.Error())
				os.Exit(15)
			}
			skConfig.RootCaData = base64.StdEncoding.EncodeToString(rootCABytes)
		}

		client, err := httpClient.NewForInit(skConfig)
		if err != nil {
			global.Log.Error(err, "error on InitHttpClient()")
			os.Exit(10)
		}
		kr := &proto.KubeconfigRequest{
			ClientAuth: client.GetClientAuth(),
		}
		kubeConfigResponse := &proto.KubeconfigResponse{}
		err = client.Do(proto.KubeconfigMeta, kr, kubeConfigResponse, nil)
		if err != nil {
			global.Log.Error(err, "error on GET kubeconfig from remote server")
			os.Exit(4)
		}
		global.Log.V(1).Info("Fetched kubeconfig from remote", "contextName", kubeConfigResponse.Context.Name)
		// ---------------------------------------------------- Override parameters
		if contextNameOverride != "" {
			kubeConfigResponse.Context.Name = contextNameOverride
		}
		if apiServerUrlOverride != "" {
			kubeConfigResponse.Cluster.ApiServerUrl = apiServerUrlOverride
		}
		if namespaceOverride != "" {
			kubeConfigResponse.Context.Namespace = namespaceOverride
		}

		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		loadingRules.ExplicitPath = kubeconfigPath // From the command line. Must take precedence
		loadingRules.WarnIfAllMissing = false
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		rawConfig, err := kubeConfig.RawConfig()
		configAccess := kubeConfig.ConfigAccess()

		contextName := kubeConfigResponse.Context.Name
		clusterName := contextName + "-cluster"
		userName := contextName + "-user"

		// In case of init blank file
		if rawConfig.Clusters == nil {
			rawConfig.Clusters = make(map[string]*api.Cluster)
		}
		if rawConfig.AuthInfos == nil {
			rawConfig.AuthInfos = make(map[string]*api.AuthInfo)
		}
		if rawConfig.Contexts == nil {
			rawConfig.Contexts = make(map[string]*api.Context)
		}
		// Test overwrite
		_, exitingContext := rawConfig.Contexts[contextName]
		if exitingContext && !force {
			_, _ = fmt.Fprintf(os.Stderr, "ERROR: context '%s' already existing in this config file (%s)\n", contextName, configAccess.GetDefaultFilename())
			os.Exit(15)
		}
		if _, ok := rawConfig.Clusters[clusterName]; ok && !force {
			_, _ = fmt.Fprintf(os.Stderr, "ERROR: cluster '%s' already existing in this config file (%s)\n", clusterName, configAccess.GetDefaultFilename())
			os.Exit(15)
		}
		if _, ok := rawConfig.AuthInfos[userName]; ok && !force {
			_, _ = fmt.Fprintf(os.Stderr, "ERROR: user '%s' already existing in this config file (%s)\n", userName, configAccess.GetDefaultFilename())
			os.Exit(15)
		}
		if exitingContext {
			fmt.Printf("Update existing context '%s' in kubeconfig file '%s'\n", contextName, configAccess.GetDefaultFilename())
		} else {
			fmt.Printf("Setup new context '%s' in kubeconfig file '%s'\n", contextName, configAccess.GetDefaultFilename())
		}

		rootCaData, err := base64.StdEncoding.DecodeString(kubeConfigResponse.Cluster.RootCaData)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "ERROR: Invalid server certificate. Refer to system administrator\n")
			os.Exit(15)
		}
		rawConfig.Clusters[clusterName] = &api.Cluster{
			Server:                   kubeConfigResponse.Cluster.ApiServerUrl,
			CertificateAuthorityData: rootCaData,
			InsecureSkipTLSVerify:    kubeConfigResponse.Cluster.InsecureSkipVerify,
		}
		rawConfig.AuthInfos[userName] = &api.AuthInfo{
			Exec: &api.ExecConfig{
				APIVersion:      "client.authentication.k8s.io/v1",
				Command:         command,
				InteractiveMode: "Always",
				Args: []string{
					"auth",
				},
				Env: []api.ExecEnvVar{
					api.ExecEnvVar{
						Name:  skclient.SK_CLIENT_URL,
						Value: skConfig.Url,
					},
					api.ExecEnvVar{
						Name:  skclient.SK_CLIENT_ROOT_CA_DATA,
						Value: skConfig.RootCaData,
					},
					api.ExecEnvVar{
						Name:  skclient.SK_CLIENT_INSECURE_SKIP_VERIFY,
						Value: strconv.FormatBool(skConfig.InsecureSkipVerify),
					},
					api.ExecEnvVar{
						Name:  skclient.SK_CLIENT_AUTH_ID,
						Value: skConfig.ClientAuth.Id,
					},
					api.ExecEnvVar{
						Name:  skclient.SK_CLIENT_AUTH_SECRET,
						Value: skConfig.ClientAuth.Secret,
					},
				},
			},
		}
		rawConfig.Contexts[contextName] = &api.Context{
			Cluster:   clusterName,
			AuthInfo:  userName,
			Namespace: kubeConfigResponse.Context.Namespace,
		}
		if rawConfig.CurrentContext == "" || !noContextSwitch {
			rawConfig.CurrentContext = contextName
		}
		err = clientcmd.ModifyConfig(configAccess, rawConfig, false)
	},
}
