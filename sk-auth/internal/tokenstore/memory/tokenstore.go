package memory

import (
	"fmt"
	"github.com/go-logr/logr"
	"math/rand"
	"skas/sk-auth/internal/config"
	"skas/sk-auth/internal/tokenstore"
	"skas/sk-common/pkg/misc"
	"skas/sk-common/proto/v1/proto"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var _ tokenstore.TokenStore = &tokenStore{}

type tokenBag struct {
	token     string
	clientId  string
	user      proto.User
	authority string
	creation  time.Time
	lastHit   time.Time
}

type tokenStore struct {
	sync.RWMutex
	config.TokenConfig
	tokenBagByToken map[string]*tokenBag
	logger          logr.Logger
}

func New(conf config.TokenConfig, logger logr.Logger) tokenstore.TokenStore {
	return &tokenStore{
		TokenConfig:     conf,
		tokenBagByToken: make(map[string]*tokenBag),
		logger:          logger,
	}
}

func (t *tokenStore) GetClientTtl() time.Duration {
	return *t.ClientTokenTTL
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func (t *tokenStore) NewToken(clientId string, user proto.User, authority string) (string, error) {
	b := make([]byte, 48)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	now := time.Now()
	tkn := string(b)
	tokenBag := &tokenBag{
		token:     tkn,
		clientId:  clientId,
		user:      user,
		authority: authority,
		lastHit:   now,
		creation:  now,
	}
	t.Lock()
	t.tokenBagByToken[tkn] = tokenBag
	t.Unlock()
	return tkn, nil
}

func (t *tokenStore) Get(token string) (*proto.User, error) {
	t.Lock()
	defer t.Unlock()
	bag, ok := t.tokenBagByToken[token]
	if !ok {
		return nil, nil
	}
	now := time.Now()
	if !t.stillValid(bag, now) {
		delete(t.tokenBagByToken, token)
		t.logger.Info(fmt.Sprintf("Token %s (login:%s) has been cleaned on Get().", misc.ShortenString(token), bag.user.Login))
		return nil, nil
	}
	touch(bag, now)
	return &bag.user, nil
}

func touch(bag *tokenBag, now time.Time) {
	bag.lastHit = now
}

func (t *tokenStore) stillValid(bag *tokenBag, now time.Time) bool {
	return bag.lastHit.Add(*t.InactivityTimeout).After(now) && bag.creation.Add(*t.SessionMaxTTL).After(now)
}

func (t *tokenStore) Clean() error {
	now := time.Now()
	t.Lock()
	defer t.Unlock()
	for token, bag := range t.tokenBagByToken {
		if !t.stillValid(bag, now) {
			t.logger.Info(fmt.Sprintf("Token %s (login:%s) has been cleaned in background.", misc.ShortenString(token), bag.user.Login))
			delete(t.tokenBagByToken, token)
		}
	}
	return nil
}
