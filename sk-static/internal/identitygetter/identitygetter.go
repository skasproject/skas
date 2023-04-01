package identitygetter

import (
	"github.com/go-logr/logr"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"skas/sk-common/pkg/misc"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-static/internal/config"
)

var _ commonHandlers.IdentityGetter = &staticIdentityGetter{}

type staticIdentityGetter struct {
	logger logr.Logger
}

func New(logger logr.Logger) commonHandlers.IdentityGetter {
	return &staticIdentityGetter{
		logger: logger,
	}
}

func (s staticIdentityGetter) GetIdentity(request proto.IdentityRequest) (*proto.IdentityResponse, misc.HttpError) {
	if request.Detailed {
		return nil, misc.NewHttpError("Can't handle detailed request", http.StatusBadRequest)
	}
	responsePayload := &proto.IdentityResponse{
		User:      proto.InitUser(request.Login),
		Status:    proto.NotFound,
		Details:   []proto.UserDetail{},
		Authority: "",
	}
	// Handle groups, even if not found
	groups, ok := config.GroupsByUser[request.Login]
	if ok {
		responsePayload.Groups = groups
	}

	user, ok := config.UserByLogin[request.Login]
	if !ok {
		s.logger.V(1).Info("User not found", "user", request.Login)
		responsePayload.Status = proto.NotFound
		return responsePayload, nil
	}
	if user.Uid != nil {
		responsePayload.Uid = *user.Uid
	}
	if len(user.CommonNames) > 0 { // Avoid copying a nil
		responsePayload.CommonNames = user.CommonNames
	}
	if len(user.Emails) > 0 { // Avoid copying a nil
		responsePayload.Emails = user.Emails
	}
	if user.Disabled != nil && *user.Disabled {
		s.logger.V(1).Info("User found but disabled", "user", request.Login)
		responsePayload.Status = proto.Disabled
	} else {
		if user.PasswordHash == "" {
			responsePayload.Status = proto.PasswordMissing
		} else if request.Password == "" {
			responsePayload.Status = proto.PasswordUnchecked
		} else {
			err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))
			if err == nil {
				responsePayload.Status = proto.PasswordChecked
			} else {
				responsePayload.Status = proto.PasswordFail
			}
		}
		s.logger.V(1).Info("User found", "user", responsePayload.Login, "status", responsePayload.Status)
	}
	return responsePayload, nil
}
