package proto

import (
	"encoding/json"
	"fmt"
	"io"
)

// -------------------- Login API

// Ti be used by any application which want to validate user's credential.
//
// One important difference between Login and UserStatus API, is Login does not provide user info if password is not validated.

const LoginUrlPath = "/v1/login"

type LoginRequest struct {
	ClientAuth ClientAuth `json:"clientAuth"`
	Login      string     `json:"login"`
	Password   string     `json:"password"`
}

type LoginResponse struct {
	Success bool `json:"success"`
	User
}

// -----------------------------------------------------

var _ Payload = &LoginRequest{}
var _ Payload = &LoginResponse{}

func (l *LoginRequest) ToJson() ([]byte, error) {
	return toJson(l)
}

func (l *LoginRequest) FromJson(r io.Reader) error {
	return fromJson(r, l)
}

func (l *LoginRequest) String() string {
	return fmt.Sprintf("LoginRequest(login:%s", l.Login)
}

func (l *LoginResponse) ToJson() ([]byte, error) {
	return toJson(l)
}

func (l *LoginResponse) FromJson(r io.Reader) error {
	return fromJson(r, l)
}

func (l *LoginResponse) String() string {
	return fmt.Sprintf("LoginResponse(found=%q, login=%s)", l.Success, l.Login)
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
