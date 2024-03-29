package handlers

//
import (
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"net/http"
	"skas/sk-auth/internal/tokenstore"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/pkg/skserver/protector"
	"skas/sk-common/proto/v1/proto"
	"strconv"
)

var _ http.Handler = &TokenReviewHandler{}

type TokenReviewHandler struct {
	commonHandlers.BaseHandler
	TokenStore tokenstore.TokenStore
	Protector  protector.TokenProtector
}

func (t *TokenReviewHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload proto.TokenReviewRequest
	err := json.NewDecoder(request.Body).Decode(&requestPayload)
	if err != nil {
		t.HttpSendError(response, err.Error(), http.StatusBadRequest)
		return
	}
	locked := t.Protector.EntryForToken()
	if locked {
		t.HttpSendError(response, "Locked", http.StatusServiceUnavailable)
		return
	}

	data := &proto.TokenReviewResponse{
		ApiVersion: requestPayload.ApiVersion,
		Kind:       requestPayload.Kind,
	}
	user, err := t.TokenStore.Get(requestPayload.Spec.Token)
	if err != nil {
		t.HttpSendError(response, "Server error. Check server logs", http.StatusInternalServerError)
		return
	}
	if user != nil {
		data.Status.Authenticated = true
		data.Status.User = &proto.TokenReviewUser{
			Username: user.Login,
			Uid:      strconv.Itoa(user.Uid),
			Groups:   user.Groups,
		}
		t.Logger.Info(fmt.Sprintf("Token '%s' OK. user:'%s'  uid:%s, groups=%v", requestPayload.Spec.Token, data.Status.User.Username, data.Status.User.Uid, data.Status.User.Groups))
	} else {
		t.Protector.TokenNotFound()
		t.Logger.Info(fmt.Sprintf("Token '%s' rejected", requestPayload.Spec.Token))
		data.Status.Authenticated = false
		data.Status.User = nil
	}
	t.ServeJSON(response, data)

}

func (t *TokenReviewHandler) GetLog() logr.Logger {
	return t.Logger
}

func (t *TokenReviewHandler) SetLog(logger logr.Logger) {
	t.Logger = logger
}
