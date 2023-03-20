package handlers

import (
	"fmt"
	"github.com/go-logr/logr"
	"net/http"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/proto/v1/proto"
)

type IdentityServerProvider interface {
	GetUserIdentity(request proto.UserIdentityRequest) (*proto.UserIdentityResponse, error)
}

type UserIdentityHandler struct {
	BaseHandler
	Provider      IdentityServerProvider
	ClientManager clientauth.Manager
}

func (h *UserIdentityHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload proto.UserIdentityRequest
	err := requestPayload.FromJson(request.Body)
	if err != nil {
		h.HttpError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	if !h.ClientManager.Validate(&requestPayload.ClientAuth) {
		h.HttpError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}
	responsePayload, err := h.Provider.GetUserIdentity(requestPayload)
	if err != nil {
		h.HttpError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	h.GetLog().Info("User status", "login", requestPayload.Login, "status", responsePayload.UserStatus)
	h.ServeJSON(response, responsePayload)
}

// Normally, we should not need to add this, as we embed commonHandlers.BaseHandler which have this function.
// But if we don't, httpserver.LogHttp will not recognize us as a LoggingHandler. May be a compiler bug ?

func (h *UserIdentityHandler) GetLog() logr.Logger {
	return h.Logger
}

func (h *UserIdentityHandler) SetLog(logger logr.Logger) {
	h.Logger = logger
}
