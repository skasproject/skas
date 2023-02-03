package tokenstore

import (
	"skas/sk-auth/k8sapi/session/v1alpha1"
	"skas/sk-common/proto/v1/proto"
	"time"
)

type TokenBag struct {
	Token     string
	TokenSpec v1alpha1.TokenSpec
	LastHit   time.Time
}

type TokenStore interface {
	NewToken(clientId string, user proto.User) (TokenBag, error)
	Get(token string) (*TokenBag, error) // Return nil, nil if token does not exists or is expired
	GetAll() ([]TokenBag, error)
	Clean() error // Remove expired token
	Delete(token string) (bool, error)
}
