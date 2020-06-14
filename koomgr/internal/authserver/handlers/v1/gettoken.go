/*
  Copyright (C) 2020 Serge ALEXANDRE

  This file is part of koobind project

  koobind is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  koobind is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with koobind.  If not, see <http://www.gnu.org/licenses/>.
*/
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
					this.Logger.Info(fmt.Sprintf("Token '%s' granted to user:'%s'  uid:%s, groups=%v", data.Token, usr.Username, usr.Uid, usr.Groups))
					this.ServeJSON(response, data)
				}
			} else {
				this.Logger.Info(fmt.Sprintf("No token granted to user '%s'. Unable to validate this login.", login))
				http.Error(response, "Unallowed", http.StatusUnauthorized)
			}
		}
	}
}
