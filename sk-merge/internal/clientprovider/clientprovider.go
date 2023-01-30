package clientprovider

import (
	"skas/sk-common/proto"
)

type ClientProvider interface {
	GetUserStatus(login, password string) (*proto.UserStatusResponse, *proto.Translated, error)
	IsCritical() bool
	GetName() string
	IsCredentialAuthority() bool
	IsGroupAuthority() bool
}
