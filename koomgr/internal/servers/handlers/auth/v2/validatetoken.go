package v2

import (
	"encoding/json"
	"fmt"
	proto "github.com/koobind/koobind/koomgr/apis/proto/auth/v2"
	"github.com/koobind/koobind/koomgr/internal/servers/handlers"
	"net/http"
)

type ValidateTokenHandler struct {
	handlers.BaseHandler
}

func (this *ValidateTokenHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload proto.ValidateTokenRequest
	err := json.NewDecoder(request.Body).Decode(&requestPayload)
	if err != nil {
		this.HttpError(response, err.Error(), http.StatusBadRequest)
	} else {
		responsePayload := proto.ValidateTokenResponse{
			Token: requestPayload.Token,
		}
		userToken, err := this.TokenBasket.Get(requestPayload.Token)
		if err != nil {
			this.HttpError(response, "Server error. Check server logs", http.StatusInternalServerError)
			return
		}
		if userToken != nil {
			responsePayload.Valid = true
			this.Logger.Info(fmt.Sprintf("Token '%s' OK. user:'%s'  uid:%s, groups=%v", userToken.Token, userToken.Spec.User.Name, userToken.Spec.User.Uid, userToken.Spec.User.Groups))
		} else {
			responsePayload.Valid = false
			this.Logger.Info(fmt.Sprintf("Token '%s' rejected", requestPayload.Token))
		}
		this.ServeJSON(response, responsePayload)
	}
}
