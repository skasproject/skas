package proto

// ------------------------- Provider API

// This is the API provided by all kind of Identity provider. Consumed by sk-merge

const UserStatusUrlPath = "/v1/userstatus"

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

type UserStatusResponse struct {
	UserStatus UserStatus `json:"userStatus"`
	User
}
