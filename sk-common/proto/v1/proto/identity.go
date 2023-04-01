package proto

import (
	"fmt"
	"io"
)

var IdentityMeta = &RequestMeta{
	Name:    "identity",
	Method:  "GET",
	UrlPath: "/v1/identity",
}

type Translated struct {
	Groups []string `yaml:"groups"`
	Uid    int      `yaml:"uid"`
}

type ProviderSpec struct {
	Name                string `json:"name"`
	CredentialAuthority bool   `json:"credentialAuthority"` // Is this provider Authority for authentication (password) for this user
	GroupAuthority      bool   `json:"groupAuthority"`      // Should we take groups in account
}

type UserDetail struct {
	User
	Status       Status       `json:"status"`
	ProviderSpec ProviderSpec `json:"providerSpec"`
	Translated   Translated   `json:"translated"`
}

var _ RequestPayload = &IdentityRequest{}

type IdentityRequest struct {
	ClientAuth ClientAuth `json:"clientAuth"`
	Login      string     `json:"login"`
	Password   string     `json:"password"`
	Detailed   bool       `json:"detailed"`
}

var _ ResponsePayload = &IdentityResponse{}

type IdentityResponse struct {
	User
	Status    Status       `json:"status"`
	Details   []UserDetail `json:"details"`   // Empty is IdentityRequest.Detail == False
	Authority string       `json:"authority"` // "" if from an identity provider
}

// ----------------------------------------------------------------------

func (u *IdentityRequest) String() string {
	return fmt.Sprintf("IdentityRequest(login=%s", u.Login)
}
func (u *IdentityRequest) ToJson() ([]byte, error) {
	return toJson(u)
}
func (u *IdentityRequest) FromJson(r io.Reader) error {
	return fromJson(r, u)
}

func (u *IdentityResponse) FromJson(r io.Reader) error {
	return fromJson(r, u)
}
