package handlers

import (
	"skas/sk-common/pkg/misc"
	"skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/pkg/skserver/protector"
	"skas/sk-common/proto/v1/proto"
)

func doLogin(identityGetter handlers.IdentityGetter, login, password string, protector protector.LoginProtector) (*proto.User /*authority*/, string, misc.HttpError) {
	response, err := identityGetter.GetIdentity(proto.IdentityRequest{
		Login:    login,
		Password: password,
		Detailed: false,
		// ClientAuth will be provided by called
	})
	if err != nil {
		return nil, "", err
	}
	protector.ProtectLoginResult(login, response.Status)
	if response.Status != proto.PasswordChecked {
		return nil, "", nil
	}
	return &response.User, response.Authority, nil
}
