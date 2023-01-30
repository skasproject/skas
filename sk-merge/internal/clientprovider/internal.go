package clientprovider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"skas/sk-common/proto"
	"skas/sk-merge/internal/config"
)

var _ ClientProvider = &clientProvider{}

type clientProvider struct {
	config.ClientProviderConfig
	httpClient *http.Client
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
	body, err := json.Marshal(proto.UserStatusRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("unable to marshal login UserStatusRequest (login:'%s'): %w", login, err)
	}
	u, err := url.JoinPath(c.HttpClientConfig.Url, proto.UserStatusUrlPath)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to join %s to %s: %w", proto.UserStatusUrlPath, c.HttpClientConfig.Url, err)
	}
	request, err := http.NewRequest("GET", u, bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to build request")
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, nil, fmt.Errorf("error on http connection: %w", err)
	}
	if response.StatusCode != 200 {
		return nil, nil, fmt.Errorf("invalid status code: %d (%s)", response.StatusCode, response.Status)
	}
	userStatusResponse := &proto.UserStatusResponse{}
	err = json.NewDecoder(response.Body).Decode(userStatusResponse)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse server response: %w", err)
	}
	translated := &proto.Translated{
		Uid:    userStatusResponse.Uid + c.UidOffset,
		Groups: make([]string, len(userStatusResponse.Groups)),
	}
	for idx, _ := range userStatusResponse.Groups {
		translated.Groups[idx] = fmt.Sprintf(c.GroupPattern, userStatusResponse.Groups[idx])
	}
	return userStatusResponse, translated, nil
}
