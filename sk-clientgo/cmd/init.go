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
	InitCmd.PersistentFlags().StringVar(&command, "command", "kubectl-skas", "The skas kubectl plugin executable")
	InitCmd.PersistentFlags().BoolVar(&noContextSwitch, "noContextSwitch", false, "Do not set default context to the newly create one.")
	InitCmd.PersistentFlags().BoolVar(&force, "force", false, "Override any already existing context")
	httpClient.AddFlags(InitCmd)
}

var InitCmd = &cobra.Command{
	Use:   "init <configuration_url>",
	Short: "Add a new context in Kubeconfig file for skas access",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// WARNING: There is a special processing for this command in the root.go file.
		client, err := httpClient.New(true)
		if err != nil {
			global.Log.Error(err, "error on InitHttpClient()")
			os.Exit(10)
		}
		if len(args) >= 1 && client.GetConfig().Url != "" {
			_, _ = fmt.Fprintf(os.Stderr, "--authServerUrl should not be set on the 'init' command\n")
			os.Exit(2)
		}
		kr := &proto.KubeconfigRequest{
			ClientAuth: client.GetClientAuth(),
		}
		kubeConfigResponse := &proto.KubeconfigResponse{}
		err = client.Do(proto.KubeconfigMeta, kr, kubeConfigResponse)
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
			_, _ = fmt.Fprintf(os.Stderr, "ERROR: context '%s' already existing in this config file (%s)\n", contextName, kubeConfig.ConfigAccess().GetDefaultFilename())
			os.Exit(15)
		}
		if _, ok := rawConfig.Clusters[clusterName]; ok && !force {
			_, _ = fmt.Fprintf(os.Stderr, "ERROR: cluster '%s' already existing in this config file (%s)\n", clusterName, kubeConfig.ConfigAccess().GetDefaultFilename())
			os.Exit(15)
		}
		if _, ok := rawConfig.AuthInfos[userName]; ok && !force {
			_, _ = fmt.Fprintf(os.Stderr, "ERROR: user '%s' already existing in this config file (%s)\n", userName, kubeConfig.ConfigAccess().GetDefaultFilename())
			os.Exit(15)
		}
		if exitingContext {
			global.Log.V(0).Info("Update existing context", "context", contextName, "kubeconfig", kubeConfig.ConfigAccess().GetDefaultFilename())
		} else {
			global.Log.V(0).Info("Setup new context", "context", contextName, "kubeconfig", kubeConfig.ConfigAccess().GetDefaultFilename())
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

//
//func setupConfig(baseConfig *apiv1.Config, added *proto.KubeconfigResponse, setCurrentContext bool, force bool) error {
//	userName := added.ContextName + "-user"
//	clusterName := added.ContextName + "-cluster"
//
//	// -------------------------------------------------------Handle cluster
//	namedCluster := apiv1.NamedCluster{
//		Name: clusterName,
//		Cluster: apiv1.Cluster{
//			Server:                   added.Cluster.ApiServerUrl,
//			CertificateAuthorityData: []byte(added.Cluster.RootCaData),
//			InsecureSkipTLSVerify:    added.Cluster.InsecureSkipVerify,
//		},
//	}
//	clusterSet := false
//	for idx, ncl := range baseConfig.Clusters {
//		if ncl.Name == clusterName {
//			if force {
//				baseConfig.Clusters[idx] = namedCluster
//				clusterSet = true
//			} else {
//				return fmt.Errorf("cluster '%s' allready existing in this config file", clusterName)
//			}
//		}
//	}
//	if !clusterSet {
//		baseConfig.Clusters = append(baseConfig.Clusters, namedCluster)
//	}
//	// -------------------------------------------------------------- Handle user
//	namedUser := apiv1.NamedAuthInfo{
//		Name: userName,
//		AuthInfo: apiv1.AuthInfo{
//			Exec: &apiv1.ExecConfig{
//				Command:    command,
//				APIVersion: "client.authentication.k8s.io/v1beta1",
//				Args: []string{
//					"auth",
//					"--authServerUrl=" + added.User.AuthServerUrl,
//					"--authInsecureSkipVerify=" + strconv.FormatBool(added.User.InsecureSkipVerify),
//					"--reset",
//				},
//			},
//		},
//	}
//	userSet := false
//	for idx, nusr := range baseConfig.AuthInfos {
//		if nusr.Name == userName {
//			if force {
//				baseConfig.AuthInfos[idx] = namedUser
//				userSet = true
//			} else {
//				return fmt.Errorf("user '%s' allready existing in this config file", userName)
//			}
//		}
//	}
//	if !userSet {
//		baseConfig.AuthInfos = append(baseConfig.AuthInfos, namedUser)
//	}
//	// --------------------------------------------------------------- Handle context
//	namedContext := apiv1.NamedContext{
//		Name: added.ContextName,
//		Context: apiv1.Context{
//			Cluster:   clusterName,
//			AuthInfo:  userName,
//			Namespace: added.Namespace,
//		},
//	}
//	ctxSet := false
//	for idx, nctx := range baseConfig.Contexts {
//		if nctx.Name == added.ContextName {
//			if force {
//				baseConfig.Contexts[idx] = namedContext
//				ctxSet = true
//			} else {
//				return fmt.Errorf("context '%s' allready existing in this config file", added.ContextName)
//			}
//		}
//	}
//	if !ctxSet {
//		baseConfig.Contexts = append(baseConfig.Contexts, namedContext)
//	}
//	// -------------------------------------------------------------
//	if setCurrentContext {
//		baseConfig.CurrentContext = added.ContextName
//	}
//	return nil
//}
