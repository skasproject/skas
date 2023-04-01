package handlers

import (
	"net/http"
	"skas/sk-auth/internal/config"
	"skas/sk-auth/internal/tokenstore"
	"skas/sk-common/pkg/misc"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
	"strings"
)

var _ commonHandlers.HttpRequestValidator = &AdminHttpRequestValidator{}

type AdminHttpRequestValidator struct {
	TokenStore     tokenstore.TokenStore
	IdentityGetter commonHandlers.IdentityGetter
}

func (a *AdminHttpRequestValidator) Validate(request *http.Request, response http.ResponseWriter) misc.HttpError {
	user, err := a.getAuthUser(request)
	if err != nil {
		return misc.NewHttpError("Server error. Check server logs", http.StatusInternalServerError)
	}
	if user == nil {
		response.Header().Set("WWW-Authenticate", "Basic realm=\"/skas\"")
		return misc.NewHttpError("Need to authenticate", http.StatusUnauthorized)
	}
	if !userInGroup(user, config.Conf.AdminGroup) {
		return misc.NewHttpError("User has no admin rights", http.StatusUnauthorized)
	}
	return nil
}

func getBearerToken(request *http.Request) string {
	authList, ok := request.Header["Authorization"]
	if !ok {
		return ""
	}
	for _, auth := range authList {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimSpace(auth[len("Bearer "):])
		}
	}
	return ""
}

func userInGroup(user *proto.User, group string) bool {
	//fmt.Printf("usenInGroup() user:%v  groups:%s", user, group)
	if user.Groups == nil {
		return false
	}
	for _, grp := range user.Groups {
		if grp == group {
			return true
		}
	}
	return false
}

func (a *AdminHttpRequestValidator) getAuthUser(request *http.Request) (*proto.User, error) {
	login, password, ok := request.BasicAuth()
	if ok {
		user, _, err := doLogin(a.IdentityGetter, login, password)
		if err != nil {
			return nil, err
		}
		return user, nil
	}
	// Try with a skas token
	token := getBearerToken(request)
	if token != "" {
		user, err := a.TokenStore.Get(token)
		if err != nil {
			return nil, err
		}
		return user, nil
	}
	// No authentication in header
	return nil, nil
}
