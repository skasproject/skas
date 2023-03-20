package proto

import (
	"fmt"
	"io"
	"skas/sk-common/pkg/misc"
	"time"
)

// This API is used by sk-cli to retrieve tokens.

// ------------------------------------------------------

var TokenCreateMeta = &RequestMeta{
	Name:    "tokenCreate",
	Method:  "POST",
	UrlPath: "/v1/tokenCreate",
}

var _ RequestPayload = &TokenCreateRequest{}

type TokenCreateRequest struct {
	ClientAuth ClientAuth `json:"clientAuth"`
	Login      string     `json:"login"`
	Password   string     `json:"password"`
}

var _ ResponsePayload = &TokenCreateResponse{}

type TokenCreateResponse struct {
	Success   bool          `json:"success"`
	Token     string        `json:"token"`
	User      User          `json:"user"`
	ClientTTL time.Duration `json:"clientTTL"`
	Authority string        `json:"authority"`
}

// ------------------------------------------------------

var TokenRenewMeta = &RequestMeta{
	Name:    "tokenRenew",
	Method:  "POST",
	UrlPath: "/v1/tokenRenew",
}

var _ RequestPayload = &TokenRenewRequest{}

type TokenRenewRequest struct {
	ClientAuth ClientAuth `json:"clientAuth"`
	Token      string     `json:"token"`
}

var _ ResponsePayload = &TokenRenewResponse{}

type TokenRenewResponse struct {
	Token string `json:"token"`
	Valid bool   `json:"valid"`
}

// ------------------------------------------------------

func (t *TokenCreateRequest) String() string {
	return fmt.Sprintf("TokenRequest(login=%s)", t.Login)
}
func (t *TokenCreateRequest) ToJson() ([]byte, error) {
	return toJson(t)
}
func (t *TokenCreateRequest) FromJson(r io.Reader) error {
	return fromJson(r, t)
}

func (t *TokenCreateResponse) FromJson(r io.Reader) error {
	return fromJson(r, t)
}

func (t *TokenRenewRequest) String() string {
	return fmt.Sprintf("TokenRenewRequest(token=%s)", misc.ShortenString(t.Token))
}
func (t *TokenRenewRequest) ToJson() ([]byte, error) {
	return toJson(t)
}
func (t *TokenRenewRequest) FromJson(r io.Reader) error {
	return fromJson(r, t)
}

func (t *TokenRenewResponse) FromJson(r io.Reader) error {
	return fromJson(r, t)
}
