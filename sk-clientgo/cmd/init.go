package cmd

import (
	"encoding/base64"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"os"
	"skas/sk-clientgo/httpClient"
	"skas/sk-clientgo/internal/global"
	"skas/sk-common/proto/v1/proto"
	"strconv"
)

var contextNameOverride string
var apiServerUrlOverride string
var authServerUrlOverride string
var namespaceOverride string
var command string
var noContextSwitch bool
var force bool

func init() {
	InitCmd.PersistentFlags().StringVar(&contextNameOverride, "contextNameOverride", "", "Override context name. (Will be used as base for cluster and user name)")
	InitCmd.PersistentFlags().StringVar(&apiServerUrlOverride, "apiServerUrlOverride", "", "Override K8s API server URL")
	InitCmd.PersistentFlags().StringVar(&authServerUrlOverride, "authServerUrlOverride", "", "Override skas auth server URL")
	InitCmd.PersistentFlags().StringVar(&namespaceOverride, "namespaceOverride", "", "Override namespace")
	InitCmd.PersistentFlags().StringVar(&command, "command", "kubectl-sk", "The skas kubectl plugin executable")
	InitCmd.PersistentFlags().BoolVar(&noContextSwitch, "noContextSwitch", false, "Do not set default context to the newly create one.")
	InitCmd.PersistentFlags().BoolVar(&force, "force", false, "Override any already existing context")
	httpClient.AddFlags(InitCmd)
}

var InitCmd = &cobra.Command{
	Use:   "init <configuration_url>",
	Short: "Add a new context in Kubeconfig file for skas access",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := httpClient.NewForInit(args[0])
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
		if authServerUrlOverride != "" {
			kubeConfigResponse.User.AuthServerUrl = authServerUrlOverride
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
					"--authServerUrl=" + kubeConfigResponse.User.AuthServerUrl,
					"--authInsecureSkipVerify=" + strconv.FormatBool(kubeConfigResponse.User.InsecureSkipVerify),
					"--clientId=" + kubeConfigResponse.User.ClientAuth.Id,
					"--clientSecret=" + kubeConfigResponse.User.ClientAuth.Secret,
					"--reset",
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
