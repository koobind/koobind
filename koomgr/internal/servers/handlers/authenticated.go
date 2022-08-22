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

package handlers

import (
	"encoding/base64"
	"github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"net/http"
	"strings"
)

type AuthHandler struct {
	BaseHandler
	Providers providers.ProviderChain
}

func (this *AuthHandler) ServeAuthenticatedHTTP(response http.ResponseWriter, request *http.Request, fn func(user common.User)) {
	authList, ok := request.Header["Authorization"]
	if !ok || len(authList) < 1 || !(strings.HasPrefix(authList[0], "Basic ") || strings.HasPrefix(authList[0], "Bearer ")) {
		response.Header().Set("WWW-Authenticate", "Basic realm=\"/koo\"")
		this.HttpError(response, "Need to authenticate", http.StatusUnauthorized)
	} else {
		var usr common.User
		var ok bool
		if strings.HasPrefix(authList[0], "Basic ") {
			b64 := authList[0][len("Basic "):]
			data, err := base64.StdEncoding.DecodeString(b64)
			if err != nil || !strings.Contains(string(data), ":") {
				this.HttpError(response, "Unable to decode Authorization header", http.StatusBadRequest)
			} else {
				up := strings.Split(string(data), ":")
				login := strings.TrimSpace(up[0])
				password := strings.TrimSpace(up[1])
				usr, ok, _, err = this.Providers.Login(login, password)
				if err != nil {
					this.HttpError(response, "Server error. Check server logs", http.StatusInternalServerError)
					return
				}
			}
		} else {
			// It is Bearer
			token := strings.TrimSpace(authList[0][len("Bearer "):])
			var err error
			usr, ok, err = this.TokenBasket.Get(token)
			if err != nil {
				this.HttpError(response, "Server error. Check server logs", http.StatusInternalServerError)
				return
			}
		}
		if ok {
			fn(usr)
		} else {
			response.Header().Set("WWW-Authenticate", "Basic realm=\"/koo\"")
			this.HttpError(response, "Need to authenticate", http.StatusUnauthorized)
		}
	}
}
