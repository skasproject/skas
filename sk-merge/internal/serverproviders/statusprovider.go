package serverproviders

import (
	"github.com/go-logr/logr"
	"skas/sk-common/pkg/httpserver/handlers"
	"skas/sk-common/proto"
	"skas/sk-merge/internal/clientproviderchain"
)

var _ handlers.StatusServerProvider = &statusServerProvider{}

type statusServerProvider struct {
	logger logr.Logger
	chain  clientproviderchain.ClientProviderChain
}

func NewStatusServerProvider(chain clientproviderchain.ClientProviderChain, logger logr.Logger) (handlers.StatusServerProvider, error) {
	return &statusServerProvider{
		logger: logger,
		chain:  chain,
	}, nil
}

func (p statusServerProvider) GetUserStatus(request proto.UserStatusRequest) (*proto.UserStatusResponse, error) {
	items, err := p.chain.Scan(request.Login, request.Password)
	if err != nil {
		return nil, err
	}
	userStatusResponse, _ := clientproviderchain.Merge(request.Login, items)
	return userStatusResponse, nil
}
