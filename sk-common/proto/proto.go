package proto

// ------------------------ Common part

type User struct {
	Login  string   `json:"login"`
	Name   string   `json:"name"`
	Uid    int64    `json:"uid"`
	Emails []string `json:"emails"`
	Groups []string `json:"groups"`
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

type GetUserStatusRequest struct {
	Login         string `json:"login"`
	Password      string `json:"password"`
	CheckPassword bool   `json:"checkPassword"`
}

type UserStatus string

const (
	NotFound          = "notFound"
	PasswordChecked   = "passwordChecked"
	PasswordFail      = "passwordFail"
	PasswordUnchecked = "passwordUnchecked"
)

type GetUserStatusResponse struct {
	UserStatus UserStatus `json:"userStatus"`
	User       User       `json:"user"` // Fulfilled if UserStatus == PasswordChecked or UserStatus == PasswordUnchecked
}
