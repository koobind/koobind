package v2

import (
	"encoding/json"
	"fmt"
	proto "github.com/koobind/koobind/koomgr/apis/proto/auth/v2"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"github.com/koobind/koobind/koomgr/internal/servers/handlers"
	"net/http"
)

type AuthLoginHandler struct {
	handlers.BaseHandler
	Providers providers.ProviderChain
}

func (this *AuthLoginHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload proto.LoginRequest
	err := json.NewDecoder(request.Body).Decode(&requestPayload)
	if err != nil {
		this.HttpError(response, err.Error(), http.StatusBadRequest)
		return
	}
	usr, ok, err := this.Providers.Login(requestPayload.Login, requestPayload.Password)
	if err != nil {
		this.HttpError(response, "Server error. Check server logs", http.StatusInternalServerError)
		return
	}
	if ok {
		responsePayload := proto.LoginResponse{
			Username:      usr.Name,
			Uid:           usr.Uid,
			EmailVerified: false,
			Groups:        usr.Groups,
			Emails:        usr.Emails,
			CommonNames:   usr.CommonNames,
		}
		if requestPayload.GenerateToken {
			userToken, err := this.TokenBasket.NewUserToken(usr)
			if err != nil {
				this.HttpError(response, "Server error. Check server logs", http.StatusInternalServerError)
				return
			}
			responsePayload.Token = userToken.Token
			responsePayload.ClientTTL = userToken.Spec.Lifecycle.ClientTTL.Duration
		}
		this.Logger.Info(fmt.Sprintf("Login successful for user:'%s'  uid:%s, groups=%v with token:'%s'", usr.Name, usr.Uid, usr.Groups, responsePayload.Token))
		this.ServeJSON(response, responsePayload)
	} else {
		this.Logger.Info(fmt.Sprintf("Invalid login for user '%s'.", requestPayload.Login))
		this.HttpError(response, "Unauthorized", http.StatusUnauthorized)
	}
}
