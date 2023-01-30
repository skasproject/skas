package proto

// -------------------- Login interface

// One important difference between Login and UserStatus API, is Lohin does not provide user info
// if password is not validated.

const LoginUrlPath = "/login"

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
	Undefined         = "undefined" // Used to mark a non-critical failing provider in userDescribe
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

type Translated struct {
	Groups []string `yaml:"groups"`
	Uid    int64    `yaml:"uid"`
}

type UserDescribeItem struct {
	UserStatusResponse UserStatusResponse `yaml:"userStatusResponse"`
	Provider           struct {
		Name                string `yaml:"name"`
		CredentialAuthority bool   `yaml:"credentialAuthority"` // Is this provider Authority for authentication (password) for this user
		GroupAuthority      bool   `yaml:"groupAuthority"`      // Should we take groups in account
	} `yaml:"provider"`
	Translated Translated `yaml:"translated"`
}

type UserDescribeResponse struct {
	Items                       []UserDescribeItem `yaml:"items"`
	Merged                      UserStatusResponse `yaml:"merged"`
	CredentialAuthorityProvider string             `yaml:"credentialAuthorityProvider"`
}
