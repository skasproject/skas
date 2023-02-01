package clientmanager

import "skas/sk-common/proto/v1/proto"

// This is the structure to define a client in our configuration

type ClientConfig struct {
	Id     string `yaml:"id"`
	Secret string `yaml:"secret"`
}

var _ ClientManager = &clientManager{}

type ClientManager interface {
	Validate(clientAuth *proto.ClientAuth) bool
}

type clientManager struct {
	secretById map[string]string
}

func New(clientConfigs []ClientConfig) ClientManager {
	cm := &clientManager{
		secretById: make(map[string]string),
	}
	for _, cc := range clientConfigs {
		cm.secretById[cc.Id] = cc.Secret
	}
	return cm
}

func (c clientManager) Validate(clientAuth *proto.ClientAuth) bool {
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
