package handlers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	userdbv1alpha1 "skas/sk-common/k8sapis/userdb/v1alpha1"
	"skas/sk-common/pkg/clientauth"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/pkg/skserver/protector"
	"skas/sk-common/proto/v1/proto"
)

var _ http.Handler = &PasswordChangeHandler{}

type PasswordChangeHandler struct {
	commonHandlers.BaseHandler
	ClientManager clientauth.Manager
	KubeClient    client.Client
	Namespace     string
	Protector     protector.Protector
}

func (p *PasswordChangeHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload = proto.PasswordChangeRequest{}
	err := requestPayload.FromJson(request.Body)
	if err != nil {
		p.HttpSendError(response, fmt.Sprintf("Payload decoding: %v", err), http.StatusBadRequest)
		return
	}
	locked := p.Protector.Entry(requestPayload.Login)
	if locked {
		p.HttpSendError(response, "Locked", http.StatusServiceUnavailable)
		return
	}
	if !p.ClientManager.Validate(&requestPayload.ClientAuth) {
		p.HttpSendError(response, "Client authentication failed", http.StatusUnauthorized)
		return
	}
	responsePayload := &proto.PasswordChangeResponse{
		Login: requestPayload.Login,
	}
	// Try to fetch user
	usr := &userdbv1alpha1.User{}
	err = p.KubeClient.Get(context.TODO(), client.ObjectKey{
		Namespace: p.Namespace,
		Name:      requestPayload.Login,
	}, usr)
	if client.IgnoreNotFound(err) != nil {
		p.HttpSendError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if err != nil {
		p.Protector.Failure(responsePayload.Login)
		p.Logger.V(0).Info("User not found", "user", requestPayload.Login)
		responsePayload.Status = proto.UnknownUser
		p.ServeJSON(response, responsePayload)
		return
	}
	// Check provided oldPassword
	err = bcrypt.CompareHashAndPassword([]byte(usr.Spec.PasswordHash), []byte(requestPayload.OldPassword))
	if err != nil {
		p.Protector.Failure(responsePayload.Login)
		p.Logger.V(0).Info("Invalid old password", "user", requestPayload.Login)
		responsePayload.Status = proto.InvalidOldPassword
		p.ServeJSON(response, responsePayload)
		return
	}
	// Check provided new password
	if len(requestPayload.NewPassword) < 3 {
		p.Logger.V(0).Info("Invalid new password", "user", requestPayload.Login)
		responsePayload.Status = proto.InvalidNewPassword
		p.ServeJSON(response, responsePayload)
		return
	}
	// Create new password
	hash, err := bcrypt.GenerateFromPassword([]byte(requestPayload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		p.HttpSendError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	usr.Spec.PasswordHash = string(hash)
	err = p.KubeClient.Update(context.Background(), usr)
	if err != nil {
		p.HttpSendError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	p.Logger.V(0).Info("Password changed", "user", requestPayload.Login)
	responsePayload.Status = proto.Done
	p.ServeJSON(response, responsePayload)
	return
}

// Normally, we should not need to add this, as we embed commonHandlers.BaseHandler which have this function.
// But if we don't, httpserver.LogHttp will not recognize us as a LoggingHandler. May be a compiler bug ?

func (p *PasswordChangeHandler) GetLog() logr.Logger {
	return p.Logger
}

func (p *PasswordChangeHandler) SetLog(logger logr.Logger) {
	p.Logger = logger
}
