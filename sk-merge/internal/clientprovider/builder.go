package clientprovider

import (
	"skas/sk-common/pkg/skhttp"
	"skas/sk-merge/internal/config"
)

func New(conf config.ClientProviderConfig) (ClientProvider, error) {
	cp := &clientProvider{
		ClientProviderConfig: conf,
	}
	var err error
	skclient, err := skhttp.New(&conf.HttpClient, config.Conf.RootCaPath, config.Conf.RootCaData)
	if err != nil {
		return nil, err
	}

	cp.httpClient = skclient.GetHttpClient()
	return cp, nil
}
