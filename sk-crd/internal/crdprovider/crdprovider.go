package crdprovider

import (
	"skas/sk-common/pkg/httpserver/handlers"
	"skas/sk-common/proto"
)

var _ handlers.StatusProvider = &crdProvider{}

type crdProvider struct {
}

func New() handlers.StatusProvider {
	return &crdProvider{}
}

func (s crdProvider) GetUserStatus(request proto.UserStatusRequest) (*proto.UserStatusResponse, error) {
	responsePayload := &proto.UserStatusResponse{
		UserStatus: proto.NotFound,
	}
	return responsePayload, nil
}
