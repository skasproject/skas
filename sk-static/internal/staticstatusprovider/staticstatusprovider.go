package staticstatusprovider

import (
	"github.com/go-logr/logr"
	"golang.org/x/crypto/bcrypt"
	"skas/sk-common/pkg/httpserver/handlers"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-static/internal/config"
)

var _ handlers.IdentityServerProvider = &staticIdentityProvider{}

type staticIdentityProvider struct {
	logger logr.Logger
}

func New(logger logr.Logger) handlers.IdentityServerProvider {
	return &staticIdentityProvider{
		logger: logger,
	}
}

func (s staticIdentityProvider) GetUserIdentity(request proto.UserIdentityRequest) (*proto.UserIdentityResponse, error) {
	responsePayload := &proto.UserIdentityResponse{
		User: proto.User{
			Login:       request.Login,
			Uid:         0,
			Emails:      []string{},
			CommonNames: []string{},
			Groups:      []string{},
		},
		UserStatus: proto.NotFound,
	}
	// Handle groups, even if not found
	groups, ok := config.GroupsByUser[request.Login]
	if ok {
		responsePayload.Groups = groups
	}

	user, ok := config.UserByLogin[request.Login]
	if !ok {
		s.logger.V(1).Info("User not found", "user", request.Login)
		responsePayload.UserStatus = proto.NotFound
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
		responsePayload.UserStatus = proto.Disabled
	} else {
		if request.Password == "" || user.PasswordHash == "" {
			responsePayload.UserStatus = proto.PasswordUnchecked
		} else {
			err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))
			if err == nil {
				responsePayload.UserStatus = proto.PasswordChecked
			} else {
				responsePayload.UserStatus = proto.PasswordFail
			}
		}
		s.logger.V(1).Info("User found", "user", responsePayload.Login, "status", responsePayload.UserStatus)
	}
	return responsePayload, nil
}
