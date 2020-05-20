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
