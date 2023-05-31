package handlers

import (
	"fmt"
	"github.com/go-logr/logr"
	"net/http"
	"skas/sk-auth/internal/tokenstore"
	"skas/sk-common/pkg/clientauth"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/pkg/skserver/protector"
	"skas/sk-common/proto/v1/proto"
)

var _ http.Handler = &TokenCreateHandler{}

type TokenCreateHandler struct {
	// Server related stuff
	commonHandlers.BaseHandler
	ClientManager  clientauth.Manager
	TokenStore     tokenstore.TokenStore
	IdentityGetter commonHandlers.IdentityGetter
	Protector      protector.LoginProtector
}

func (t *TokenCreateHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload = proto.TokenCreateRequest{}
	err := requestPayload.FromJson(request.Body)
	if err != nil {
		t.HttpSendError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	locked := t.Protector.EntryForLogin(requestPayload.Login)
	if locked {
		t.HttpSendError(response, "Locked", http.StatusServiceUnavailable)
		return
	}
	if !t.ClientManager.Validate(&requestPayload.ClientAuth) {
		t.HttpSendError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}
	user, authority, err := doLogin(t.IdentityGetter, requestPayload.Login, requestPayload.Password, t.Protector)
	if err != nil {
		t.HttpSendError(response, fmt.Sprintf("Error on downside login request: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	var responsePayload *proto.TokenCreateResponse
	if user == nil {
		responsePayload = &proto.TokenCreateResponse{
			Success: false,
		}
	} else {
		token, err := t.TokenStore.NewToken(requestPayload.ClientAuth.Id, *user, authority)
		if err != nil {
			t.HttpSendError(response, fmt.Sprintf("Error on token creation for login '%s': %s", requestPayload.Login, err.Error()), http.StatusInternalServerError)
			return
		}
		responsePayload = &proto.TokenCreateResponse{
			Success:   true,
			Token:     token,
			User:      *user,
			ClientTTL: t.TokenStore.GetClientTtl(),
			Authority: authority,
		}

	}
	t.GetLog().Info("Token request", "login", requestPayload.Login, "success", responsePayload.Success, "groups", responsePayload.User.Groups)
	t.ServeJSON(response, responsePayload)
}

// Normally, we should not need to add this, as we embed commonHandlers.BaseHandler which have this function.
// But if we don't, httpserver.LogHttp will not recognize us as a LoggingHandler. May be a compiler bug ?

func (t *TokenCreateHandler) GetLog() logr.Logger {
	return t.Logger
}

func (t *TokenCreateHandler) SetLog(logger logr.Logger) {
	t.Logger = logger
}
