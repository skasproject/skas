package handlers

import (
	"fmt"
	"github.com/go-logr/logr"
	"net/http"
	"skas/sk-auth/internal/tokenstore"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/misc"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/pkg/skserver/protector"
	"skas/sk-common/proto/v1/proto"
)

var _ http.Handler = &TokenRenewHandler{}

type TokenRenewHandler struct {
	// Server related stuff
	commonHandlers.BaseHandler
	ClientManager clientauth.Manager
	TokenStore    tokenstore.TokenStore
	Protector     protector.TokenProtector
}

func (t *TokenRenewHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload = proto.TokenRenewRequest{}
	err := requestPayload.FromJson(request.Body)
	if err != nil {
		t.HttpSendError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	locked := t.Protector.EntryForToken()
	if locked {
		t.HttpSendError(response, "Locked", http.StatusServiceUnavailable)
		return
	}
	if !t.ClientManager.Validate(&requestPayload.ClientAuth) {
		t.HttpSendError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}
	user, err := t.TokenStore.Get(requestPayload.Token)
	if err != nil {
		t.HttpSendError(response, fmt.Sprintf("Error while retreiving token in the store: %v", err.Error()), http.StatusUnauthorized)
		return
	}
	if user == nil {
		t.Protector.TokenNotFound()
	}
	responsePayload := &proto.TokenRenewResponse{
		Token: requestPayload.Token,
		Valid: user != nil,
	}
	t.GetLog().Info("Token renew", "token", misc.ShortenString(requestPayload.Token), "valid", responsePayload.Valid)
	t.ServeJSON(response, responsePayload)
}

// Normally, we should not need to add this, as we embed commonHandlers.BaseHandler which have this function.
// But if we don't, httpserver.LogHttp will not recognize us as a LoggingHandler. May be a compiler bug ?

func (t *TokenRenewHandler) GetLog() logr.Logger {
	return t.Logger
}

func (t *TokenRenewHandler) SetLog(logger logr.Logger) {
	t.Logger = logger
}
