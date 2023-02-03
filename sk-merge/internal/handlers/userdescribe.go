package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"net/http"
	"skas/sk-common/pkg/clientauth"
	"skas/sk-common/pkg/httpserver"
	commonHandlers "skas/sk-common/pkg/httpserver/handlers"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-merge/internal/clientproviderchain"
)

var _ http.Handler = &UserDescribeHandler{}

var _ httpserver.LoggingHandler = &UserDescribeHandler{}

type UserDescribeHandler struct {
	commonHandlers.BaseHandler
	Chain         clientproviderchain.ClientProviderChain
	ClientManager clientauth.Manager
}

func (u UserDescribeHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload proto.UserDescribeRequest
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&requestPayload)
	if err != nil {
		u.HttpError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	if !u.ClientManager.Validate(&requestPayload.ClientAuth) {
		u.HttpError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}
	items, err := u.Chain.Scan(requestPayload.Login, requestPayload.Password)
	if err != nil {
		u.HttpError(response, fmt.Sprintf("Providers scan: %v", err), http.StatusInternalServerError)
		return
	}
	merged, credentialAuthorityProvider := clientproviderchain.Merge(requestPayload.Login, items)

	responsePayload := &proto.UserDescribeResponse{
		Items:                       make([]proto.UserDescribeItem, 0, u.Chain.GetLength()),
		Merged:                      *merged,
		CredentialAuthorityProvider: credentialAuthorityProvider,
	}
	for idx, _ := range items {
		udi := &proto.UserDescribeItem{
			UserStatusResponse: *items[idx].UserStatusResponse,
			Translated:         *items[idx].Translated,
		}
		udi.Provider.Name = (*items[idx].Provider).GetName()
		udi.Provider.CredentialAuthority = (*items[idx].Provider).IsCredentialAuthority()
		udi.Provider.GroupAuthority = (*items[idx].Provider).IsGroupAuthority()
		responsePayload.Items = append(responsePayload.Items, *udi)
	}
	u.GetLog().Info("User describe", "login", requestPayload.Login, "status", responsePayload.Merged.UserStatus)
	u.ServeJSON(response, responsePayload)
}

// Normally, we should not need to add this, as we embed commonHandlers.BaseHandler which have this function.
// But if w don't, httpserver.LogHttp will not recognize us as a LoggingHandler. May be a compiler bug ?

func (u UserDescribeHandler) GetLog() logr.Logger {
	return u.Logger
}
