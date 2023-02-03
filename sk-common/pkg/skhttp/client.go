package skhttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"skas/sk-common/proto/v1/proto"
)

type Client interface {
	Do(urlPath string, request proto.Payload, response proto.Payload) error
	GetHttpClient() *http.Client // TODO: Remove this as only for compatibility before refactoring
	GetClientAuth() proto.ClientAuth
}

var _ Client = &client{}

type client struct {
	Config
	httpClient *http.Client
}

func (c client) GetClientAuth() proto.ClientAuth {
	return proto.ClientAuth{
		Id:     c.ClientAuth.Id,
		Secret: c.ClientAuth.Secret,
	}
}

// TODO: Remove this as only for compatibility before refactoring

func (c client) GetHttpClient() *http.Client {
	return c.httpClient
}

func (c client) Do(urlPath string, request proto.Payload, response proto.Payload) error {

	body, err := request.ToJson()
	if err != nil {
		return fmt.Errorf("unable to marshal %s: %w", request, err)
	}
	u, err := url.JoinPath(c.Url, urlPath)
	if err != nil {
		return fmt.Errorf("unable to join %s to %s: %w", urlPath, c.Url, err)
	}
	req, err := http.NewRequest("GET", u, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("unable to build request")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error on http connection: %w", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid status code: %d (%s)", resp.StatusCode, resp.Status)
	}
	err = response.FromJson(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to unmarshal response for request %s: %w", request, err)
	}
	return nil
}

func (c client) Do2(urlPath string, request proto.Payload, response proto.Payload) (io.ReadCloser, error) {

	body, err := request.ToJson()
	if err != nil {
		return nil, fmt.Errorf("unable to marshal %s: %w", request, err)
	}
	u, err := url.JoinPath(c.Url, urlPath)
	if err != nil {
		return nil, fmt.Errorf("unable to join %s to %s: %w", urlPath, c.Url, err)
	}
	req, err := http.NewRequest("GET", u, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("unable to build request")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error on http connection: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status code: %d (%s)", resp.StatusCode, resp.Status)
	}
	return resp.Body, nil
}
