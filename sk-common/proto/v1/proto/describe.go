package proto

// ----------------------------------- UserDescribe interface

// This is issued by sk-cli to sk-auth, which validate the token.
// Then, it is forwarded to sk-merge, without Token but with ClientAuth

const UserDescribeUrlPath = "/v1/userdescribe"

type UserDescribeRequest struct {
	ClientAuth ClientAuth `json:"clientAuth"`
	Token      string     `json:"token"`
	Login      string     `json:"login"`
	Password   string     `json:"password"` // Optional
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
