package handlers

import (
	"fmt"
	"github.com/go-logr/logr"
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

func (k *KubeconfigHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload = proto.KubeconfigRequest{}
	err := requestPayload.FromJson(request.Body)
	if err != nil {
		k.HttpSendError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	if !k.ClientManager.Validate(&requestPayload.ClientAuth) {
		k.HttpSendError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}
	var responsePayload = &proto.KubeconfigResponse{
		KubeconfigConfig: config.Conf.Kubeconfig,
	}
	k.GetLog().Info("Kubeconfig request")
	k.ServeJSON(response, responsePayload)
}

// Normally, we should not need to add this, as we embed commonHandlers.BaseHandler which have this function.
// But if w don't, httpserver.LogHttp will not recognize us as a LoggingHandler. May be a compiler bug ?

func (k *KubeconfigHandler) GetLog() logr.Logger {
	return k.Logger
}

func (k *KubeconfigHandler) SetLog(logger logr.Logger) {
	k.Logger = logger
}
