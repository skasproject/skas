package clientproviderchain

import (
	"fmt"
	"github.com/go-logr/logr"
	"skas/sk-merge/internal/clientprovider"
	"skas/sk-merge/internal/config"
)

func New(logger logr.Logger) (ClientProviderChain, error) {
	chain := &clientProviderChain{
		providers: make([]clientprovider.ClientProvider, 0, len(config.Conf.Providers)),
		logger:    logger,
	}
	for _, clientProviderConfig := range config.Conf.Providers {
		if *clientProviderConfig.Enabled {
			clientProvider, err := clientprovider.New(clientProviderConfig)
			if err != nil {
				return nil, fmt.Errorf("unable to intialize provider '%s': %w", clientProviderConfig.Name, err)
			}
			chain.providers = append(chain.providers, clientProvider)
		}
	}
	return chain, nil
}
