package handlers

import (
	"fmt"
	"github.com/go-logr/logr"
	"net/http"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/skserver"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
)

var _ http.Handler = &LoginHandler{}

var _ skserver.LoggingHandler = &LoginHandler{}

type LoginHandler struct {
	commonHandlers.BaseHandler
	ClientManager  clientauth.Manager
	IdentityGetter commonHandlers.IdentityGetter
}

func (l *LoginHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload proto.LoginRequest
	err := requestPayload.FromJson(request.Body)
	if err != nil {
		l.HttpSendError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	if !l.ClientManager.Validate(&requestPayload.ClientAuth) {
		l.HttpSendError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}
	user, _, err := doLogin(l.IdentityGetter, requestPayload.Login, requestPayload.Password)
	if err != nil {
		l.HttpSendError(response, fmt.Sprintf("Error on downside login request: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	var responsePayload *proto.LoginResponse
	if user == nil {
		responsePayload = &proto.LoginResponse{
			Success: false,
			User:    proto.InitUser(requestPayload.Login),
		}

	} else {
		responsePayload = &proto.LoginResponse{
			Success: true,
			User:    *user,
		}
	}
	l.GetLog().Info("User login", "login", requestPayload.Login, "success", responsePayload.Success, "groups", responsePayload.Groups)
	l.ServeJSON(response, responsePayload)
}

// Normally, we should not need to add this, as we embed commonHandlers.BaseHandler which have this function.
// But if we don't, httpserver.LogHttp will not recognize us as a LoggingHandler. May be a compiler bug ?

func (l *LoginHandler) GetLog() logr.Logger {
	return l.Logger
}

func (l *LoginHandler) SetLog(logger logr.Logger) {
	l.Logger = logger
}
