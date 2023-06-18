package proto

import (
	"fmt"
	"io"
)

var PasswordStrengthMeta = &RequestMeta{
	Name:    "passwordStrength",
	Method:  "GET",
	UrlPath: "/v1/passwordStrength",
}

var _ RequestPayload = &PasswordStrengthRequest{}

type PasswordStrengthRequest struct {
	ClientAuth ClientAuth `json:"clientAuth"`
	Password   string     `json:"password"`
	UserInputs []string   `json:"userInputs"`
}

var _ ResponsePayload = &PasswordStrengthResponse{}

type PasswordStrengthResponse struct {
	Password   string `json:"password"`
	Score      int    `json:"score"`
	IsCommon   bool   `json:"isCommon"`
	Acceptable bool   `json:"acceptable"` // Related to configured requirement
}

// ------------------------------------------------------------------------

func (p *PasswordStrengthRequest) String() string {
	return fmt.Sprintf("PasswordStrengthRequest()")
}

func (p *PasswordStrengthRequest) ToJson() ([]byte, error) {
	return toJson(p)
}

func (p *PasswordStrengthRequest) FromJson(r io.Reader) error {
	return fromJson(r, p)
}

func (p *PasswordStrengthResponse) FromJson(r io.Reader) error {
	return fromJson(r, p)
}
