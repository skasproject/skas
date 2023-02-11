package proto

import (
	"encoding/json"
	"fmt"
	"io"
)

// ----------------------------Shared stuff

// This structure is also used in configuration. So the yaml

type ClientAuth struct {
	Id     string `json:"id" yaml:"id"`
	Secret string `json:"secret" yaml:"secret"`
}

type RequestMeta struct {
	Method  string
	UrlPath string
}

// This object is also used in Token K8s api in sk-auth

// +kubebuilder:object:generate:=true
type User struct {
	Login       string   `json:"login"`
	Uid         int64    `json:"uid"`
	CommonNames []string `json:"commonNames"`
	Emails      []string `json:"emails"`
	Groups      []string `json:"groups"`
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
