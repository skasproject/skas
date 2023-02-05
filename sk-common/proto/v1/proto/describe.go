package proto

import (
	"fmt"
	"io"
)

// ----------------------------------- UserDescribe interface

// This is issued by sk-cli to sk-auth, which validate the token.
// Then, it is forwarded to sk-merge, without Token but with ClientAuth

const UserDescribeUrlPath = "/v1/userdescribe"

var _ RequestPayload = &UserDescribeRequest{}

type UserDescribeRequest struct {
	ClientAuth ClientAuth `json:"clientAuth"`
	Token      string     `json:"token"`
	Login      string     `json:"login"`
	Password   string     `json:"password"` // Optional
}

var _ ResponsePayload = &UserDescribeResponse{}

type UserDescribeResponse struct {
	Items     []UserDescribeItem `yaml:"items"`
	Merged    UserStatusResponse `yaml:"merged"`
	Authority string             `yaml:"authority"`
}

type Translated struct {
	Groups []string `yaml:"groups"`
	Uid    int64    `yaml:"uid"`
}

type UserDescribeItem struct {
	UserStatusResponse UserStatusResponse `yaml:"userStatusResponse"`
	Provider           struct {
		Name                string `yaml:"name"`
		CredentialAuthority bool   `yaml:"credentialAuthority"` // Is this provider Authority for authentication (password) for this user
		GroupAuthority      bool   `yaml:"groupAuthority"`      // Should we take groups in account
	} `yaml:"provider"`
	Translated Translated `yaml:"translated"`
}

// -------------------------------------------------------------------------

func (u *UserDescribeRequest) String() string {
	return fmt.Sprintf("UserDescribeRequest(login=%s)", u.Login)
}
func (u *UserDescribeRequest) ToJson() ([]byte, error) {
	return toJson(u)
}
func (u *UserDescribeRequest) FromJson(r io.Reader) error {
	return fromJson(r, u)
}

func (u *UserDescribeResponse) FromJson(r io.Reader) error {
	return fromJson(r, u)
}
