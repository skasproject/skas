package provider

import (
	"fmt"
	"github.com/go-logr/logr"
	"skas/sk-common/pkg/skclient"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-merge/internal/config"
)

var _ Provider = &provider{}

type provider struct {
	config.ProviderConfig
	httpClient skclient.SkClient
	logger     logr.Logger
}

func (p *provider) GetName() string {
	return p.Name
}

func (p *provider) GetUserDetail(login, password string) (*proto.UserDetail, error) {
	usr := &proto.IdentityRequest{
		Login:      login,
		Password:   password,
		ClientAuth: p.httpClient.GetClientAuth(),
	}
	identityResponse := &proto.IdentityResponse{}
	err := p.httpClient.Do(proto.IdentityMeta, usr, identityResponse, nil)
	if err != nil {
		if *p.Critical {
			p.logger.Error(err, "Provider error. aborting")
			return nil, fmt.Errorf("error on provider '%s': %w", p.Name, err)
		} else {
			p.logger.Error(err, "Provider error. Skipping")
			return &proto.UserDetail{
				User:         proto.InitUser(login),
				Status:       proto.Undefined,
				ProviderSpec: p.getSpec(),
				Translated: proto.Translated{
					Uid:    0,
					Groups: []string{},
				},
			}, nil
		}
	}
	userDetail := &proto.UserDetail{
		User:         identityResponse.User,
		Status:       identityResponse.Status,
		ProviderSpec: p.getSpec(),
		Translated: proto.Translated{
			Uid:    identityResponse.Uid + p.UidOffset,
			Groups: make([]string, len(identityResponse.Groups)),
		},
	}
	for idx := range identityResponse.Groups {
		userDetail.Translated.Groups[idx] = fmt.Sprintf(p.GroupPattern, identityResponse.Groups[idx])
	}
	return userDetail, nil
}

func (p *provider) getSpec() proto.ProviderSpec {
	return proto.ProviderSpec{
		Name:                p.Name,
		CredentialAuthority: *p.CredentialAuthority,
		GroupAuthority:      *p.GroupAuthority,
	}
}

func (p *provider) ChangePassword(request proto.PasswordChangeRequest) (*proto.PasswordChangeResponse, error) {
	// Forward the message 'as is', except our authentication
	request.ClientAuth = p.httpClient.GetClientAuth()
	passwordChangeResponse := &proto.PasswordChangeResponse{}
	err := p.httpClient.Do(proto.PasswordChangeMeta, &request, passwordChangeResponse, nil)
	if err != nil {
		if _, ok := err.(*skclient.NotFoundError); ok {
			passwordChangeResponse.Status = proto.Unsupported
			passwordChangeResponse.Login = request.Login
			return passwordChangeResponse, nil
		} else {
			return nil, err // Do() return a documented message
		}
	}
	return passwordChangeResponse, nil
}
