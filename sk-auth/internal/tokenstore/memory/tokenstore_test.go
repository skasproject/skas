package memory

import (
	"github.com/bombsimon/logrusr/v4"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
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

var configMaxTtl = config.TokenConfig{
	InactivityTimeout: ParseDurationOrPanic("1h"),
	SessionMaxTTL:     ParseDurationOrPanic("3s"),
	ClientTokenTTL:    ParseDurationOrPanic("10s"),
}

func getLogger() logr.Logger {
	l := logrus.New()
	l.SetLevel(logrus.InfoLevel)
	l.SetFormatter(&logrus.TextFormatter{})
	return logrusr.New(l)
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	store := New(config3s, getLogger())
	user := proto.User{
		Login: "Alfred",
	}
	userToken, err := store.NewToken("testClient", user, "auth")
	assert.Nil(t, err)
	user2, err := store.Get(userToken)
	assert.Nil(t, err)
	assert.NotNil(t, user2, "user should be found")
	assert.Equal(t, "Alfred", user2.Login, "Login should be Alfred")
}

func TestTimeout1(t *testing.T) {
	store := New(config2s, getLogger())
	user := proto.User{
		Login: "Alfred",
	}
	userToken, err := store.NewToken("testClient", user, "auth")
	assert.Nil(t, err)
	time.Sleep(time.Second * 3)
	user2, err := store.Get(userToken)
	assert.Nil(t, err)
	assert.Nil(t, user2, "userToken should be nil (Not found)")
}

func TestTimeout2(t *testing.T) {
	store := New(config2s, getLogger())
	user := proto.User{
		Login: "Alfred",
	}
	token, err := store.NewToken("testClient", user, "auth")
	assert.Nil(t, err)

	time.Sleep(time.Second)

	user2, err := store.Get(token)
	assert.Nil(t, err)
	assert.NotNil(t, user2, "user2 should be found")
	assert.Equal(t, "Alfred", user2.Login, "Login should be Alfred")

	time.Sleep(time.Second)

	user2, err = store.Get(token)
	assert.Nil(t, err)
	assert.NotNil(t, user2, "user2 should be found")
	assert.Equal(t, "Alfred", user2.Login, "Login should be Alfred")

	time.Sleep(time.Second)

	user2, err = store.Get(token)
	assert.Nil(t, err)
	assert.NotNil(t, user2, "user2 should be found")
	assert.Equal(t, "Alfred", user2.Login, "Login should be Alfred")

	time.Sleep(time.Second * 3)

	user2, err = store.Get(token)
	assert.Nil(t, err)
	assert.Nil(t, user2, "user2 should not be found")
}

func TestTimeout3(t *testing.T) {

	store := New(configMaxTtl, getLogger())
	user := proto.User{
		Login:       "Alfred",
		Emails:      []string{},
		Groups:      []string{"xx"},
		CommonNames: []string{},
		Uid:         2,
	}
	token, err := store.NewToken("testClient", user, "auth")
	assert.Nil(t, err)

	time.Sleep(time.Second)

	user2, err := store.Get(token)
	assert.Nil(t, err)
	assert.NotNil(t, user2, "user2 should be found")
	assert.Equal(t, "Alfred", user2.Login, "Login should be Alfred")

	time.Sleep(time.Second)

	user2, err = store.Get(token)
	assert.Nil(t, err)
	assert.NotNil(t, user2, "user2 should be found")
	assert.Equal(t, "Alfred", user2.Login, "Login should be Alfred")

	time.Sleep(time.Second * 3)

	user2, err = store.Get(token)
	assert.Nil(t, err)
	assert.Nil(t, user2, "user2 should not be found")
}
