package crdprovider

import (
	"context"
	"github.com/go-logr/logr"
	"golang.org/x/crypto/bcrypt"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"skas/sk-common/pkg/httpserver/handlers"
	"skas/sk-common/proto"
	userdbv1alpha1 "skas/sk-crd/k8sapis/userdb/v1alpha1"
)

var _ handlers.StatusProvider = &crdProvider{}

type crdProvider struct {
	kubeClient client.Client
	namespace  string
	logger     logr.Logger
}

func New(kubeClient client.Client, namespace string, logger logr.Logger) handlers.StatusProvider {
	return &crdProvider{
		kubeClient: kubeClient,
		namespace:  namespace,
		logger:     logger,
	}
}

func (p crdProvider) GetUserStatus(request proto.UserStatusRequest) (*proto.UserStatusResponse, error) {
	responsePayload := &proto.UserStatusResponse{
		UserStatus: proto.NotFound,
		User: &proto.User{
			Login:       request.Login,
			Emails:      []string{},
			CommonNames: []string{},
			Groups:      []string{},
		},
	}
	// ------------------- Handle groups (Even if notFound)
	list := userdbv1alpha1.GroupBindingList{}
	err := p.kubeClient.List(context.TODO(), &list, client.MatchingFields{"userkey": request.Login}, client.InNamespace(p.namespace))
	if err != nil {
		return responsePayload, err
	}
	if len(list.Items) > 0 {
		responsePayload.User.Groups = make([]string, 0, len(list.Items))
		for idx, _ := range list.Items {
			responsePayload.User.Groups = append(responsePayload.User.Groups, list.Items[idx].Spec.Group)
		}
	}
	// Try to fetch user
	usr := userdbv1alpha1.User{}
	err = p.kubeClient.Get(context.TODO(), client.ObjectKey{
		Namespace: p.namespace,
		Name:      request.Login,
	}, &usr)
	if client.IgnoreNotFound(err) != nil {
		return responsePayload, err
	}
	if err != nil {
		p.logger.V(1).Info("User not found", "user", request.Login)
		responsePayload.UserStatus = proto.NotFound
		return responsePayload, nil
	}
	if usr.Spec.Uid != nil {
		responsePayload.User.Uid = int64(*usr.Spec.Uid)
	}
	if len(usr.Spec.CommonNames) > 0 { // Avoid copying a nil
		responsePayload.User.CommonNames = usr.Spec.CommonNames
	}
	if len(usr.Spec.Emails) > 0 { // Avoid copying a nil
		responsePayload.User.Emails = usr.Spec.Emails
	}
	if usr.Spec.Disabled != nil && *usr.Spec.Disabled {
		p.logger.V(1).Info("User found but disabled", "user", request.Login)
		responsePayload.UserStatus = proto.Disabled
	} else {
		if request.Password != "" && usr.Spec.PasswordHash != "" {
			err := bcrypt.CompareHashAndPassword([]byte(usr.Spec.PasswordHash), []byte(request.Password))
			if err == nil {
				responsePayload.UserStatus = proto.PasswordChecked
			} else {
				responsePayload.UserStatus = proto.PasswordFail
			}
		} else {
			responsePayload.UserStatus = proto.PasswordUnchecked
		}
		p.logger.V(1).Info("User found", "user", responsePayload.User.Login, "status", responsePayload.UserStatus)
	}
	return responsePayload, nil
}
