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

var _ http.Handler = &TokenGenerateHandler{}

type TokenGenerateHandler struct {
	// Server related stuff
	commonHandlers.BaseHandler
	ClientManager clientauth.Manager
	TokenStore    tokenstore.TokenStore
	// Login client related stuff
	LoginClient skhttp.Client
}

func (t TokenGenerateHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload = proto.TokenGenerateRequest{}
	err := requestPayload.FromJson(request.Body)
	if err != nil {
		t.HttpError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	if !t.ClientManager.Validate(&requestPayload.ClientAuth) {
		t.HttpError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}
	user, authority, err := t.login(requestPayload.Login, requestPayload.Password)
	if err != nil {
		t.HttpError(response, fmt.Sprintf("Error on downside login request: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	var responsePayload *proto.TokenGenerateResponse
	if user == nil {
		responsePayload = &proto.TokenGenerateResponse{
			Success: false,
		}
	} else {
		token, err := t.TokenStore.NewToken(requestPayload.ClientAuth.Id, *user, authority)
		if err != nil {
			t.HttpError(response, fmt.Sprintf("Error on token creation for login '%s': %s", requestPayload.Login, err.Error()), http.StatusInternalServerError)
			return
		}
		responsePayload = &proto.TokenGenerateResponse{
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

func (t TokenGenerateHandler) login(login, password string) (*proto.User /*authority*/, string, error) {
	lr := &proto.LoginRequest{
		Login:      login,
		Password:   password,
		ClientAuth: t.LoginClient.GetClientAuth(),
	}
	loginResponse := &proto.LoginResponse{}
	err := t.LoginClient.Do(proto.LoginMeta, lr, loginResponse)
	if err != nil {
		return nil, "", fmt.Errorf("error on exchange on %s: %w", proto.LoginMeta.UrlPath, err) // Do() return a documented message
	}
	if loginResponse.Success {
		return &loginResponse.User, loginResponse.Authority, nil
	} else {
		return nil, "", nil
	}
}
