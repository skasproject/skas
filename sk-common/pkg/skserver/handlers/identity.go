package handlers

import (
	"fmt"
	"github.com/go-logr/logr"
	"net/http"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/misc"
	"skas/sk-common/pkg/skserver/protector"
	"skas/sk-common/proto/v1/proto"
)

type IdentityGetter interface {
	// GetIdentity - We pass request by value, as we may modify it
	GetIdentity(request proto.IdentityRequest) (*proto.IdentityResponse, misc.HttpError)
}

type HttpRequestValidator interface {
	Validate(request *http.Request, response http.ResponseWriter) misc.HttpError
}

type IdentityHandler struct {
	BaseHandler
	IdentityGetter       IdentityGetter
	ClientManager        clientauth.Manager
	HttpRequestValidator HttpRequestValidator
	Protector            protector.Protector
}

func (h *IdentityHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload proto.IdentityRequest
	err := requestPayload.FromJson(request.Body)
	if err != nil {
		h.HttpSendError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	id, locked := h.Protector.Entry(requestPayload.Login)
	defer h.Protector.Exit(id, requestPayload.Login)
	if locked {
		h.HttpSendError(response, "??", http.StatusServiceUnavailable)
		return
	}
	if !h.ClientManager.Validate(&requestPayload.ClientAuth) {
		h.HttpSendError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}
	if h.HttpRequestValidator != nil {
		httpError := h.HttpRequestValidator.Validate(request, response)
		if httpError != nil {
			h.HttpSendError(response, httpError.Error(), httpError.GetStatusCode())
			return
		}
	}
	responsePayload, httpError := h.IdentityGetter.GetIdentity(requestPayload)
	if httpError != nil {
		h.HttpSendError(response, httpError.Error(), httpError.GetStatusCode())
		return
	}
	// Failure cases are:
	// - PasswordFail if password is provided on request
	// - NotFound in all cases
	if (requestPayload.Password != "" && responsePayload.Status == proto.PasswordFail) || (responsePayload.Status == proto.NotFound) {
		h.Protector.Failure(id, responsePayload.Login)
	} else {
		h.Protector.Success(id, responsePayload.Login)
	}
	h.GetLog().Info("User status", "login", requestPayload.Login, "status", responsePayload.Status)
	h.ServeJSON(response, responsePayload)
}

// Normally, we should not need to add this, as we embed commonHandlers.BaseHandler which have this function.
// But if we don't, httpserver.LogHttp will not recognize us as a LoggingHandler. May be a compiler bug ?

func (h *IdentityHandler) GetLog() logr.Logger {
	return h.Logger
}

func (h *IdentityHandler) SetLog(logger logr.Logger) {
	h.Logger = logger
}
