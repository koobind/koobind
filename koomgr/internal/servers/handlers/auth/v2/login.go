package v2

import (
	"encoding/json"
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
	if clientId := this.LookupClient(requestPayload.Client); clientId == "" {
		this.Logger.Info("Invalid clientId or clientSecret for login.", "clientId", requestPayload.Client.Id, "user", requestPayload.Login)
		this.HttpError(response, "Invalid clientId/clientSecret", http.StatusForbidden)
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
		this.Logger.Info("Login successful", "clientId", requestPayload.Client.Id, "user", usr.Name, "groups", usr.Groups, "token", responsePayload.Token)
		this.ServeJSON(response, responsePayload)
	} else {
		this.Logger.Info("Invalid login", "user", requestPayload.Client.Id, requestPayload.Login)
		this.HttpError(response, "Unauthorized", http.StatusUnauthorized)
	}
}
