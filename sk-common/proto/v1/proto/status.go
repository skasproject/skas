package proto

import (
	"fmt"
	"io"
)

// ------------------------- Provider API

// This is the API provided by all kind of Identity provider. Consumed by sk-merge

const UserStatusUrlPath = "/v1/userstatus"

var _ RequestPayload = &UserStatusRequest{}

type UserStatusRequest struct {
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

var _ ResponsePayload = &UserStatusResponse{}

type UserStatusResponse struct {
	UserStatus UserStatus `json:"userStatus"`
	User
}

// ---------------------------------------------------------------

func (u *UserStatusRequest) String() string {
	return fmt.Sprintf("UserStatusRequest(login=%s", u.Login)
}
func (u *UserStatusRequest) ToJson() ([]byte, error) {
	return toJson(u)
}
func (u *UserStatusRequest) FromJson(r io.Reader) error {
	return fromJson(r, u)
}

func (u *UserStatusResponse) FromJson(r io.Reader) error {
	return fromJson(r, u)
}
