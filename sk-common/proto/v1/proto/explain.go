package proto

import (
	"fmt"
	"io"
)

// ----------------------------------- UserDescribe interface

// This is issued by sk-cli to sk-auth, which validate the token.
// Then, it is forwarded to sk-merge, without Token but with ClientAuth

var _ RequestPayload = &UserExplainRequest{}

var UserExplainMeta = &RequestMeta{
	Method:  "GET",
	UrlPath: "/v1/userExplain",
}

type UserExplainRequest struct {
	ClientAuth ClientAuth `json:"clientAuth"`
	Token      string     `json:"token"`
	Login      string     `json:"login"`
	Password   string     `json:"password"` // Optional
}

var _ ResponsePayload = &UserExplainResponse{}

type UserExplainResponse struct {
	Items     []UserExplainItem    `yaml:"items"`
	Merged    UserIdentityResponse `yaml:"merged"`
	Authority string               `yaml:"authority"`
}

type Translated struct {
	Groups []string `yaml:"groups"`
	Uid    int64    `yaml:"uid"`
}

type UserExplainItem struct {
	UserIdentityResponse UserIdentityResponse `yaml:"userIdentityResponse"`
	Provider             struct {
		Name                string `yaml:"name"`
		CredentialAuthority bool   `yaml:"credentialAuthority"` // Is this provider Authority for authentication (password) for this user
		GroupAuthority      bool   `yaml:"groupAuthority"`      // Should we take groups in account
	} `yaml:"provider"`
	Translated Translated `yaml:"translated"`
}

// -------------------------------------------------------------------------

func (u *UserExplainRequest) String() string {
	return fmt.Sprintf("UserExplainRequest(login=%s)", u.Login)
}
func (u *UserExplainRequest) ToJson() ([]byte, error) {
	return toJson(u)
}
func (u *UserExplainRequest) FromJson(r io.Reader) error {
	return fromJson(r, u)
}

func (u *UserExplainResponse) FromJson(r io.Reader) error {
	return fromJson(r, u)
}
