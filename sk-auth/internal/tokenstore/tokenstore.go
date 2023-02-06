package tokenstore

import (
	"skas/sk-common/proto/v1/proto"
	"time"
)

type TokenStore interface {
	NewToken(clientId string, user proto.User, authority string) (string, error)
	Get(token string) (*proto.User, error) // Return nil, nil if token does not exists or is expired
	Clean() error                          // Remove expired token
	// Delete(token string) (bool, error)
	GetClientTtl() time.Duration
}
