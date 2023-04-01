package provider

import (
	"fmt"
	"github.com/go-logr/logr"
	"skas/sk-common/pkg/skclient"
	"skas/sk-merge/internal/config"
)

func New(conf config.ProviderConfig, logger logr.Logger) (Provider, error) {
	httpClient, err := skclient.New(&conf.HttpClient, config.Conf.RootCaPath, config.Conf.RootCaData)
	if err != nil {
		return nil, err
	}
	logger.Info(fmt.Sprintf("Provider '%s' configured", conf.Name), "authority", conf.CredentialAuthority, "critical", conf.Critical, "url", conf.HttpClient.Url)
	return &provider{
		ProviderConfig: conf,
		httpClient:     httpClient,
		logger:         logger,
	}, nil
}
