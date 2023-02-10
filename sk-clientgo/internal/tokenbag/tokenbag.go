package tokenbag

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	osuser "os/user"
	"path"
	"skas/sk-clientgo/internal/config"
	"skas/sk-clientgo/internal/kubecontext"
	"skas/sk-clientgo/internal/loadsave"
	"skas/sk-clientgo/internal/log"
	"skas/sk-common/pkg/misc"
	"skas/sk-common/proto/v1/proto"
	"time"
)

type TokenBag struct {
	Token      string        `yaml:"token"`
	User       proto.User    `yaml:"user"`
	Authority  string        `yaml:"authority"`
	ClientTTL  time.Duration `yaml:"clientTTL"`
	LastAccess time.Time     `yaml:"lastAccess"`
}

// Retrieve the tokenBag locally. If expired, validate again against the server. Return "" if there is no valid token
// In case of error, return ""

func Retrieve() *TokenBag {
	tokenBag := load()
	if tokenBag != nil {
		now := time.Now()
		if now.Before(tokenBag.LastAccess.Add(tokenBag.ClientTTL)) {
			// tokenBag still valid
			return tokenBag
		} else {
			if validateToken(tokenBag.Token) {
				tokenBag.LastAccess = time.Now()
				save(tokenBag)
				return tokenBag
			} else {
				DeleteTokenBag()
				return nil
			}
		}
	} else {
		return nil
	}
}

func load() *TokenBag {
	tokenBagPath := buildPath()
	var tokenBag TokenBag
	if loadsave.LoadStuff(tokenBagPath, func(decoder *yaml.Decoder) error {
		return decoder.Decode(&tokenBag)
	}) {
		log.Log.V(1).Info("LoadTokenBag()", "path", tokenBagPath, "token", misc.ShortenString(tokenBag.Token), "ClientTtl", tokenBag.ClientTTL.String(), "lastAccess", tokenBag.LastAccess)
		return &tokenBag
	} else {
		log.Log.V(1).Info("LoadTokenBag() -> nil", "path", tokenBagPath)
		return nil
	}
}

func save(tokenBag *TokenBag) {
	tokenBagPath := buildPath()
	log.Log.V(1).Info("SaveTokenBag() %s token:%s  ttl:%s  created:%s)", "path", tokenBagPath, "token", tokenBag.Token, "clientTTL", tokenBag.ClientTTL, "lastAccess", tokenBag.LastAccess)
	loadsave.SaveStuff(tokenBagPath, func(encoder *yaml.Encoder) error {
		return encoder.Encode(tokenBag)
	})
}

// Better to test and remove. Alternate would be to remove without testing, but this may hide some errors

func DeleteTokenBag() {
	tokenBagPath := buildPath()
	log.Log.V(1).Info("DeleteTokenBag()", "path", tokenBagPath)
	_, err := os.Stat(tokenBagPath)
	if !os.IsNotExist(err) {
		err := os.Remove(tokenBagPath)
		if err != nil {
			panic(err)
		}
	}
	return
}

func buildPath() string {
	usr, err := osuser.Current()
	if err != nil {
		panic(err)
	}
	return path.Join(usr.HomeDir, fmt.Sprintf(".kube/cache/sas/%s/tokenbag.json", kubecontext.KubeContext))
}

// Return false in case of error, whatever error is.

func validateToken(token string) bool {
	trr := &proto.TokenRenewRequest{
		Token:      token,
		ClientAuth: config.SkhttpClient.GetClientAuth(),
	}
	tokenRenewResponse := &proto.TokenRenewResponse{}
	err := config.SkhttpClient.Do(proto.TokenRenewMeta, trr, tokenRenewResponse)
	if err != nil {
		log.Log.Error(err, "error on ValidateToken()")
		return false
	}
	return tokenRenewResponse.Valid
}