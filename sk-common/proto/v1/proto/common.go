package proto

import (
	"encoding/json"
	"fmt"
	"io"
)

// ----------------------------Shared stuff

type ClientAuth struct {
	Id     string `json:"id"`
	Secret string `json:"secret"`
}

type RequestMeta struct {
	Name    string // For debug and message
	Method  string
	UrlPath string
}

type Status string

// If password is not provided in the request and there is no password in the user definition, status should be 'passwordMissing' (Not 'passwordUnchecked')
const (
	// ---- Following for identity and login response
	UserNotFound      = "userNotFound"
	Disabled          = "disabled"
	PasswordChecked   = "passwordChecked"
	PasswordFail      = "passwordFail"
	PasswordUnchecked = "passwordUnchecked" // Because password was not provided in the request
	PasswordMissing   = "passwordMissing"   // Because this provider does not store a password for this user
	Undefined         = "undefined"         // Used to mark a non-critical failing provider in userDescribe
	// ---- Following is specific to passwordChange
	PasswordChanged    = "passwordChanged"
	UnknownProvider    = "unknownProvider"
	InvalidOldPassword = "invalidOldPassword"
	InvalidNewPassword = "invalidNewPassword" // If some password rules are implemented
	Unsupported        = "unsupported"        // This provider does not support password change
)

// This object is also used in Token K8s api in sk-auth

// +kubebuilder:object:generate:=true
type User struct {
	Login       string   `json:"login"`
	Uid         int      `json:"uid"`
	CommonNames []string `json:"commonNames"`
	Emails      []string `json:"emails"`
	Groups      []string `json:"groups"`
}

func InitUser(login string) User {
	return User{
		Login:       login,
		Uid:         0,
		CommonNames: []string{},
		Emails:      []string{},
		Groups:      []string{},
	}
}

type RequestPayload interface {
	fmt.Stringer // For debug & error message
	ToJson() ([]byte, error)
	FromJson(r io.Reader) error
}

type ResponsePayload interface {
	FromJson(r io.Reader) error
}

// -----------------------------------------------------

func toJson(payload interface{}) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}

func fromJson(r io.Reader, payload interface{}) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	return decoder.Decode(payload)
}
