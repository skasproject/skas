package serverproviders

import (
	"github.com/go-logr/logr"
	"skas/sk-common/pkg/httpserver/handlers"
	"skas/sk-common/proto"
)

var _ handlers.StatusServerProvider = &statusServerProvider{}

type statusServerProvider struct {
	logger logr.Logger
}

func NewStatusServerProvider(logger logr.Logger) (handlers.StatusServerProvider, error) {
	return &statusServerProvider{
		logger: logger,
	}, nil
}

func (m statusServerProvider) GetUserStatus(request proto.UserStatusRequest) (*proto.UserStatusResponse, error) {
	//TODO implement me
	panic("implement me")
}
