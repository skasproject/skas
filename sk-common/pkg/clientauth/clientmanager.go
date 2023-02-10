package clientauth

import (
	"skas/sk-common/proto/v1/proto"
)

var _ Manager = &manager{}

type Manager interface {
	Validate(clientAuth *proto.ClientAuth) bool
}

type manager struct {
	secretById map[string]string
}

func New(configs []Config) Manager {
	cm := &manager{
		secretById: make(map[string]string),
	}
	for _, cc := range configs {
		cm.secretById[cc.Id] = cc.Secret
	}
	return cm
}

func (c manager) Validate(clientAuth *proto.ClientAuth) bool {
	scrt, ok := c.secretById[clientAuth.Id]
	if ok {
		return scrt == clientAuth.Secret || scrt == "*"
	}
	// Not found. Try anonymous
	scrt, ok = c.secretById["*"]
	if ok {
		return scrt == clientAuth.Secret || scrt == "*"
	}
	return false
}