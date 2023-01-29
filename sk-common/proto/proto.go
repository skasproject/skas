package proto

// -------------------- Login interface

type LoginRequest struct {
	Client   string `json:"client"` // A client identifier. For information purpose
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Login       string   `json:"login"`
	Success     bool     `json:"success"`
	Uid         int64    `json:"uid"`
	CommonNames []string `json:"commonNames"`
	Emails      []string `json:"emails"`
	Groups      []string `json:"groups"`
}

// ------------------------- Provider interface

const UserStatusUrlPath = "/userstatus"

type UserStatusRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserStatus string

const (
	NotFound          = "notFound"
	Disabled          = "disabled"
	PasswordChecked   = "passwordChecked"
	PasswordFail      = "passwordFail"
	PasswordUnchecked = "passwordUnchecked"
)

type UserStatusResponse struct {
	Login       string     `json:"login"`
	UserStatus  UserStatus `json:"userStatus"`
	Uid         int64      `json:"uid"`
	CommonNames []string   `json:"commonNames"`
	Emails      []string   `json:"emails"`
	Groups      []string   `json:"groups"`
}

// ----------------------------------- UserDescribe interface

const UserDescribeUrlPath = "/userdescribe"

type UserDescribeRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserDescribeItem struct {
	UserStatusResponse
	ProviderName string `yaml:"providerName"`
	Authority    bool   `yaml:"authority"` // Is this provider Authority for authentication (password) for this user
}

type UserDescribeResponse struct {
	Items []UserDescribeItem `yaml:"items"`
}
