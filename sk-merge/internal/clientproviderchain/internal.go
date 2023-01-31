package clientproviderchain

import (
	"fmt"
	"github.com/go-logr/logr"
	"skas/sk-common/proto"
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
		var item ScanItem
		userStatusResponse, translated, err := provider.GetUserStatus(login, password)
		if err != nil {
			if provider.IsCritical() {
				c.logger.Error(err, "Provider error. aborting", "provider", provider.GetName())
				return nil, fmt.Errorf("error on provider '%s': %w", provider.GetName(), err)
			} else {
				c.logger.Error(err, "Provider error. Skipping", "provider", provider.GetName())
				// Build a fake ScanItem
				item = ScanItem{
					Provider: &c.providers[idx], // NOT &provider
					UserStatusResponse: &proto.UserStatusResponse{
						Login:      login,
						UserStatus: proto.Undefined,
					},
					Translated: &proto.Translated{
						Uid:    0,
						Groups: []string{},
					},
				}
			}
		} else {
			item = ScanItem{
				Provider:           &c.providers[idx], // NOT &provider
				UserStatusResponse: userStatusResponse,
				Translated:         translated,
			}

		}
		result = append(result, item)
	}
	return result, nil
}

func (c clientProviderChain) GetLength() int {
	return len(c.providers)
}

func isUserFound(st proto.UserStatus) bool {
	return st == proto.PasswordChecked || st == proto.PasswordUnchecked || st == proto.PasswordFail
}
