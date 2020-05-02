package v1

import (
	"encoding/json"
	"fmt"
	. "github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koomgr/internal/authserver/handlers"
	"net/http"
)

type ValidateTokenHandler struct {
	handlers.BaseHandler
}

func (this *ValidateTokenHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	// POST is from Api server while GET is from our client
	if request.Method == "POST" || request.Method == "GET" {
		var requestPayload ValidateTokenRequest
		err := json.NewDecoder(request.Body).Decode(&requestPayload)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
		} else {
			data := ValidateTokenResponse{
				ApiVersion: requestPayload.ApiVersion,
				Kind:       requestPayload.Kind,
			}
			usr, ok, err := this.TokenBasket.Get(requestPayload.Spec.Token)
			if err != nil {
				http.Error(response, "Server error. Check server logs", http.StatusInternalServerError)
				return
			}
			if ok {
				this.Logger.Info(fmt.Sprintf("Token '%s' OK. user:'%s'  uid:%s, groups=%v", requestPayload.Spec.Token, usr.Username, usr.Uid, usr.Groups))
				data.Status.Authenticated = true
				data.Status.User = &usr
			} else {
				this.Logger.Info(fmt.Sprintf("Token '%s' rejected", requestPayload.Spec.Token))
				data.Status.Authenticated = false
				data.Status.User = nil
			}
			this.ServeJSON(response, data)
		}
	} else {
		http.Error(response, "Not Found", http.StatusNotFound)
	}
}
