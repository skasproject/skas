package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"net/http"
	"skas/sk-common/pkg/httpserver"
	commonHandlers "skas/sk-common/pkg/httpserver/handlers"
	"skas/sk-common/proto"
	"skas/sk-merge/internal/clientproviderchain"
)

var _ http.Handler = &UserDescribeHandler{}

var _ httpserver.LoggingHandler = &UserDescribeHandler{}

type UserDescribeHandler struct {
	commonHandlers.BaseHandler
	Chain clientproviderchain.ClientProviderChain
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
	items, err := u.Chain.Scan(requestPayload.Login, requestPayload.Password)
	if err != nil {
		u.HttpError(response, fmt.Sprintf("Providers scan: %v", err), http.StatusInternalServerError)
		return
	}
	responsePayload := &proto.UserDescribeResponse{
		Items: make([]proto.UserDescribeItem, 0, u.Chain.GetLength()),
	}
	for idx, _ := range items {
		fmt.Printf("**************** provider name: %s\n", (*items[idx].Provider).GetName())
		udi := &proto.UserDescribeItem{
			UserStatusResponse: *items[idx].UserStatusResponse,
			ProviderName:       (*items[idx].Provider).GetName(),
			Authority:          (*items[idx].Provider).IsAuthority(),
		}
		responsePayload.Items = append(responsePayload.Items, *udi)

		//responsePayload.Items = append(responsePayload.Items, proto.UserDescribeItem{
		//	UserStatusResponse: *items[idx].UserStatusResponse,
		//	ProviderName:       (*items[idx].Provider).GetName(),
		//	Authority:          (*items[idx].Provider).IsAuthority(),
		//})
	}
	u.GetLog().Info("User describe", "login", requestPayload.Login)
	u.ServeJSON(response, responsePayload)
}

// Normally, we should not need to add this, as we embed commonHandlers.BaseHandler which have this function.
// But if w don't, httpserver.LogHttp will not recognize us as a LoggingHandler. May be a compiler bug ?

func (u UserDescribeHandler) GetLog() logr.Logger {
	return u.Logger
}
