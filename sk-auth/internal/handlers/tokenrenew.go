package handlers

import (
	"fmt"
	"net/http"
	"skas/sk-auth/internal/tokenstore"
	"skas/sk-common/pkg/clientauth"
	commonHandlers "skas/sk-common/pkg/httpserver/handlers"
	"skas/sk-common/pkg/misc"
	"skas/sk-common/proto/v1/proto"
)

var _ http.Handler = &TokenRenewHandler{}

type TokenRenewHandler struct {
	// Server related stuff
	commonHandlers.BaseHandler
	ClientManager clientauth.Manager
	TokenStore    tokenstore.TokenStore
}

func (t TokenRenewHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload = proto.TokenRenewRequest{}
	err := requestPayload.FromJson(request.Body)
	if err != nil {
		t.HttpError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	if !t.ClientManager.Validate(&requestPayload.ClientAuth) {
		t.HttpError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}

	user, err := t.TokenStore.Get(requestPayload.Token)
	if err != nil {
		t.HttpError(response, fmt.Sprintf("Error while retreiving token in the store: %v", err.Error()), http.StatusUnauthorized)
		return
	}
	responsePayload := &proto.TokenRenewResponse{
		Token: requestPayload.Token,
		Valid: user != nil,
	}
	t.GetLog().Info("Token renew", "token", misc.ShortenString(requestPayload.Token), "valid", responsePayload.Valid)
	t.ServeJSON(response, responsePayload)
}
