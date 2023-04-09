package providerchain

import (
	"fmt"
	"github.com/go-logr/logr"
	"skas/sk-merge/internal/config"
	"skas/sk-merge/internal/provider"
)

func New(logger logr.Logger) (ProviderChain, error) {
	if config.Conf.Providers == nil || len(config.Conf.Providers) == 0 {
		return nil, fmt.Errorf("no client provider defined")
	}
	chain := &providerChain{
		providers: make([]provider.Provider, 0, len(config.Conf.Providers)),
		logger:    logger,
	}
	for _, clientProviderConfig := range config.Conf.Providers {
		clientProvider, err := provider.New(clientProviderConfig, logger.WithName(clientProviderConfig.Name))
		if err != nil {
			return nil, fmt.Errorf("unable to intialize provider '%s': %w", clientProviderConfig.Name, err)
		}
		chain.providers = append(chain.providers, clientProvider)
	}
	return chain, nil
}
