package clientproviderchain

import (
	"skas/sk-common/proto"
	"skas/sk-merge/internal/clientprovider"
)

type ScanItem struct {
	UserStatusResponse *proto.UserStatusResponse
	Provider           *clientprovider.ClientProvider
	Translated         *proto.Translated
}

type ClientProviderChain interface {
	Scan(login, password string) ([]ScanItem, error)
	GetLength() int
}
