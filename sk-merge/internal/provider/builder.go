package provider

import (
	"fmt"
	"github.com/go-logr/logr"
	"skas/sk-common/pkg/skclient"
	"skas/sk-merge/internal/config"
)

func New(conf config.ProviderConfig, logger logr.Logger) (Provider, error) {
	link, ok := config.Conf.ProviderInfo[conf.Name]
	if !ok {
		// This error should never occurs. Such case has been tested during config validation
		return nil, fmt.Errorf("provider '%s' has no info definition", conf.Name)
	}
	httpClient, err := skclient.New(link, config.Conf.RootCaPath, config.Conf.RootCaData)
	if err != nil {
		return nil, err
	}
	logger.Info(fmt.Sprintf("Provider '%s' configured", conf.Name), "authority", conf.CredentialAuthority, "critical", conf.Critical, "url", link.Url)
	return &provider{
		ProviderConfig: conf,
		httpClient:     httpClient,
		logger:         logger,
	}, nil
}
