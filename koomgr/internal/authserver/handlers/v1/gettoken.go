package v1

import (
	"encoding/base64"
	"fmt"
	. "github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koomgr/internal/authserver/handlers"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"net/http"
	"strings"
)

type GetTokenHandler struct {
	handlers.BaseHandler
	Providers providers.ProviderChain
}

func (this *GetTokenHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		this.RequestId++
		authList, ok := request.Header["Authorization"]
		if !ok || len(authList) < 1 || !strings.HasPrefix(authList[0], "Basic ") {
			response.Header().Set("WWW-Authenticate", "Basic realm=\"/getToken\"")
			http.Error(response, "Need to authenticate", http.StatusUnauthorized)
		} else {
			b64 := authList[0][len("Basic "):]
			data, err := base64.StdEncoding.DecodeString(b64)
			if err != nil || !strings.Contains(string(data), ":") {
				http.Error(response, "Unable to decode Authorization header", http.StatusBadRequest)
			} else {
				up := strings.Split(string(data), ":")
				login := up[0]
				password := up[1]
				usr, ok, _, err := this.Providers.Login(login, password)
				if err != nil {
					http.Error(response, "Server error. Check server logs", http.StatusInternalServerError)
					return
				}
				if ok {
					userToken, err := this.TokenBasket.NewUserToken(usr)
					if err != nil {
						http.Error(response, "Server error. Check server logs", http.StatusInternalServerError)
					} else {
						data := GetTokenResponse{
							Token:     userToken.Token,
							ClientTTL: userToken.Lifecycle.ClientTTL,
						}
						this.ServeJSON(response, data)
						this.Logger.Info(fmt.Sprintf("Token '%s' granted to user:'%s'  uid:%s, groups=%v", data.Token, usr.Username, usr.Uid, usr.Groups))
					}
				} else {
					this.Logger.Info(fmt.Sprintf("No token granted to user '%s'. Unable to validate this login.", login))
					http.Error(response, "Unallowed", http.StatusUnauthorized)
				}
			}
		}
	} else {
		http.Error(response, "Not found", http.StatusNotFound)
	}
}
