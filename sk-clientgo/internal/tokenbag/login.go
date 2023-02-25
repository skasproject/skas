package tokenbag

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"skas/sk-clientgo/internal/config"
	"skas/sk-clientgo/internal/log"
	"skas/sk-common/proto/v1/proto"
	"strings"
	"syscall"
	"time"
)

func InteractiveLogin(login, password string) *TokenBag {
	maxTry := 3
	if login != "" && password != "" {
		maxTry = 1 // If all is provided on command line, do not prompt in case of failure
	}
	for i := 0; i < maxTry; i++ {
		login, password = inputCredentials(login, password)
		createTokenResponse := createToken(login, password)
		if createTokenResponse != nil && createTokenResponse.Success {
			tokenBag := &TokenBag{
				Token:      createTokenResponse.Token,
				ClientTTL:  createTokenResponse.ClientTTL,
				LastAccess: time.Now(),
				User:       createTokenResponse.User,
				Authority:  createTokenResponse.Authority,
			}
			save(tokenBag)
			return tokenBag
		}
		_, _ = fmt.Fprintf(os.Stderr, "Invalid login!\n")
		login = ""
		password = ""
	}
	if maxTry > 1 {
		_, _ = fmt.Fprintf(os.Stderr, "Too many failure !!!\n")
	}
	return nil
}

func inputCredentials(login, password string) (string, string) {
	if login == "" {
		_, err := fmt.Fprint(os.Stderr, "Login:")
		if err != nil {
			panic(err)
		}
		r := bufio.NewReader(os.Stdin)
		login, err = r.ReadString('\n')
		if err != nil {
			panic(err)
		}
		login = strings.TrimSpace(login)
	}
	if password == "" {
		password = inputPassword("Password:")
	}
	return login, password
}

func inputPassword(prompt string) string {
	_, err := fmt.Fprint(os.Stderr, prompt)
	if err != nil {
		panic(err)
	}
	bytePassword, err2 := terminal.ReadPassword(syscall.Stdin)
	if err2 != nil {
		panic(err2)
	}
	_, _ = fmt.Fprintf(os.Stderr, "\n")
	return strings.TrimSpace(string(bytePassword))
}

func createToken(login, password string) *proto.TokenCreateResponse {
	tgr := &proto.TokenCreateRequest{
		ClientAuth: config.SkhttpClient.GetClientAuth(),
		Login:      login,
		Password:   password,
	}
	tokenGenerateResponse := &proto.TokenCreateResponse{}
	err := config.SkhttpClient.Do(proto.TokenCreateMeta, tgr, tokenGenerateResponse)
	if err != nil {
		log.Log.Error(err, "error on getToken()")
		return nil
	}
	return tokenGenerateResponse
}
