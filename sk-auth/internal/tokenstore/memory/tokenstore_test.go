package memory

import (
	"github.com/bombsimon/logrusr/v4"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"skas/sk-auth/internal/config"
	"skas/sk-common/proto/v1/proto"
	"testing"
	"time"
)

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
}

var config3s = config.TokenConfig{
	InactivityTimeout: ParseDurationOrPanic("3s"),
	SessionMaxTTL:     ParseDurationOrPanic("24h"),
	ClientTokenTTL:    ParseDurationOrPanic("10s"),
}

func getLogger() logr.Logger {
	l := logrus.New()
	l.SetLevel(logrus.InfoLevel)
	l.SetFormatter(&logrus.TextFormatter{})
	return logrusr.New(l)
}

func TestNew(t *testing.T) {
	store := New(config3s, getLogger())
	user := proto.User{
		Login: "Alfred",
	}
	userToken, err := store.NewToken("testClient", user)
	assert.Nil(t, err)
	userToken2, err := store.Get(userToken.Token)
	assert.Nil(t, err)
	assert.NotNil(t, userToken2, "userToken should be found")
	assert.Equal(t, "Alfred", userToken2.TokenSpec.User.Login, "Login should be Alfred")
}

func TestTimeout1(t *testing.T) {
	store := New(config2s, getLogger())
	user := proto.User{
		Login: "Alfred",
	}
	userToken, err := store.NewToken("testClient", user)
	assert.Nil(t, err)
	time.Sleep(time.Second * 3)
	userToken2, err := store.Get(userToken.Token)
	assert.Nil(t, err)
	assert.Nil(t, userToken2, "userToken should be nil (Not found)")
}

func TestTimeout2(t *testing.T) {
	store := New(config2s, getLogger())
	user := proto.User{
		Login: "Alfred",
	}
	userToken, err := store.NewToken("testClient", user)
	assert.Nil(t, err)
	token := userToken.Token

	time.Sleep(time.Second)

	userToken2, err := store.Get(token)
	assert.Nil(t, err)
	assert.NotNil(t, userToken2, "userToken2 should be found")
	assert.Equal(t, "Alfred", userToken2.TokenSpec.User.Login, "Login should be Alfred")

	time.Sleep(time.Second)

	userToken2, err = store.Get(token)
	assert.Nil(t, err)
	assert.NotNil(t, userToken2, "userToken2 should be found")
	assert.Equal(t, "Alfred", userToken2.TokenSpec.User.Login, "Login should be Alfred")

	time.Sleep(time.Second)

	userToken2, err = store.Get(token)
	assert.Nil(t, err)
	assert.NotNil(t, userToken2, "userToken2 should be found")
	assert.Equal(t, "Alfred", userToken2.TokenSpec.User.Login, "Login should be Alfred")

	time.Sleep(time.Second * 3)

	userToken2, err = store.Get(token)
	assert.Nil(t, err)
	assert.Nil(t, userToken2, "userToken2 should not be found")
}
