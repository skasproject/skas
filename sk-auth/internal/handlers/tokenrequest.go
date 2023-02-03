package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"skas/sk-auth/internal/tokenstore"
	"skas/sk-common/pkg/client"
	commonHandlers "skas/sk-common/pkg/httpserver/handlers"
	"skas/sk-common/proto/v1/proto"
)

var _ http.Handler = &TokenRequestHandler{}

type TokenRequestHandler struct {
	commonHandlers.BaseHandler
	ClientManager client.Manager
	TokenStore    tokenstore.TokenStore
	HttpClient    *http.Client
}

func (t TokenRequestHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload proto.LoginRequest
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&requestPayload)
	if err != nil {
		t.HttpError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	if !t.ClientManager.Validate(&requestPayload.ClientAuth) {
		t.HttpError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}

}

func (t TokenRequestHandler) login(login, password string) (proto.User, error) {

}
