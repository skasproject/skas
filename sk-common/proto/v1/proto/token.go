package proto

import (
	"fmt"
	"io"
	"skas/sk-common/pkg/misc"
	"time"
)

// This API is used by sk-cli to retrieve tokens.

// ------------------------------------------------------

const TokenRequestUrlPath = "/v1/token"

var _ RequestPayload = &TokenRequest{}

type TokenRequest struct {
	ClientAuth ClientAuth `json:"clientAuth"`
	Login      string     `json:"login"`
	Password   string     `json:"password"`
}

var _ ResponsePayload = &TokenResponse{}

type TokenResponse struct {
	Success   bool          `json:"success"`
	Token     string        `json:"token"`
	User      User          `json:"user"`
	ClientTTL time.Duration `json:"clientTTL"`
	Authority string        `json:"authority"`
}

// ------------------------------------------------------

const TokenRenewUrlPath = "/v1/tokenrenew"

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

func (t *TokenRequest) String() string {
	return fmt.Sprintf("TokenRequest(login=%s)", t.Login)
}
func (t *TokenRequest) ToJson() ([]byte, error) {
	return toJson(t)
}
func (t *TokenRequest) FromJson(r io.Reader) error {
	return fromJson(r, t)
}

func (t *TokenResponse) FromJson(r io.Reader) error {
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
