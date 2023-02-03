package clientprovider

import (
	"fmt"
	"skas/sk-common/pkg/skhttp"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-merge/internal/config"
)

var _ ClientProvider = &clientProvider{}

type clientProvider struct {
	config.ClientProviderConfig
	httpClient skhttp.Client
}

func (c clientProvider) IsGroupAuthority() bool {
	return *c.GroupAuthority
}

func (c clientProvider) IsCredentialAuthority() bool {
	return *c.CredentialAuthority
}

func (c clientProvider) IsCritical() bool {
	return *c.Critical
}

func (c clientProvider) GetName() string {
	return c.Name
}

func (c clientProvider) GetUserStatus(login, password string) (*proto.UserStatusResponse, *proto.Translated, error) {
	usr := &proto.UserStatusRequest{
		Login:      login,
		Password:   password,
		ClientAuth: c.httpClient.GetClientAuth(),
	}
	userStatusResponse := &proto.UserStatusResponse{}
	err := c.httpClient.Do(proto.UserStatusUrlPath, usr, userStatusResponse)
	if err != nil {
		return nil, nil, err // Do() return a documented message
	}
	translated := &proto.Translated{
		Uid:    userStatusResponse.Uid + c.UidOffset,
		Groups: make([]string, len(userStatusResponse.Groups)),
	}
	for idx := range userStatusResponse.Groups {
		translated.Groups[idx] = fmt.Sprintf(c.GroupPattern, userStatusResponse.Groups[idx])
	}
	return userStatusResponse, translated, nil
}
