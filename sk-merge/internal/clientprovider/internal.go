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

func (c clientProvider) GetUserIdentity(login, password string) (*proto.UserIdentityResponse, *proto.Translated, error) {
	usr := &proto.UserIdentityRequest{
		Login:      login,
		Password:   password,
		ClientAuth: c.httpClient.GetClientAuth(),
	}
	userIdentityResponse := &proto.UserIdentityResponse{}
	err := c.httpClient.Do(proto.UserIdentityMeta, usr, userIdentityResponse, nil)
	if err != nil {
		return nil, nil, err // Do() return a documented message
	}
	translated := &proto.Translated{
		Uid:    userIdentityResponse.Uid + c.UidOffset,
		Groups: make([]string, len(userIdentityResponse.Groups)),
	}
	for idx := range userIdentityResponse.Groups {
		translated.Groups[idx] = fmt.Sprintf(c.GroupPattern, userIdentityResponse.Groups[idx])
	}
	return userIdentityResponse, translated, nil
}

func (c clientProvider) ChangePassword(request *proto.PasswordChangeRequest) (*proto.PasswordChangeResponse, error) {
	// Forward the message 'as is', except out authentication
	request.ClientAuth = c.httpClient.GetClientAuth()
	passwordChangeResponse := &proto.PasswordChangeResponse{}
	err := c.httpClient.Do(proto.PasswordChangeMeta, request, passwordChangeResponse, nil)
	if err != nil {
		if _, ok := err.(*skhttp.NotFoundError); ok {
			passwordChangeResponse.Status = proto.Unsupported
			passwordChangeResponse.Login = request.Login
			return passwordChangeResponse, nil
		} else {
			return nil, err // Do() return a documented message
		}
	}
	return passwordChangeResponse, nil
}
