package clientprovider

import (
	"skas/sk-common/pkg/skclient"
	"skas/sk-merge/internal/config"
)

func New(conf config.ClientProviderConfig) (ClientProvider, error) {
	httpClient, err := skclient.New(&conf.HttpClient, config.Conf.RootCaPath, config.Conf.RootCaData)
	if err != nil {
		return nil, err
	}
	return &clientProvider{
		ClientProviderConfig: conf,
		httpClient:           httpClient,
	}, nil

}
