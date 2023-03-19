package handlers

import (
	"fmt"
	"skas/sk-common/pkg/skclient"
	"skas/sk-common/proto/v1/proto"
)

func doLogin(loginProvider skclient.SkClient, login, password string) (*proto.User /*authority*/, string, error) {
	lr := &proto.LoginRequest{
		Login:      login,
		Password:   password,
		ClientAuth: loginProvider.GetClientAuth(),
	}
	loginResponse := &proto.LoginResponse{}
	err := loginProvider.Do(proto.LoginMeta, lr, loginResponse, nil)
	if err != nil {
		return nil, "", fmt.Errorf("error on exchange on %s: %w", proto.LoginMeta.UrlPath, err) // Do() return a documented message
	}
	if loginResponse.Success {
		return &loginResponse.User, loginResponse.Authority, nil
	} else {
		return nil, "", nil
	}
}
