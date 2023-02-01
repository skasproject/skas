package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"skas/sk-common/pkg/clientmanager"
	"skas/sk-common/proto/v1/proto"
)

type StatusServerProvider interface {
	GetUserStatus(request proto.UserStatusRequest) (*proto.UserStatusResponse, error)
}

type UserStatusHandler struct {
	BaseHandler
	Provider      StatusServerProvider
	ClientManager clientmanager.ClientManager
}

func (h *UserStatusHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload proto.UserStatusRequest
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&requestPayload)
	if err != nil {
		h.HttpError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	if !h.ClientManager.Validate(&requestPayload.ClientAuth) {
		h.HttpError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}
	responsePayload, err := h.Provider.GetUserStatus(requestPayload)
	if err != nil {
		h.HttpError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	h.GetLog().Info("User status", "login", requestPayload.Login, "status", responsePayload.UserStatus)
	h.ServeJSON(response, responsePayload)
}
