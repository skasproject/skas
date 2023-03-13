package skhttp

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"skas/sk-common/proto/v1/proto"
)

type HttpAuth struct {
	Login    string
	Password string
	Token    string
}

type Client interface {
	Do(meta *proto.RequestMeta, request proto.RequestPayload, response proto.ResponsePayload, httpAuth *HttpAuth) error
	GetClientAuth() proto.ClientAuth
	GetConfig() *Config
}

var _ Client = &client{}

type client struct {
	Config
	httpClient *http.Client
}

func (c client) GetConfig() *Config {
	return &c.Config
}

func (c client) GetClientAuth() proto.ClientAuth {
	return proto.ClientAuth{
		Id:     c.ClientAuth.Id,
		Secret: c.ClientAuth.Secret,
	}
}

type UnauthorizedError struct{}

func (e *UnauthorizedError) Error() string {
	return "Unauthorized"
}

type NotFoundError struct {
	url string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("Resource '%s' not found", e.url)
}

func (c client) Do(meta *proto.RequestMeta, request proto.RequestPayload, response proto.ResponsePayload, httpAuth *HttpAuth) error {
	body, err := request.ToJson()
	if err != nil {
		return fmt.Errorf("unable to marshal %s: %w", request.String(), err)
	}
	u, err := url.JoinPath(c.Url, meta.UrlPath)
	if err != nil {
		return fmt.Errorf("unable to join %s to %s: %w", meta.UrlPath, c.Url, err)
	}
	req, err := http.NewRequest(meta.Method, u, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("unable to build request")
	}
	if httpAuth != nil {
		if httpAuth.Login != "" {
			req.SetBasicAuth(httpAuth.Login, httpAuth.Password)
		}
		if httpAuth.Token != "" {
			req.Header.Set("Authorization", "Bearer "+httpAuth.Token)
		}
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error on http connection: %w", err)
	}
	if resp.StatusCode == 401 {
		// This is not a system error, but a user's one. So this special handling
		return &UnauthorizedError{}
	}
	if resp.StatusCode == 404 {
		// Some caller may need to handle this specifically
		return &NotFoundError{
			url: u,
		}
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid status code: %d (%s)", resp.StatusCode, resp.Status)
	}
	err = response.FromJson(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to unmarshal response for request %s: %w", request.String(), err)
	}
	return nil
}
