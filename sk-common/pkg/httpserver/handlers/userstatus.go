package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"skas/sk-common/proto"
)

type StatusProvider interface {
	GetUserStatus(request proto.UserStatusRequest) (*proto.UserStatusResponse, error)
}

type UserStatusHandler struct {
	BaseHandler
	Provider StatusProvider
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
	responsePayload, err := h.Provider.GetUserStatus(requestPayload)
	if err != nil {
		h.HttpError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	h.GetLog().Info("User status", "login", requestPayload.Login, "status", responsePayload.UserStatus)
	h.ServeJSON(response, responsePayload)
}
