package clientproviderchain

import (
	"fmt"
	"github.com/go-logr/logr"
	"skas/sk-common/proto/v1/proto"
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
		userIdentityResponse, translated, err := provider.GetUserIdentity(login, password)
		if err != nil {
			if provider.IsCritical() {
				c.logger.Error(err, "Provider error. aborting", "provider", provider.GetName())
				return nil, fmt.Errorf("error on provider '%s': %w", provider.GetName(), err)
			} else {
				c.logger.Error(err, "Provider error. Skipping", "provider", provider.GetName())
				// Build a fake ScanItem
				item = ScanItem{
					Provider: &c.providers[idx], // NOT &provider
					UserIdentityResponse: &proto.UserIdentityResponse{
						User: proto.User{
							Login:       login,
							CommonNames: []string{},
							Emails:      []string{},
							Groups:      []string{},
						},
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
				Provider:             &c.providers[idx], // NOT &provider
				UserIdentityResponse: userIdentityResponse,
				Translated:           translated,
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

func (c clientProviderChain) lookupProvider(name string) clientprovider.ClientProvider {
	for idx, prvd := range c.providers {
		if prvd.GetName() == name {
			return c.providers[idx]
		}
	}
	return nil
}

func (c clientProviderChain) ChangePassword(request *proto.PasswordChangeRequest) (*proto.PasswordChangeResponse, error) {
	prvd := c.lookupProvider(request.Provider)
	if prvd == nil {
		return &proto.PasswordChangeResponse{
			Login:  request.Login,
			Status: proto.UnknownProvider,
		}, nil
	}
	return prvd.ChangePassword(request)
}
