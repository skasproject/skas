package proto

import "time"

// This API is used by sk-cli to retrieve tokens.

// ------------------------------------------------------

var TokenUrlPath = "/v1/token"

type TokenRequest struct {
	ClientAuth ClientAuth `json:"clientAuth"`
	Login      string     `json:"login"`
	Password   string     `json:"password"`
}

type TokenResponse struct {
	Token string `json:"token"`
	User
	ClientTTL time.Duration `json:"clientTTL"`
}

// ------------------------------------------------------

var TokenRenewUrlPath = "/v1/tokenrenew"

type TokenRenewRequest struct {
	ClientAuth ClientAuth `json:"clientAuth"`
	Token      string     `json:"token"`
}

type TokenRenewResponse struct {
	Token string `json:"token"`
	Valid bool   `json:"valid"`
}
