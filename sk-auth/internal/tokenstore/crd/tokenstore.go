package crd

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	defaultLifecycle *v1alpha1.TokenLifecycle
	kubeClient       client.Client
	namespace        string
	lastHitStep      time.Duration
	logger           logr.Logger
}

func New(conf config.TokenConfig, kubeClient client.Client, logger logr.Logger) tokenstore.TokenStore {
	defaultLifecycle := &v1alpha1.TokenLifecycle{
		InactivityTimeout: metav1.Duration{Duration: *conf.InactivityTimeout},
		MaxTTL:            metav1.Duration{Duration: *conf.SessionMaxTTL},
		ClientTTL:         metav1.Duration{Duration: *conf.ClientTokenTTL},
	}
	// Convert lastHitStep from % to Duration
	lhStep := (defaultLifecycle.InactivityTimeout.Duration / time.Duration(1000)) * time.Duration(conf.LastHitStep)
	return &tokenStore{
		defaultLifecycle: defaultLifecycle,
		kubeClient:       kubeClient,
		namespace:        conf.Namespace,
		lastHitStep:      lhStep,
		logger:           logger,
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func (t *tokenStore) NewToken(clientId string, user proto.User) (tokenstore.TokenBag, error) {
	b := make([]byte, 48)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	tkn := string(b)
	now := time.Now()
	crdToken := &v1alpha1.Token{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tkn,
			Namespace: t.namespace,
		},
		Spec: v1alpha1.TokenSpec{
			Client:    clientId,
			User:      user,
			Creation:  metav1.Time{Time: now},
			Lifecycle: *t.defaultLifecycle,
		},
		Status: v1alpha1.TokenStatus{
			LastHit: metav1.Time{Time: now},
		},
	}
	tokenBag := tokenstore.TokenBag{
		Token:     tkn,
		TokenSpec: v1alpha1.TokenSpec{},
		LastHit:   now,
	}
	crdToken.Spec.DeepCopyInto(&tokenBag.TokenSpec)
	err := t.kubeClient.Create(context.TODO(), crdToken)
	if err != nil {
		t.logger.Error(err, "token create failed", "login", user.Login)
		return tokenstore.TokenBag{}, err
	}
	// Update status, as it is a subresource, so, not updated by the Create. Must set again the status
	crdToken.Status.LastHit = metav1.Time{Time: now}
	err = t.kubeClient.Status().Update(context.TODO(), crdToken)
	if err != nil {
		t.logger.Error(err, "token status update failed", "login", user.Login)
		return tokenstore.TokenBag{}, err
	}
	t.logger.V(0).Info("Token created", "token", misc.ShortenString(tkn), "login", user.Login)

	return tokenBag, nil
}

func (t *tokenStore) getToken(token string) (v1alpha1.Token, bool, error) {
	crdToken := v1alpha1.Token{}
	for retry := 0; retry < 4; retry++ {
		err := t.kubeClient.Get(context.TODO(), client.ObjectKey{
			Namespace: t.namespace,
			Name:      token,
		}, &crdToken)
		if err == nil {
			return crdToken, true, nil
		}
		if client.IgnoreNotFound(err) != nil {
			t.logger.Error(err, "token Get() failed", "token", misc.ShortenString(token))
			return crdToken, false, err
		}
		time.Sleep(time.Millisecond * 500)
	}
	return crdToken, false, nil // Not found is not an error. May be token has been cleaned up.
}

func (t *tokenStore) Get(token string) (*tokenstore.TokenBag, error) {
	crdToken, found, err := t.getToken(token)
	if !found {
		return nil, err
	}
	now := time.Now()
	if stillValid(&crdToken, now) {
		err := t.touch(&crdToken, now)
		if err != nil {
			t.logger.Error(err, "token touch on Get() failed", "token", misc.ShortenString(token), "login", crdToken.Spec.User.Login)
			return nil, err
		}
		tokenBag := tokenstore.TokenBag{
			Token:     token,
			TokenSpec: v1alpha1.TokenSpec{},
			LastHit:   crdToken.Status.LastHit.Time,
		}
		crdToken.Spec.DeepCopyInto(&tokenBag.TokenSpec)

		return &tokenBag, nil
	} else {
		err := t.delete(&crdToken)
		if err != nil {
			return nil, err
		}
		t.logger.Info("Token has been cleaned on Get()", "token", misc.ShortenString(token), "login", crdToken.Spec.User.Login)
		return nil, nil
	}

}

func (t *tokenStore) delete(tkn *v1alpha1.Token) error {
	return t.kubeClient.Delete(context.TODO(), tkn, client.GracePeriodSeconds(0))
}

func stillValid(tkn *v1alpha1.Token, now time.Time) bool {
	return tkn.Status.LastHit.Add(tkn.Spec.Lifecycle.InactivityTimeout.Duration).After(now) && tkn.Spec.Creation.Add(tkn.Spec.Lifecycle.MaxTTL.Duration).After(now)
}

func (t *tokenStore) touch(tkn *v1alpha1.Token, now time.Time) error {
	if now.After(tkn.Status.LastHit.Add(t.lastHitStep)) {
		t.logger.V(1).Info("Will effectively update LastHit", "token", tkn.Name, "login", tkn.Spec.User.Login)
		tkn.Status.LastHit = metav1.Time{Time: now}
		err := t.kubeClient.Status().Update(context.TODO(), tkn)
		if err != nil {
			return err
		}
	} else {
		t.logger.V(1).Info("LastHit update skipped, as too early", "token", tkn.Name, "user", tkn.Spec.User.Login)
	}
	return nil
}

func (t *tokenStore) GetAll() ([]tokenstore.TokenBag, error) {
	list := v1alpha1.TokenList{}
	err := t.kubeClient.List(context.TODO(), &list, client.InNamespace(t.namespace))
	if err != nil {
		t.logger.Error(err, "token List failed")
		return nil, err
	}
	slice := make([]tokenstore.TokenBag, 0, len(list.Items))
	for i := 0; i < len(list.Items); i++ {
		tokenBag := tokenstore.TokenBag{
			Token:     list.Items[i].Name,
			TokenSpec: v1alpha1.TokenSpec{},
			LastHit:   list.Items[i].Status.LastHit.Time,
		}
		list.Items[i].Spec.DeepCopyInto(&tokenBag.TokenSpec)
		slice = append(slice, tokenBag)
	}
	sort.Slice(slice, func(i, j int) bool {
		return slice[i].TokenSpec.Creation.Before(&slice[j].TokenSpec.Creation)
	})
	return slice, nil
}

func (t *tokenStore) Clean() error {
	now := time.Now()
	list := v1alpha1.TokenList{}
	err := t.kubeClient.List(context.TODO(), &list, client.InNamespace(t.namespace))
	if err != nil {
		t.logger.Error(err, "Token Cleaner. List failed")
		return err
	}
	for i := 0; i < len(list.Items); i++ {
		crdToken := list.Items[i]
		if !stillValid(&crdToken, now) {
			t.logger.Info(fmt.Sprintf("Token %s (login:%s) has been cleaned in background.", misc.ShortenString(crdToken.Name), crdToken.Spec.User.Login))
			err := t.delete(&crdToken)
			if err != nil {
				t.logger.Error(err, "Error on delete", "token", misc.ShortenString(crdToken.Name), "login", crdToken.Spec.User.Login)
				return err
			}
		}
	}
	return nil

}

func (t *tokenStore) Delete(token string) (bool, error) {
	crdToken := v1alpha1.Token{}
	err := t.kubeClient.Get(context.TODO(), client.ObjectKey{
		Namespace: t.namespace,
		Name:      token,
	}, &crdToken)
	if client.IgnoreNotFound(err) != nil {
		t.logger.Error(err, "token Get() failed", "token", misc.ShortenString(token))
		return false, err
	}
	if err != nil {
		// Token not found. Not an error (May be cleaned)
		return false, nil
	}
	err = t.delete(&crdToken)
	if err != nil {
		t.logger.Error(err, "Error on delete", "token", misc.ShortenString(crdToken.Name), "login", crdToken.Spec.User.Login)
		return false, err
	}
	return true, nil

}
