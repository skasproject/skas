package proto

// ------------------------ Common part

type User struct {
	Login       string   `json:"login"`
	Uid         int64    `json:"uid"`
	CommonNames []string `json:"commonNames"`
	Emails      []string `json:"emails"`
	Groups      []string `json:"groups"`
}

// -------------------- Login interface

type LoginRequest struct {
	Client   string `json:"client"` // A client identifier. For information purpose
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool `json:"success"`
	User    User `json:"user"` // Fulfilled only if Success == True
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
	UserStatus UserStatus `json:"userStatus"`
	User       *User      `json:"user,omitempty"` // Fulfilled if UserStatus == PasswordChecked or UserStatus == PasswordUnchecked
}
