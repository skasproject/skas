package handlers

import (
	"net/http"
	"skas/sk-auth/internal/config"
	"skas/sk-auth/internal/tokenstore"
	"skas/sk-common/pkg/misc"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/pkg/skserver/protector"
	"skas/sk-common/proto/v1/proto"
	"strings"
)

var _ commonHandlers.HttpRequestValidator = &AdminHttpRequestValidator{}

type AdminHttpRequestValidator struct {
	TokenStore     tokenstore.TokenStore
	IdentityGetter commonHandlers.IdentityGetter
	Protector      protector.Protector
}

func (a *AdminHttpRequestValidator) Validate(request *http.Request, response http.ResponseWriter) misc.HttpError {
	user, err, locked := a.getAuthUser(request)
	if err != nil {
		return misc.NewHttpError("Server error. Check server logs", http.StatusInternalServerError)
	}
	if locked {
		return misc.NewHttpError("Locked", http.StatusServiceUnavailable)
	}
	if user == nil {
		response.Header().Set("WWW-Authenticate", "Basic realm=\"/skas\"")
		return misc.NewHttpError("Need to authenticate", http.StatusUnauthorized)
	}
	if !userInGroups(user, config.Conf.AdminGroups) {
		return misc.NewHttpError("User has no admin rights", http.StatusUnauthorized)
	}
	return nil
}

func getBearerToken(request *http.Request) string {
	authList, ok := request.Header["Authorization"]
	if !ok {
		return ""
	}
	for _, auth := range authList {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimSpace(auth[len("Bearer "):])
		}
	}
	return ""
}

func userInGroups(user *proto.User, groups []string) bool {
	//fmt.Printf("userInGroups() user:%v  groups:%v", user, groups)
	if user.Groups == nil || groups == nil {
		return false
	}
	for _, grp := range user.Groups {
		for _, g2 := range groups {
			if grp == g2 {
				return true
			}
		}
	}
	return false
}

func (a *AdminHttpRequestValidator) getAuthUser(request *http.Request) (*proto.User, error /* locked */, bool) {
	login, password, ok := request.BasicAuth()
	if ok {
		locked := a.Protector.EntryForLogin(login)
		if locked {
			return nil, nil, true
		}
		user, _, err := doLogin(a.IdentityGetter, login, password, a.Protector)
		if err != nil {
			return nil, err, false
		}
		return user, nil, false
	}
	// Try with a skas token
	token := getBearerToken(request)
	if token != "" {
		locked := a.Protector.EntryForToken()
		if locked {
			return nil, nil, true
		}
		user, err := a.TokenStore.Get(token)
		if err != nil {
			return nil, err, false
		}
		if user == nil {
			a.Protector.TokenNotFound()
		}
		return user, nil, false
	}
	// No authentication in header
	return nil, nil, false
}
