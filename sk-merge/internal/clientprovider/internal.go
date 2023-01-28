package clientprovider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"skas/sk-common/proto"
	"skas/sk-merge/internal/config"
)

var _ ClientProvider = &clientProvider{}

type clientProvider struct {
	config.ClientProviderConfig
	httpClient *http.Client
}

func (c clientProvider) IsCritical() bool {
	return *c.Critical
}

func (c clientProvider) GetName() string {
	return c.Name
}

func (c clientProvider) GetUserStatus(login, password string) (*proto.UserStatusResponse, error) {
	body, err := json.Marshal(proto.UserStatusRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to marshal login UserStatusRequest (login:'%s'): %w", login, err)
	}
	request, err := http.NewRequest("GET", c.HttpClientConfig.Url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("unable to build request")
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error on http connection: %w", err)
	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status code: %d (%s)", response.StatusCode, response.Status)
	}
	userStatusResponse := &proto.UserStatusResponse{}
	err = json.NewDecoder(response.Body).Decode(userStatusResponse)
	if err != nil {
		return nil, fmt.Errorf("unable to parse server response: %w", err)
	}
	return userStatusResponse, nil
}
