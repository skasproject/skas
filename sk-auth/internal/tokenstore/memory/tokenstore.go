package memory

import (
	"fmt"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"skas/sk-auth/internal/config"
	"skas/sk-auth/internal/tokenstore"
	"skas/sk-auth/k8sapis/session/v1alpha1"
	"skas/sk-common/pkg/misc"
	"skas/sk-common/proto/v1/proto"
	"sort"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var _ tokenstore.TokenStore = &tokenStore{}

type tokenStore struct {
	sync.RWMutex
	tokenBagByToken  map[string]*tokenstore.TokenBag
	defaultLifecycle *v1alpha1.TokenLifecycle
	logger           logr.Logger
}

func New(conf config.TokenConfig, logger logr.Logger) tokenstore.TokenStore {
	return &tokenStore{
		tokenBagByToken: make(map[string]*tokenstore.TokenBag),
		defaultLifecycle: &v1alpha1.TokenLifecycle{
			InactivityTimeout: metav1.Duration{Duration: *conf.InactivityTimeout},
			MaxTTL:            metav1.Duration{Duration: *conf.SessionMaxTTL},
			ClientTTL:         metav1.Duration{Duration: *conf.ClientTokenTTL},
		},
		logger: logger,
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func (t *tokenStore) NewToken(clientId string, user proto.User, authority string) (tokenstore.TokenBag, error) {
	b := make([]byte, 48)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	now := time.Now()
	tokenBag := tokenstore.TokenBag{
		Token: string(b),
		TokenSpec: v1alpha1.TokenSpec{
			Client:    clientId,
			User:      user,
			Creation:  metav1.Time{Time: now},
			Lifecycle: *t.defaultLifecycle,
			Authority: authority,
		},
		LastHit: now,
	}
	t.Lock()
	t.tokenBagByToken[tokenBag.Token] = &tokenBag
	t.Unlock()
	return tokenBag, nil
}

func (t *tokenStore) Get(token string) (*tokenstore.TokenBag, error) {
	t.Lock()
	defer t.Unlock()
	bag, ok := t.tokenBagByToken[token]
	if !ok {
		return nil, nil
	}
	now := time.Now()
	if !stillValid(bag, now) {
		delete(t.tokenBagByToken, token)
		t.logger.Info(fmt.Sprintf("Token %s (login:%s) has been cleaned on Get().", misc.ShortenString(token), bag.TokenSpec.User.Login))
		return nil, nil
	}
	touch(bag, now)
	return bag, nil
}

func touch(bag *tokenstore.TokenBag, now time.Time) {
	bag.LastHit = now
}

func stillValid(bag *tokenstore.TokenBag, now time.Time) bool {
	return bag.LastHit.Add(bag.TokenSpec.Lifecycle.InactivityTimeout.Duration).After(now) && bag.TokenSpec.Creation.Add(bag.TokenSpec.Lifecycle.MaxTTL.Duration).After(now)
}

func (t *tokenStore) GetAll() ([]tokenstore.TokenBag, error) {
	t.RLock()
	slice := make([]tokenstore.TokenBag, 0, len(t.tokenBagByToken))
	for _, value := range t.tokenBagByToken {
		slice = append(slice, *value)
	}
	t.RUnlock()
	// Sort by creation
	sort.Slice(slice, func(i, j int) bool {
		return slice[i].TokenSpec.Creation.Before(&slice[j].TokenSpec.Creation)
	})
	return slice, nil
}

func (t *tokenStore) Clean() error {
	now := time.Now()
	t.Lock()
	defer t.Unlock()
	for key, value := range t.tokenBagByToken {
		if !stillValid(value, now) {
			t.logger.Info(fmt.Sprintf("Token %s (login:%s) has been cleaned in background.", misc.ShortenString(key), value.TokenSpec.User.Login))
			delete(t.tokenBagByToken, key)
		}
	}
	return nil
}

func (t *tokenStore) Delete(token string) (bool, error) {
	t.Lock()
	defer t.Unlock()
	bag, ok := t.tokenBagByToken[token]
	if ok {
		t.logger.Info(fmt.Sprintf("Token %s (login:%s) has been cleaned in background.", misc.ShortenString(bag.Token), bag.TokenSpec.User.Login))
		delete(t.tokenBagByToken, token)
	}
	return ok, nil
}
