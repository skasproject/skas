package handlers

import (
	"fmt"
	"net/http"
	"skas/sk-auth/internal/config"
	"skas/sk-common/pkg/clientauth"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
)

var _ http.Handler = &KubeconfigHandler{}

type KubeconfigHandler struct {
	commonHandlers.BaseHandler
	ClientManager clientauth.Manager
}

func (k KubeconfigHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload = proto.KubeconfigRequest{}
	err := requestPayload.FromJson(request.Body)
	if err != nil {
		k.HttpError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	if !k.ClientManager.Validate(&requestPayload.ClientAuth) {
		k.HttpError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}
	var responsePayload = &proto.KubeconfigResponse{
		KubeconfigConfig: config.Conf.Kubeconfig,
	}
	k.GetLog().Info("Kubeconfig request")
	k.ServeJSON(response, responsePayload)
}
