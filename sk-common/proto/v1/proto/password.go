package proto

import (
	"fmt"
	"io"
)

// ----------------------------------- PasswordChange interface

var PasswordChangeMeta = &RequestMeta{
	Method:  "POST",
	UrlPath: "/v1/passwordChange",
}

var _ RequestPayload = &PasswordChangeRequest{}

type PasswordChangeRequest struct {
	ClientAuth  ClientAuth `json:"clientAuth"`
	Provider    string     `json:"provider"`
	Login       string     `json:"login"`
	OldPassword string     `json:"oldPassword"`
	NewPassword string     `json:"newPassword"`
}

type PasswordChangeStatus string

const (
	Done               = "done"
	UnknownProvider    = "unknownProvider"
	UnknownUser        = "unknownUser"
	InvalidOldPassword = "invalidOldPassword"
	InvalidNewPassword = "invalidNewPassword" // If some password rules are implemented
	Unsupported        = "unsupported"        // This provider does not support password change
)

var _ ResponsePayload = &PasswordChangeResponse{}

type PasswordChangeResponse struct {
	Login  string               `json:"login"`
	Status PasswordChangeStatus `json:"status"`
}

// ------------------------------------------------------------------------
func (p *PasswordChangeRequest) String() string {
	return fmt.Sprintf("PasswordChangeRequest(login=%s)", p.Login)
}

func (p *PasswordChangeRequest) ToJson() ([]byte, error) {
	return toJson(p)
}

func (p *PasswordChangeRequest) FromJson(r io.Reader) error {
	return fromJson(r, p)
}

func (p *PasswordChangeResponse) FromJson(r io.Reader) error {
	return fromJson(r, p)
}
