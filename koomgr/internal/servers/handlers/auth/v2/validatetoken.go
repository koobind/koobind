package v2

import (
	"encoding/json"
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
		return
	}
	if clientId := this.LookupClient(requestPayload.Client); clientId == "" {
		this.Logger.Info("Invalid clientId or clientSecret for token validation.", "clientId", requestPayload.Client.Id, "token", requestPayload.Token)
		this.HttpError(response, "Invalid clientId/clientSecret", http.StatusForbidden)
		return
	}
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
		this.Logger.Info("Token OK", "clientId", requestPayload.Client.Id, "token", userToken.Token, "user", userToken.Spec.User.Name, "groups", userToken.Spec.User.Groups)
	} else {
		responsePayload.Valid = false
		this.Logger.Info("Token rejected", "clientId", requestPayload.Client.Id, "token", requestPayload.Token)
	}
	this.ServeJSON(response, responsePayload)
}
