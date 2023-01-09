package handlers

import (
	"encoding/json"
	"net/http"
	"skas/sk-common/pkg/httpserver"
	"skas/sk-common/proto"
)

type UserStatusHandler struct {
	httpserver.BaseHandler
}

func (h *UserStatusHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload proto.UserStatusRequest
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&requestPayload)
	if err != nil {
		h.HttpError(response, err.Error(), http.StatusBadRequest)
		return
	}
	var userStatus proto.UserStatus

	user := proto.User{}

	var responsePayload *proto.UserStatusResponse
	if userStatus == proto.NotFound || userStatus == proto.PasswordFail {
		responsePayload = &proto.UserStatusResponse{
			UserStatus: userStatus,
		}
	} else {
		responsePayload = &proto.UserStatusResponse{
			User: proto.User{
				Login:       user.Login,
				Uid:         user.Uid,
				CommonNames: user.CommonNames,
				Emails:      user.Emails,
				Groups:      user.Groups,
			},
			UserStatus: userStatus,
		}

	}
	h.ServeJSON(response, responsePayload)
}
