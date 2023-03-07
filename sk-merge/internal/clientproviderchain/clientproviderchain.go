package clientproviderchain

import (
	"skas/sk-common/proto/v1/proto"
	"skas/sk-merge/internal/clientprovider"
)

type ScanItem struct {
	UserIdentityResponse *proto.UserIdentityResponse
	Provider             *clientprovider.ClientProvider
	Translated           *proto.Translated
}

type ClientProviderChain interface {
	Scan(login, password string) ([]ScanItem, error)
	GetLength() int
}
