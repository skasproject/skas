package identitygetterr

import (
	"context"
	"github.com/go-logr/logr"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	userdbv1alpha1 "skas/sk-common/k8sapis/userdb/v1alpha1"
	"skas/sk-common/pkg/misc"
	"skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
)

var _ handlers.IdentityGetter = &crdIdentityGetter{}

type crdIdentityGetter struct {
	kubeClient client.Client
	namespace  string
	logger     logr.Logger
}

func New(kubeClient client.Client, namespace string, logger logr.Logger) handlers.IdentityGetter {
	return &crdIdentityGetter{
		kubeClient: kubeClient,
		namespace:  namespace,
		logger:     logger,
	}
}

func (p crdIdentityGetter) GetIdentity(request proto.IdentityRequest) (*proto.IdentityResponse, misc.HttpError) {
	if request.Detailed {
		return nil, misc.NewHttpError("Can't handle detailed request", http.StatusBadRequest)
	}
	responsePayload := &proto.IdentityResponse{
		Status:    proto.UserNotFound,
		User:      proto.InitUser(request.Login),
		Details:   []proto.UserDetail{},
		Authority: "",
	}
	// ------------------- Handle groups (Even if notFound)
	list := userdbv1alpha1.GroupBindingList{}
	err := p.kubeClient.List(context.TODO(), &list, client.MatchingFields{"userkey": request.Login}, client.InNamespace(p.namespace))
	if err != nil {
		return responsePayload, misc.NewHttpError(err.Error(), http.StatusInternalServerError)
	}
	if len(list.Items) > 0 {
		responsePayload.Groups = make([]string, 0, len(list.Items))
		for idx, _ := range list.Items {
			responsePayload.Groups = append(responsePayload.Groups, list.Items[idx].Spec.Group)
		}
	}
	// Try to fetch user
	usr := userdbv1alpha1.User{}
	err = p.kubeClient.Get(context.TODO(), client.ObjectKey{
		Namespace: p.namespace,
		Name:      request.Login,
	}, &usr)
	if client.IgnoreNotFound(err) != nil {
		return responsePayload, misc.NewHttpError(err.Error(), http.StatusInternalServerError)
	}
	if err != nil {
		p.logger.V(1).Info("User not found", "user", request.Login)
		responsePayload.Status = proto.UserNotFound
		return responsePayload, nil
	}
	if usr.Spec.Uid != nil {
		responsePayload.Uid = *usr.Spec.Uid
	}
	if len(usr.Spec.CommonNames) > 0 { // Avoid copying a nil
		responsePayload.CommonNames = usr.Spec.CommonNames
	}
	if len(usr.Spec.Emails) > 0 { // Avoid copying a nil
		responsePayload.Emails = usr.Spec.Emails
	}
	if usr.Spec.Disabled != nil && *usr.Spec.Disabled {
		p.logger.V(1).Info("User found but disabled", "user", request.Login)
		responsePayload.Status = proto.Disabled
	} else {

		if usr.Spec.PasswordHash == "" {
			responsePayload.Status = proto.PasswordMissing
		} else if request.Password == "" {
			responsePayload.Status = proto.PasswordUnchecked
		} else {
			err := bcrypt.CompareHashAndPassword([]byte(usr.Spec.PasswordHash), []byte(request.Password))
			if err == nil {
				responsePayload.Status = proto.PasswordChecked
			} else {
				responsePayload.Status = proto.PasswordFail
			}
		}
		p.logger.V(1).Info("User found", "user", responsePayload.Login, "status", responsePayload.Status)
	}
	return responsePayload, nil
}
