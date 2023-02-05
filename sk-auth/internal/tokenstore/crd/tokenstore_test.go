package crd

import (
	"github.com/bombsimon/logrusr/v4"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"skas/sk-auth/internal/config"
	"skas/sk-auth/k8sapis/session/v1alpha1"
	"skas/sk-common/proto/v1/proto"
	"testing"
	"time"
)

func newClient() client.Client {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = "~/.kube/config"
	}
	myconfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}
	crScheme := runtime.NewScheme()
	err = v1alpha1.AddToScheme(crScheme)
	if err != nil {
		panic(err)
	}
	myclient, err := client.New(myconfig, client.Options{
		Scheme: crScheme,
	})
	if err != nil {
		panic(err)
	}
	return myclient
}

func getLogger() logr.Logger {
	l := logrus.New()
	l.SetLevel(logrus.InfoLevel)
	l.SetFormatter(&logrus.TextFormatter{})
	return logrusr.New(l)
}

func ParseDurationOrPanic(d string) *time.Duration {
	duration, err := time.ParseDuration(d)
	if err != nil {
		panic(err)
	}
	return &duration
}

var config2s = config.TokenConfig{
	InactivityTimeout: ParseDurationOrPanic("2s"),
	SessionMaxTTL:     ParseDurationOrPanic("24h"),
	ClientTokenTTL:    ParseDurationOrPanic("10s"),
	Namespace:         "skas-system",
	LastHitStep:       3,
}

var config3s = config.TokenConfig{
	InactivityTimeout: ParseDurationOrPanic("3s"),
	SessionMaxTTL:     ParseDurationOrPanic("24h"),
	ClientTokenTTL:    ParseDurationOrPanic("10s"),
	Namespace:         "skas-system",
	LastHitStep:       3,
}

func TestMain(m *testing.M) {

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	store := New(config3s, newClient(), getLogger())
	user := proto.User{
		Login:       "Alfred",
		Emails:      []string{},
		Groups:      []string{"xx"},
		CommonNames: []string{},
		Uid:         0,
	}
	tokenBag, err := store.NewToken("testClient", user)
	assert.Nil(t, err)
	assert.NotNil(t, tokenBag)
	time.Sleep(time.Second * 1)
	tokenBag2, err := store.Get(tokenBag.Token)
	assert.Nil(t, err)
	assert.NotNil(t, tokenBag2, "tokenBag2 should be found")
	assert.Equal(t, "Alfred", tokenBag2.TokenSpec.User.Login, "Login should be Alfred")
}

func TestTimeout1(t *testing.T) {
	store := New(config2s, newClient(), getLogger())
	user := proto.User{
		Login:       "Alfred",
		Emails:      []string{},
		Groups:      []string{"xx"},
		CommonNames: []string{},
		Uid:         1,
	}
	tokenBag, err := store.NewToken("testClient", user)
	assert.Nil(t, err)
	time.Sleep(time.Second * 3)
	userToken2, err := store.Get(tokenBag.Token)
	assert.Nil(t, err)
	assert.Nil(t, userToken2, "userToken should be nil (Not found)")
}

func TestTimeout2(t *testing.T) {
	store := New(config3s, newClient(), getLogger())
	user := proto.User{
		Login:       "Alfred",
		Emails:      []string{},
		Groups:      []string{"xx"},
		CommonNames: []string{},
		Uid:         2,
	}
	tokenBag, err := store.NewToken("testClient", user)
	assert.Nil(t, err)

	token := tokenBag.Token

	time.Sleep(time.Second)

	tokenBag2, err := store.Get(token)
	assert.Nil(t, err)
	assert.NotNil(t, tokenBag2, "tokenBag2 should be found")
	assert.Equal(t, "Alfred", tokenBag2.TokenSpec.User.Login, "Login should be Alfred")

	time.Sleep(time.Second)

	tokenBag2, err = store.Get(token)
	assert.Nil(t, err)
	assert.NotNil(t, tokenBag2, "tokenBag2 should be found")
	assert.Equal(t, "Alfred", tokenBag2.TokenSpec.User.Login, "Login should be Alfred")

	time.Sleep(time.Second)

	tokenBag2, err = store.Get(token)
	assert.Nil(t, err)
	assert.NotNil(t, tokenBag2, "tokenBag2 should be found")
	assert.Equal(t, "Alfred", tokenBag2.TokenSpec.User.Login, "Login should be Alfred")

	time.Sleep(time.Second * 3)

	tokenBag2, err = store.Get(token)
	assert.Nil(t, err)
	assert.Nil(t, tokenBag2, "tokenBag2 should not be found")
}

func TestMultipleGet(t *testing.T) {
	store := New(config3s, newClient(), getLogger())
	user := proto.User{
		Login:       "Alfred",
		Emails:      []string{},
		Groups:      []string{"xx"},
		CommonNames: []string{},
		Uid:         3,
	}
	tokenBag, err := store.NewToken("testClient", user)
	assert.Nil(t, err)

	userToken2, err := store.Get(tokenBag.Token)
	assert.Nil(t, err)
	assert.NotNil(t, userToken2, "userToken2 should be found")
	assert.Equal(t, "Alfred", userToken2.TokenSpec.User.Login, "Login should be Alfred")

	time.Sleep(time.Second)
	userToken2, err = store.Get(tokenBag.Token)
	assert.Nil(t, err)
	assert.NotNil(t, userToken2, "userToken2 should be found")
	assert.Equal(t, "Alfred", userToken2.TokenSpec.User.Login, "Login should be Alfred")

	userToken2, err = store.Get(tokenBag.Token)
	assert.Nil(t, err)
	assert.NotNil(t, userToken2, "userToken2 should be found")
	assert.Equal(t, "Alfred", userToken2.TokenSpec.User.Login, "Login should be Alfred")
}

func TestMultipleBasket(t *testing.T) {
	user := proto.User{
		Login:       "Alfred",
		Emails:      []string{},
		Groups:      []string{"xx"},
		CommonNames: []string{},
		Uid:         4,
	}
	basket1 := New(config3s, newClient(), getLogger())
	basket2 := New(config3s, newClient(), getLogger())
	userToken, err := basket1.NewToken("testClient", user)
	assert.Nil(t, err)

	userToken2, err := basket1.Get(userToken.Token)
	assert.Nil(t, err)
	assert.NotNil(t, userToken2, "userToken2 should be found")
	assert.Equal(t, "Alfred", userToken2.TokenSpec.User.Login, "Login should be Alfred")

	//time.Sleep(time.Second * 2)
	userToken2, err = basket2.Get(userToken.Token)
	assert.Nil(t, err)
	assert.NotNil(t, userToken2, "userToken2 should be found")
	assert.Equal(t, "Alfred", userToken2.TokenSpec.User.Login, "Login should be Alfred")

	time.Sleep(time.Second)
	userToken2, err = basket1.Get(userToken.Token)
	assert.Nil(t, err)
	assert.NotNil(t, userToken2, "ok should be found")
	assert.Equal(t, "Alfred", userToken2.TokenSpec.User.Login, "Login should be Alfred")
}
