package clientprovider

import (
	"skas/sk-common/proto/v1/proto"
)

type ClientProvider interface {
	GetUserStatus(login, password string) (*proto.UserIdentityResponse, *proto.Translated, error)
	IsCritical() bool
	GetName() string
	IsCredentialAuthority() bool
	IsGroupAuthority() bool
}
