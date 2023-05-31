package protector

import "skas/sk-common/proto/v1/proto"

var _ Protector = &empty{}

type empty struct{}

func (e empty) ProtectLoginResult(login string, status proto.Status) {
}

func (e empty) EntryForLogin(login string) (locked bool) {
	return false
}

func (e empty) EntryForToken() (locked bool) {
	return false
}

func (e empty) TokenNotFound() {
}
