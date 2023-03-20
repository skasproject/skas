package proto

import (
	"fmt"
	"io"
)

// ------------------------- Provider API

// This is the API provided by all kind of Identity provider. Consumed by sk-merge

var UserIdentityMeta = &RequestMeta{
	Name:    "userIdentity",
	Method:  "GET",
	UrlPath: "/v1/userIdentity",
}

var _ RequestPayload = &UserIdentityRequest{}

type UserIdentityRequest struct {
	ClientAuth ClientAuth `json:"clientAuth"`
	Login      string     `json:"login"`
	Password   string     `json:"password"`
}

type UserStatus string

const (
	NotFound          = "notFound"
	Disabled          = "disabled"
	PasswordChecked   = "passwordChecked"
	PasswordFail      = "passwordFail"
	PasswordUnchecked = "passwordUnchecked"
	Undefined         = "undefined" // Used to mark a non-critical failing provider in userDescribe
)

var _ ResponsePayload = &UserIdentityResponse{}

type UserIdentityResponse struct {
	UserStatus UserStatus `json:"userStatus"`
	User
}

// ---------------------------------------------------------------

func (u *UserIdentityRequest) String() string {
	return fmt.Sprintf("UserIdentityRequest(login=%s", u.Login)
}
func (u *UserIdentityRequest) ToJson() ([]byte, error) {
	return toJson(u)
}
func (u *UserIdentityRequest) FromJson(r io.Reader) error {
	return fromJson(r, u)
}

func (u *UserIdentityResponse) FromJson(r io.Reader) error {
	return fromJson(r, u)
}
