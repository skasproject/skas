package handlers

import (
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"skas/sk-common/pkg/httpserver"
	"skas/sk-common/proto"
	"skas/sk-static/internal/config"
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
	user, ok := config.Config.UserByLogin[requestPayload.Login]
	if !ok {
		userStatus = proto.NotFound
	} else {
		if requestPayload.Password == "" || user.PasswordHash == "" {
			userStatus = proto.PasswordUnchecked
		} else {
			err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(requestPayload.Password))
			if err == nil {
				userStatus = proto.PasswordChecked
			} else {
				userStatus = proto.PasswordFail
			}
		}
	}
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
