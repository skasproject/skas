package handlers

import (
	"fmt"
	"github.com/go-logr/logr"
	"net/http"
	"skas/sk-common/pkg/clientauth"
	commonHandlers "skas/sk-common/pkg/httpserver/handlers"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-merge/internal/clientproviderchain"
)

var _ http.Handler = &PasswordChangeHandler{}

type PasswordChangeHandler struct {
	commonHandlers.BaseHandler
	Chain         clientproviderchain.ClientProviderChain
	ClientManager clientauth.Manager
}

func (p PasswordChangeHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload = proto.PasswordChangeRequest{}
	err := requestPayload.FromJson(request.Body)
	if err != nil {
		p.HttpError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	if !p.ClientManager.Validate(&requestPayload.ClientAuth) {
		p.HttpError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}
	changePasswordResponse, err := p.Chain.ChangePassword(&requestPayload)
	if err != nil {
		p.HttpError(response, fmt.Sprintf("Providers change password: %v", err), http.StatusInternalServerError)
		return

	}
	p.ServeJSON(response, changePasswordResponse)
	return
}

// Normally, we should not need to add this, as we embed commonHandlers.BaseHandler which have this function.
// But if w don't, httpserver.LogHttp will not recognize us as a LoggingHandler. May be a compiler bug ?

func (p PasswordChangeHandler) GetLog() logr.Logger {
	return p.Logger
}
