package handlers

import (
	"fmt"
	"net/http"
	"skas/sk-auth/internal/tokenstore"
	"skas/sk-common/pkg/clientauth"
	commonHandlers "skas/sk-common/pkg/httpserver/handlers"
	"skas/sk-common/pkg/skhttp"
	"skas/sk-common/proto/v1/proto"
)

var _ http.Handler = &TokenRequestHandler{}

type TokenRequestHandler struct {
	// Server related stuff
	commonHandlers.BaseHandler
	ClientManager clientauth.Manager
	TokenStore    tokenstore.TokenStore
	// Login client related stuff
	LoginClient skhttp.Client
}

func (t TokenRequestHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload = proto.TokenRequest{}
	err := requestPayload.FromJson(request.Body)
	if err != nil {
		t.HttpError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	if !t.ClientManager.Validate(&requestPayload.ClientAuth) {
		t.HttpError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}
	user, err := t.login(requestPayload.Login, requestPayload.Password)
	if err != nil {
		t.HttpError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	var responsePayload *proto.TokenResponse
	if user == nil {
		responsePayload = &proto.TokenResponse{
			Success: false,
		}
	} else {
		tokenBag, err := t.TokenStore.NewToken(requestPayload.ClientAuth.Id, *user)
		if err != nil {
			t.HttpError(response, fmt.Sprintf("Error on token creation for login '%s': %s", requestPayload.Login, err.Error()), http.StatusInternalServerError)
			return
		}
		responsePayload = &proto.TokenResponse{
			Success:   true,
			Token:     tokenBag.Token,
			User:      *user,
			ClientTTL: tokenBag.TokenSpec.Lifecycle.ClientTTL.Duration,
		}

	}
	t.GetLog().Info("Token request", "login", requestPayload.Login, "success", responsePayload.Success, "groups", responsePayload.User.Groups)
	t.ServeJSON(response, responsePayload)
}

func (t TokenRequestHandler) login(login, password string) (*proto.User, error) {
	lr := &proto.LoginRequest{
		Login:      login,
		Password:   password,
		ClientAuth: t.LoginClient.GetClientAuth(),
	}
	loginResponse := &proto.LoginResponse{}
	err := t.LoginClient.Do(proto.LoginUrlPath, lr, loginResponse)
	if err != nil {
		return nil, err // Do() return a documented message
	}
	if loginResponse.Success {
		return &loginResponse.User, nil
	} else {
		return nil, nil
	}
}
