package handlers

import (
	"fmt"
	"github.com/go-logr/logr"
	"net/http"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/skserver"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-merge/internal/clientproviderchain"
)

var _ http.Handler = &LoginHandler{}

var _ skserver.LoggingHandler = &LoginHandler{}

type LoginHandler struct {
	commonHandlers.BaseHandler
	Chain         clientproviderchain.ClientProviderChain
	ClientManager clientauth.Manager
}

func (l LoginHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload proto.LoginRequest
	err := requestPayload.FromJson(request.Body)
	if err != nil {
		l.HttpError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	if !l.ClientManager.Validate(&requestPayload.ClientAuth) {
		l.HttpError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}
	items, err := l.Chain.Scan(requestPayload.Login, requestPayload.Password)
	if err != nil {
		l.HttpError(response, fmt.Sprintf("Providers scan: %v", err), http.StatusInternalServerError)
		return
	}
	merged, authority := clientproviderchain.Merge(requestPayload.Login, items)

	var responsePayload *proto.LoginResponse
	if merged.UserStatus == proto.PasswordChecked {
		responsePayload = &proto.LoginResponse{
			Success: true,
			User: proto.User{
				Login:       merged.Login,
				Uid:         merged.Uid,
				CommonNames: merged.CommonNames,
				Emails:      merged.Emails,
				Groups:      merged.Groups,
			},
			Authority: authority,
		}
	} else {
		responsePayload = &proto.LoginResponse{
			Success: false,
			User: proto.User{
				Login:       merged.Login,
				Uid:         0,
				CommonNames: []string{},
				Emails:      []string{},
				Groups:      []string{},
			},
			Authority: "",
		}
	}
	l.GetLog().Info("User login", "login", requestPayload.Login, "success", responsePayload.Success, "groups", responsePayload.Groups)
	l.ServeJSON(response, responsePayload)
}

// Normally, we should not need to add this, as we embed commonHandlers.BaseHandler which have this function.
// But if w don't, httpserver.LogHttp will not recognize us as a LoggingHandler. May be a compiler bug ?

func (l LoginHandler) GetLog() logr.Logger {
	return l.Logger
}
