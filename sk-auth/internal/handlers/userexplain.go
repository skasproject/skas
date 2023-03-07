package handlers

import (
	"fmt"
	"net/http"
	"skas/sk-auth/internal/config"
	"skas/sk-auth/internal/tokenstore"
	"skas/sk-common/pkg/clientauth"
	commonHandlers "skas/sk-common/pkg/httpserver/handlers"
	"skas/sk-common/pkg/skhttp"
	"skas/sk-common/proto/v1/proto"
	"strings"
)

var _ http.Handler = &TokenRenewHandler{}

type UserExplainHandler struct {
	// Server related stuff
	commonHandlers.BaseHandler
	ClientManager clientauth.Manager
	TokenStore    tokenstore.TokenStore
	// Login client related stuff. Also
	LoginClient skhttp.Client
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

func (t UserExplainHandler) getAuthUser(request *http.Request) (*proto.User, error) {
	login, password, ok := request.BasicAuth()
	if ok {
		user, _, err := doLogin(t.LoginClient, login, password)
		if err != nil {
			return nil, err
		}
		return user, nil
	}
	// Try with a skas token
	token := getBearerToken(request)
	if token != "" {
		user, err := t.TokenStore.Get(token)
		if err != nil {
			return nil, err
		}
		return user, nil
	}
	// No authentication in header
	return nil, nil
}

func (t UserExplainHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload = proto.UserExplainRequest{}
	err := requestPayload.FromJson(request.Body)
	if err != nil {
		t.HttpError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	if !t.ClientManager.Validate(&requestPayload.ClientAuth) {
		t.HttpError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}
	user, err := t.getAuthUser(request)
	if err != nil {
		t.HttpError(response, "Server error. Check server logs", http.StatusInternalServerError)
		return
	}
	if user == nil {
		response.Header().Set("WWW-Authenticate", "Basic realm=\"/koo\"")
		t.HttpError(response, "Need to authenticate", http.StatusUnauthorized)
		return
	}
	if !userInGroup(user, config.Conf.AdminGroup) {
		t.HttpError(response, "User has no admin rights", http.StatusUnauthorized)
		return
	}
	explainResponse := &proto.UserExplainResponse{}
	err = t.LoginClient.Do(proto.UserExplainMeta, &requestPayload, explainResponse, nil)
	if err != nil {
		t.HttpError(response, fmt.Sprintf("Error on downside login request: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	t.GetLog().Info("User explain response", "user", explainResponse.Merged.User)
	t.ServeJSON(response, explainResponse)
}
