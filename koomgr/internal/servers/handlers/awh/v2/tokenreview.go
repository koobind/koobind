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

package v2

//
import (
	"encoding/json"
	"fmt"
	proto "github.com/koobind/koobind/koomgr/apis/proto/awh/v2"
	"github.com/koobind/koobind/koomgr/internal/servers/handlers"
	"net/http"
)

type TokenReviewHandler struct {
	handlers.BaseHandler
}

func (this *TokenReviewHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var requestPayload proto.TokenReviewRequest
	err := json.NewDecoder(request.Body).Decode(&requestPayload)
	if err != nil {
		this.HttpError(response, err.Error(), http.StatusBadRequest)
	} else {
		data := proto.TokenReviewResponse{
			ApiVersion: requestPayload.ApiVersion,
			Kind:       requestPayload.Kind,
		}
		userToken, err := this.TokenBasket.Get(requestPayload.Spec.Token)
		if err != nil {
			this.HttpError(response, "Server error. Check server logs", http.StatusInternalServerError)
			return
		}
		if userToken != nil {
			data.Status.Authenticated = true
			data.Status.User = &proto.TokenReviewUser{
				Username: userToken.Spec.User.Name,
				Uid:      userToken.Spec.User.Uid,
				Groups:   userToken.Spec.User.Groups,
			}
			this.Logger.Info(fmt.Sprintf("Token '%s' OK. user:'%s'  uid:%s, groups=%v", requestPayload.Spec.Token, data.Status.User.Username, data.Status.User.Uid, data.Status.User.Groups))
		} else {
			this.Logger.Info(fmt.Sprintf("Token '%s' rejected", requestPayload.Spec.Token))
			data.Status.Authenticated = false
			data.Status.User = nil
		}
		this.ServeJSON(response, data)
	}
}
