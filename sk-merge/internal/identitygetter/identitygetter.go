package identitygetter

import (
	"github.com/go-logr/logr"
	"net/http"
	"skas/sk-common/pkg/misc"
	"skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-merge/internal/providerchain"
)

var _ handlers.IdentityGetter = &identityGetter{}

type identityGetter struct {
	logger logr.Logger
	chain  providerchain.ProviderChain
}

func New(chain providerchain.ProviderChain, logger logr.Logger) handlers.IdentityGetter {
	return &identityGetter{
		logger: logger,
		chain:  chain,
	}
}

func (p identityGetter) GetIdentity(request proto.IdentityRequest) (*proto.IdentityResponse, misc.HttpError) {
	identityResponse, err := p.chain.GetIdentity(request.Login, request.Password, request.Detailed)
	if err != nil {
		return nil, misc.NewHttpError(err.Error(), http.StatusBadGateway)
	}
	return identityResponse, nil
}
