package handlers

import (
	"fmt"
	"github.com/go-logr/logr"
	"net/http"
	"skas/sk-auth/internal/config"
	"skas/sk-auth/internal/tokenstore"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/skclient"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
	"strings"
)

var _ http.Handler = &TokenRenewHandler{}

type UserDescribeHandler struct {
	// Server related stuff
	commonHandlers.BaseHandler
	ClientManager clientauth.Manager
	TokenStore    tokenstore.TokenStore
	// Login client related stuff. Also
	Provider skclient.SkClient
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
	fmt.Printf("usenInGroup() user:%v  groups:%s", user, group)
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

func (u *UserDescribeHandler) getAuthUser(request *http.Request) (*proto.User, error) {
	login, password, ok := request.BasicAuth()
	if ok {
		user, _, err := doLogin(u.Provider, login, password)
		if err != nil {
			return nil, err
		}
		return user, nil
	}
	// Try with a skas token
	token := getBearerToken(request)
	if token != "" {
		user, err := u.TokenStore.Get(token)
		if err != nil {
			return nil, err
		}
		return user, nil
	}
	// No authentication in header
	return nil, nil
}

func (u *UserDescribeHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload = proto.UserDescribeRequest{}
	err := requestPayload.FromJson(request.Body)
	if err != nil {
		u.HttpError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	if !u.ClientManager.Validate(&requestPayload.ClientAuth) {
		u.HttpError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}
	user, err := u.getAuthUser(request)
	if err != nil {
		u.HttpError(response, "Server error. Check server logs", http.StatusInternalServerError)
		return
	}
	if user == nil {
		response.Header().Set("WWW-Authenticate", "Basic realm=\"/koo\"")
		u.HttpError(response, "Need to authenticate", http.StatusUnauthorized)
		return
	}
	if !userInGroup(user, config.Conf.AdminGroup) {
		u.HttpError(response, "User has no admin rights", http.StatusUnauthorized)
		return
	}
	describeResponse := &proto.UserDescribeResponse{}
	err = u.Provider.Do(proto.UserDescribeMeta, &requestPayload, describeResponse, nil)
	if err != nil {
		u.HttpError(response, fmt.Sprintf("Error on downside login request: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	u.GetLog().Info("User describe response", "user", describeResponse.Merged.User)
	u.ServeJSON(response, describeResponse)
}

// Normally, we should not need to add this, as we embed commonHandlers.BaseHandler which have this function.
// But if w don't, httpserver.LogHttp will not recognize us as a LoggingHandler. May be a compiler bug ?

func (u *UserDescribeHandler) GetLog() logr.Logger {
	return u.Logger
}

func (u *UserDescribeHandler) SetLog(logger logr.Logger) {
	u.Logger = logger
}
