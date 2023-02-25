package serverproviders

import (
	"github.com/go-logr/logr"
	"skas/sk-common/pkg/httpserver/handlers"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-merge/internal/clientproviderchain"
)

var _ handlers.IdentityServerProvider = &identityServerProvider{}

type identityServerProvider struct {
	logger logr.Logger
	chain  clientproviderchain.ClientProviderChain
}

func NewIdentityServerProvider(chain clientproviderchain.ClientProviderChain, logger logr.Logger) (handlers.IdentityServerProvider, error) {
	return &identityServerProvider{
		logger: logger,
		chain:  chain,
	}, nil
}

func (p identityServerProvider) GetUserIdentity(request proto.UserIdentityRequest) (*proto.UserIdentityResponse, error) {
	items, err := p.chain.Scan(request.Login, request.Password)
	if err != nil {
		return nil, err
	}
	userStatusResponse, _ := clientproviderchain.Merge(request.Login, items)
	return userStatusResponse, nil
}
