package handlers

//
import (
	"encoding/json"
	"fmt"
	"net/http"
	"skas/sk-auth/internal/tokenstore"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
	"strconv"
)

type TokenReviewHandler struct {
	commonHandlers.BaseHandler
	TokenStore tokenstore.TokenStore
}

func (t *TokenReviewHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload proto.TokenReviewRequest
	err := json.NewDecoder(request.Body).Decode(&requestPayload)
	if err != nil {
		t.HttpError(response, err.Error(), http.StatusBadRequest)
	} else {
		data := &proto.TokenReviewResponse{
			ApiVersion: requestPayload.ApiVersion,
			Kind:       requestPayload.Kind,
		}
		user, err := t.TokenStore.Get(requestPayload.Spec.Token)
		if err != nil {
			t.HttpError(response, "Server error. Check server logs", http.StatusInternalServerError)
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
			t.Logger.Info(fmt.Sprintf("Token '%s' rejected", requestPayload.Spec.Token))
			data.Status.Authenticated = false
			data.Status.User = nil
		}
		t.ServeJSON(response, data)
	}
}
