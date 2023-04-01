package tokenstore

import (
	"skas/sk-common/proto/v1/proto"
	"time"
)

type TokenStore interface {
	NewToken(clientId string, user proto.User, authority string) (string, error)
	Get(token string) (*proto.User, error) // Return nil, nil if token does not exists or is expired. Touch token if valid
	Clean() error                          // Remove expired token
	GetClientTtl() time.Duration
	// Delete(token string) (bool, error)
}
