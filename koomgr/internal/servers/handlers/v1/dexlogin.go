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
	"encoding/json"
	"fmt"
	"github.com/koobind/koobind/koomgr/apis/proto"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"github.com/koobind/koobind/koomgr/internal/servers/handlers"
	"net/http"
)

type DexLoginHandler struct {
	handlers.BaseHandler
	Providers providers.ProviderChain
}

func (this *DexLoginHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload proto.DexLoginRequest
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
		userToken, err := this.TokenBasket.NewUserToken("dex", usr)
		if err != nil {
			this.HttpError(response, "Server error. Check server logs", http.StatusInternalServerError)
		} else {
			data := proto.DexLoginResponse{
				Name:   userToken.Spec.User.Name,
				Uid:    userToken.Spec.User.Uid,
				Groups: userToken.Spec.User.Groups,
				Token:  userToken.Token,
			}
			if len(userToken.Spec.User.CommonNames) > 0 {
				data.CommonName = userToken.Spec.User.CommonNames[0]
			}
			if len(userToken.Spec.User.Emails) > 0 {
				data.Email = userToken.Spec.User.Emails[0]
			}
			this.Logger.Info(fmt.Sprintf("Login successful for user:'%s'  uid:%s, groups=%v with token:'%s'", usr.Name, usr.Uid, usr.Groups, data.Token))
			this.ServeJSON(response, data)
		}
	} else {
		this.Logger.Info(fmt.Sprintf("Invalid login for user '%s'.", requestPayload.Login))
		this.HttpError(response, "Unauthorized", http.StatusUnauthorized)
	}
}
