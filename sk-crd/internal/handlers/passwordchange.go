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
	locked := p.Protector.EntryForLogin(requestPayload.Login)
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
		p.Protector.ProtectLoginResult("", proto.UserNotFound)
		p.Logger.V(0).Info("User not found", "user", requestPayload.Login)
		responsePayload.Status = proto.UserNotFound
		p.ServeJSON(response, responsePayload)
		return
	}
	// Check provided oldPassword
	err = bcrypt.CompareHashAndPassword([]byte(usr.Spec.PasswordHash), []byte(requestPayload.OldPassword))
	if err != nil {
		p.Protector.ProtectLoginResult(responsePayload.Login, proto.InvalidOldPassword)
		p.Logger.V(0).Info("Invalid old password", "user", requestPayload.Login)
		responsePayload.Status = proto.InvalidOldPassword
		p.ServeJSON(response, responsePayload)
		return
	}
	if requestPayload.NewPasswordHash == "" {
		// We are called from a 0.2.0 client
		p.HttpSendError(response, "Protocol error. Please, update your kubectl-sk client", http.StatusBadRequest)
		return
	}
	// Test if NewPasswordHash look like a hash
	err = bcrypt.CompareHashAndPassword([]byte(requestPayload.NewPasswordHash), []byte("xxxxx"))
	if err != nil && err != bcrypt.ErrMismatchedHashAndPassword {
		p.HttpSendError(response, "Protocol error. Invalid password hash", http.StatusBadRequest)
		return
	}

	usr.Spec.PasswordHash = requestPayload.NewPasswordHash
	err = p.KubeClient.Update(context.Background(), usr)
	if err != nil {
		p.HttpSendError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	p.Logger.V(0).Info("Password changed", "user", requestPayload.Login)
	responsePayload.Status = proto.PasswordChanged
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
