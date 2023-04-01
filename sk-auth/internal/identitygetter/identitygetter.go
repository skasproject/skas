package identitygetter

import (
	"net/http"
	"skas/sk-common/pkg/misc"
	"skas/sk-common/pkg/skclient"
	"skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
)

var _ handlers.IdentityGetter = &identityGetter{}

type identityGetter struct {
	provider skclient.SkClient
}

func New(provider skclient.SkClient) handlers.IdentityGetter {
	return &identityGetter{
		provider: provider,
	}
}

func (i identityGetter) GetIdentity(request proto.IdentityRequest) (*proto.IdentityResponse, misc.HttpError) {
	// Forward the message 'as is', except our authentication
	request.ClientAuth = i.provider.GetClientAuth()
	response := &proto.IdentityResponse{}
	err := i.provider.Do(proto.IdentityMeta, &request, response, nil)
	if err != nil {
		return nil, misc.NewHttpError(err.Error(), http.StatusBadGateway)
	}
	return response, nil
}
