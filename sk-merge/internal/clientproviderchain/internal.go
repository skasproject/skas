package clientproviderchain

import (
	"fmt"
	"github.com/go-logr/logr"
	"skas/sk-merge/internal/clientprovider"
)

var _ ClientProviderChain = &clientProviderChain{}

type clientProviderChain struct {
	providers []clientprovider.ClientProvider
	logger    logr.Logger
}

func (c clientProviderChain) Scan(login, password string) ([]ScanItem, error) {
	result := make([]ScanItem, 0, len(c.providers))
	for idx, provider := range c.providers {
		userStatusResponse, err := provider.GetUserStatus(login, password)
		if err != nil {
			if provider.IsCritical() {
				c.logger.Error(err, "Provider error. aborting", "provider", provider.GetName())
				return nil, fmt.Errorf("error on provider %s: %w", provider.GetName(), err)
			} else {
				c.logger.Error(err, "Provider error. Skipping", "provider", provider.GetName())
			}
		}
		item := ScanItem{
			UserStatusResponse: userStatusResponse,
			Provider:           &c.providers[idx], // NOT &provider
		}
		result = append(result, item)
	}
	return result, nil
}

func (c clientProviderChain) GetLength() int {
	return len(c.providers)
}
