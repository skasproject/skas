package handlers

import (
	"fmt"
	"net/http"
	"skas/sk-auth/internal/tokenstore"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/skclient"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
)

var _ http.Handler = &TokenCreateHandler{}

type TokenCreateHandler struct {
	// Server related stuff
	commonHandlers.BaseHandler
	ClientManager clientauth.Manager
	TokenStore    tokenstore.TokenStore
	// Login client related stuff
	Provider skclient.SkClient
}

func (t TokenCreateHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload = proto.TokenCreateRequest{}
	err := requestPayload.FromJson(request.Body)
	if err != nil {
		t.HttpError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	if !t.ClientManager.Validate(&requestPayload.ClientAuth) {
		t.HttpError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}
	user, authority, err := doLogin(t.Provider, requestPayload.Login, requestPayload.Password)
	if err != nil {
		t.HttpError(response, fmt.Sprintf("Error on downside login request: %s", err.Error()), http.StatusInternalServerError)
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
			t.HttpError(response, fmt.Sprintf("Error on token creation for login '%s': %s", requestPayload.Login, err.Error()), http.StatusInternalServerError)
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
