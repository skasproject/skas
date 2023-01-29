package clientprovider

import (
	"skas/sk-common/proto"
)

type ClientProvider interface {
	GetUserStatus(login, password string) (*proto.UserStatusResponse, error)
	IsCritical() bool
	GetName() string
	IsAuthority() bool
}
