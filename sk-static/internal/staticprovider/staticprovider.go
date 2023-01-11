package staticprovider

import (
	"golang.org/x/crypto/bcrypt"
	"skas/sk-common/pkg/httpserver/handlers"
	"skas/sk-common/proto"
	"skas/sk-static/internal/config"
)

var _ handlers.StatusProvider = &staticProvider{}

type staticProvider struct {
}

func New() handlers.StatusProvider {
	return &staticProvider{}
}

func (s staticProvider) GetUserStatus(request proto.UserStatusRequest) (*proto.UserStatusResponse, error) {
	responsePayload := &proto.UserStatusResponse{
		UserStatus: proto.NotFound,
	}
	user, ok := config.UserByLogin[request.Login]
	if ok {
		responsePayload.User = &proto.User{
			Login:       user.Login,
			Uid:         user.Uid,
			CommonNames: user.CommonNames,
			Emails:      user.Emails,
			Groups:      user.Groups,
		}
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
	}
	return responsePayload, nil
}
