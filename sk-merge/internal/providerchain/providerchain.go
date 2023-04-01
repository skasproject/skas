package providerchain

import (
	"skas/sk-common/proto/v1/proto"
	"skas/sk-merge/internal/provider"
)

type ScanItem struct {
	UserIdentityResponse *proto.IdentityResponse
	Provider             *provider.Provider
	Translated           *proto.Translated
}

type ProviderChain interface {
	GetIdentity(login, password string, detailed bool) (*proto.IdentityResponse, error)
	// ChangePassword - We pass request by value, as we may modify it
	ChangePassword(request proto.PasswordChangeRequest) (*proto.PasswordChangeResponse, error)

	//Scan(login, password string) ([]ScanItem, error)
	//GetLength() int

}
