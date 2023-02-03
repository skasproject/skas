package proto

// ----------------------------Shared stuff

type ClientAuth struct {
	Id     string `json:"id"`
	Secret string `json:"secret"`
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
