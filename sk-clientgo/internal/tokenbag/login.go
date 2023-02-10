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
		generateTokenResponse := generateToken(login, password)
		if generateTokenResponse != nil && generateTokenResponse.Success {
			tokenBag := &TokenBag{
				Token:      generateTokenResponse.Token,
				ClientTTL:  generateTokenResponse.ClientTTL,
				LastAccess: time.Now(),
				User:       generateTokenResponse.User,
				Authority:  generateTokenResponse.Authority,
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

func generateToken(login, password string) *proto.TokenGenerateResponse {
	tgr := &proto.TokenGenerateRequest{
		ClientAuth: config.SkhttpClient.GetClientAuth(),
		Login:      login,
		Password:   password,
	}
	tokenGenerateResponse := &proto.TokenGenerateResponse{}
	err := config.SkhttpClient.Do(proto.TokenGenerateMeta, tgr, tokenGenerateResponse)
	if err != nil {
		log.Log.Error(err, "error on getToken()")
		return nil
	}
	return tokenGenerateResponse
}
