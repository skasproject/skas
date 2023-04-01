package provider

import (
	"skas/sk-common/proto/v1/proto"
)

type Provider interface {
	GetUserDetail(login, password string) (*proto.UserDetail, error)
	// ChangePassword - We pass request by value as we may modify it
	ChangePassword(request proto.PasswordChangeRequest) (*proto.PasswordChangeResponse, error)
	GetName() string

	//GetIdentity(login, password string) (*proto.IdentityResponse, error)
	//Translate(response *proto.IdentityResponse) *proto.Translated
	//IsCritical() bool
	//IsCredentialAuthority() bool
	//IsGroupAuthority() bool
}
