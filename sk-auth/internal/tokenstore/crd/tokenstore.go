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
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var _ tokenstore.TokenStore = &tokenStore{}

type tokenStore struct {
	sync.RWMutex
	config.TokenConfig
	kubeClient  client.Client
	lastHitStep time.Duration
	logger      logr.Logger
}

func New(conf config.TokenConfig, kubeClient client.Client, logger logr.Logger) tokenstore.TokenStore {
	// Convert lastHitStep from % to Duration
	lhStep := (*conf.InactivityTimeout / time.Duration(1000)) * time.Duration(conf.LastHitStep)
	return &tokenStore{
		TokenConfig: conf,
		kubeClient:  kubeClient,
		lastHitStep: lhStep,
		logger:      logger,
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
	tkn := string(b)
	now := time.Now()
	crdToken := &v1alpha1.Token{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tkn,
			Namespace: t.Namespace,
		},
		Spec: v1alpha1.TokenSpec{
			Client:    clientId,
			User:      user,
			Creation:  metav1.Time{Time: now},
			Authority: authority,
		},
		Status: v1alpha1.TokenStatus{
			LastHit: metav1.Time{Time: now},
		},
	}
	err := t.kubeClient.Create(context.TODO(), crdToken)
	if err != nil {
		t.logger.Error(err, "token create failed", "login", user.Login)
		return "", err
	}
	// Update status, as it is a subresource, so, not updated by the Create. Must set again the status
	crdToken.Status.LastHit = metav1.Time{Time: now}
	err = t.kubeClient.Status().Update(context.TODO(), crdToken)
	if err != nil {
		t.logger.Error(err, "token status update failed", "login", user.Login)
		return "", err
	}
	t.logger.V(0).Info("Token created", "token", misc.ShortenString(tkn), "login", user.Login)
	return tkn, nil
}

func (t *tokenStore) getToken(token string) (*v1alpha1.Token, error) {
	crdToken := v1alpha1.Token{}
	for retry := 0; retry < 4; retry++ {
		err := t.kubeClient.Get(context.TODO(), client.ObjectKey{
			Namespace: t.Namespace,
			Name:      token,
		}, &crdToken)
		if err == nil {
			t.logger.V(1).Info("getToken() ok", "login", crdToken.Spec.User.Login, "lastHit", crdToken.Status.LastHit)
			return &crdToken, nil
		}
		if client.IgnoreNotFound(err) != nil {
			t.logger.Error(err, "token Get() failed", "token", misc.ShortenString(token))
			return nil, err
		}
		time.Sleep(time.Millisecond * 500)
	}
	return nil, nil // Not found is not an error. May be token has been cleaned up.
}

func (t *tokenStore) Get(token string) (*proto.User, error) {
	crdToken, err := t.getToken(token)
	if crdToken == nil {
		return nil, err // two cases: err==nil => not found   err!=nil => real problem
	}
	now := time.Now()
	if t.stillValid(crdToken, now) {
		err := t.touch(crdToken, now)
		if err != nil {
			t.logger.Error(err, "token touch on Get() failed. Will retry", "token", misc.ShortenString(token), "login", crdToken.Spec.User.Login)
			//  the object has been modified; please apply your changes to the latest version and try again
			crdToken, err := t.getToken(token)
			if crdToken == nil {
				return nil, err // two cases: err==nil => not found   err!=nil => real problem
			}
			err = t.touch(crdToken, now)
			if err != nil {
				t.logger.Error(err, "token touch on Get() failed a second time. Aborting", "token", misc.ShortenString(token), "login", crdToken.Spec.User.Login)
				return nil, err
			}
		}
		user := &proto.User{}
		crdToken.Spec.User.DeepCopyInto(user)
		return user, nil
	} else {
		err := t.delete(crdToken)
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

func (t *tokenStore) stillValid(tkn *v1alpha1.Token, now time.Time) bool {
	t.logger.V(1).Info("stillValid()", "login", tkn.Spec.User.Login, "lastHit", tkn.Status.LastHit, "creation", tkn.Spec.Creation, "now", now)
	return tkn.Status.LastHit.Add(*t.InactivityTimeout).After(now) && tkn.Spec.Creation.Add(*t.SessionMaxTTL).After(now)
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

func (t *tokenStore) Clean() error {
	now := time.Now()
	list := v1alpha1.TokenList{}
	err := t.kubeClient.List(context.TODO(), &list, client.InNamespace(t.Namespace))
	if err != nil {
		t.logger.Error(err, "Token Cleaner. List failed")
		return err
	}
	for i := 0; i < len(list.Items); i++ {
		crdToken := list.Items[i]
		if !t.stillValid(&crdToken, now) {
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
