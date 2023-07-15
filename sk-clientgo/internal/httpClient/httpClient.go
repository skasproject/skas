package httpClient

import (
	"fmt"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"skas/sk-clientgo/internal/global"
	"skas/sk-common/pkg/skclient"
	"strconv"
)

func NewForInit(config *skclient.Config) (skclient.SkClient, error) {
	return skclient.New(config, "", "")
}

func New() (skclient.SkClient, error) {
	skConfig := &skclient.Config{}
	// First, lookup in environment
	skConfig.Url = os.Getenv(skclient.SK_CLIENT_URL)
	var err error
	if skConfig.Url != "" {
		// Must fetch all remaining variables from env
		skConfig.InsecureSkipVerify, err = strconv.ParseBool(os.Getenv(skclient.SK_CLIENT_INSECURE_SKIP_VERIFY))
		if err != nil {
			return nil, fmt.Errorf("error in kubconfig: Unable to parse '%s' as boolean for '%s'. Try kubectl sk init --force https://..... ", os.Getenv(skclient.SK_CLIENT_INSECURE_SKIP_VERIFY), skclient.SK_CLIENT_INSECURE_SKIP_VERIFY)
		}
		skConfig.RootCaData = os.Getenv(skclient.SK_CLIENT_ROOT_CA_DATA)
		skConfig.ClientAuth.Id = os.Getenv(skclient.SK_CLIENT_AUTH_ID)
		skConfig.ClientAuth.Secret = os.Getenv(skclient.SK_CLIENT_AUTH_SECRET)
	} else {
		// Must dig directly in the kubernetes config file
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		loadingRules.ExplicitPath = global.KubeconfigPath
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		//kubeconfigFile = kubeConfig.ConfigAccess().GetDefaultFilename()
		rawConfig, err := kubeConfig.RawConfig()
		if err != nil {
			return nil, fmt.Errorf("error in kubconfig. Try kubectl sk init --force https://..... ")
		}
		kubeContext := rawConfig.CurrentContext
		if kubeContext == "" {
			kubeContext = "default"
		}
		context := rawConfig.Contexts[kubeContext]
		if context == nil {
			return nil, fmt.Errorf("error in kubconfig: Unable locate context '%s'. Try kubectl sk init --force https://..... ", kubeContext)
		}
		user := rawConfig.AuthInfos[context.AuthInfo]
		if user == nil {
			return nil, fmt.Errorf("error in kubconfig: Unable locate user '%s'. Try kubectl sk init --force https://..... ", context.AuthInfo)
		}
		varFromName := make(map[string]string)
		if user.Exec == nil || user.Exec.Env == nil {
			return nil, fmt.Errorf("error in kubeconfig. Does not look like a SKAS configuration. Are you logged as a SKAS user? ")
		}
		for _, execEnvVar := range user.Exec.Env {
			varFromName[execEnvVar.Name] = execEnvVar.Value
		}
		skConfig.Url = varFromName[skclient.SK_CLIENT_URL]
		skConfig.InsecureSkipVerify, err = strconv.ParseBool(varFromName[skclient.SK_CLIENT_INSECURE_SKIP_VERIFY])
		if err != nil {
			return nil, fmt.Errorf("error in kubconfig: Unable to parse '%s' as boolean for '%s'. Try kubectl sk init --force https://...... ", varFromName[skclient.SK_CLIENT_INSECURE_SKIP_VERIFY], skclient.SK_CLIENT_INSECURE_SKIP_VERIFY)
		}
		skConfig.RootCaData = varFromName[skclient.SK_CLIENT_ROOT_CA_DATA]
		skConfig.ClientAuth.Id = varFromName[skclient.SK_CLIENT_AUTH_ID]
		skConfig.ClientAuth.Secret = varFromName[skclient.SK_CLIENT_AUTH_SECRET]
	}
	return skclient.New(skConfig, "", "")
}
