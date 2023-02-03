package clientprovider

import (
	"skas/sk-common/pkg/httpclient"
	"skas/sk-merge/internal/config"
)

func New(conf config.ClientProviderConfig) (ClientProvider, error) {
	cp := &clientProvider{
		ClientProviderConfig: conf,
	}
	var err error
	cp.httpClient, err = httpclient.NewHTTPClient(&conf.HttpClientConfig, config.Conf.RootCaPath, config.Conf.RootCaData)
	if err != nil {
		return nil, err
	}
	return cp, nil
}
