package handlers

import (
	"fmt"
	"net/http"
	"skas/sk-auth/internal/config"
	"skas/sk-auth/internal/passwordvalidator"
	"skas/sk-common/pkg/clientauth"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
)

var _ http.Handler = &PasswordStrengthHandler{}

type PasswordStrengthHandler struct {
	commonHandlers.BaseHandler
	ClientManager clientauth.Manager
}

func (p PasswordStrengthHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload = proto.PasswordStrengthRequest{}
	err := requestPayload.FromJson(request.Body)
	if err != nil {
		p.HttpSendError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	if !p.ClientManager.Validate(&requestPayload.ClientAuth) {
		p.HttpSendError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}

	score, isCommon := passwordvalidator.Validate(requestPayload.Password, requestPayload.UserInputs)
	acceptable := score >= config.Conf.PasswordStrength.MinimumScore && !(config.Conf.PasswordStrength.ForbidCommon && isCommon)

	responsePayload := &proto.PasswordStrengthResponse{
		Password:   requestPayload.Password,
		Score:      score,
		IsCommon:   isCommon,
		Acceptable: acceptable,
	}

	p.Logger.V(0).Info("Password strength", "score", score, "isCommon", isCommon, "acceptable", acceptable)
	p.ServeJSON(response, responsePayload)
	return

}
