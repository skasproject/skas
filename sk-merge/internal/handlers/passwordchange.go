package handlers

import (
	"fmt"
	"github.com/go-logr/logr"
	"net/http"
	"skas/sk-common/pkg/clientauth"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-merge/internal/providerchain"
)

var _ http.Handler = &PasswordChangeHandler{}

type PasswordChangeHandler struct {
	commonHandlers.BaseHandler
	Chain         providerchain.ProviderChain
	ClientManager clientauth.Manager
}

func (p *PasswordChangeHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload = proto.PasswordChangeRequest{}
	err := requestPayload.FromJson(request.Body)
	if err != nil {
		p.HttpSendError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	if !p.ClientManager.Validate(&requestPayload.ClientAuth) {
		p.HttpSendError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}
	changePasswordResponse, err := p.Chain.ChangePassword(requestPayload)
	if err != nil {
		p.HttpSendError(response, fmt.Sprintf("Providers change password: %v", err), http.StatusBadGateway)
		return
	}
	p.ServeJSON(response, changePasswordResponse)
	return
}

// Normally, we should not need to add this, as we embed commonHandlers.BaseHandler which have this function.
// But if w don't, httpserver.LogHttp will not recognize us as a LoggingHandler. May be a compiler bug ?

func (p *PasswordChangeHandler) GetLog() logr.Logger {
	return p.Logger
}

func (p *PasswordChangeHandler) SetLog(logger logr.Logger) {
	p.Logger = logger
}
