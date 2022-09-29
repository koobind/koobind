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
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	proto_v2 "github.com/koobind/koobind/koomgr/apis/proto/auth/v2"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/token"
	"net/http"
	"strings"
)

type BaseHandler struct {
	Logger      logr.Logger
	TokenBasket token.TokenBasket
	RequestId   int
}

// Each REST call must be concluded by one of these function
func (this *BaseHandler) ServeJSON(response http.ResponseWriter, data interface{}) {
	response.Header().Set("Content-Type", "application/json")
	if this.Logger.V(1).Enabled() {
		this.Logger.V(1).Info(fmt.Sprintf("------ httpClose:Emit JSON:%s", json2String(data)))
	}
	response.WriteHeader(http.StatusOK)
	err := json.NewEncoder(response).Encode(data)
	if err != nil {
		panic(err)
	}
}

func (this *BaseHandler) HttpError(response http.ResponseWriter, message string, httpCode int) {
	if this.Logger.V(1).Enabled() {
		this.Logger.V(1).Info("------ httpError", "message", message, "httpCode", httpCode)
	}
	http.Error(response, message, httpCode)
}

func (this *BaseHandler) HttpClose(response http.ResponseWriter, message string, httpCode int) {
	if this.Logger.V(1).Enabled() {
		this.Logger.V(1).Info("------ httpClose", "message", message, "httpCode", httpCode)
	}
	if message != "" {
		response.Header().Set("Content-Type", "text/plain; charset=utf-8")
		response.Header().Set("X-Content-Type-Options", "nosniff")
		response.WriteHeader(httpCode)
		_, _ = fmt.Fprintln(response, message)
	} else {
		response.WriteHeader(httpCode)
	}
}

func json2String(data interface{}) string {
	builder := &strings.Builder{}
	_ = json.NewEncoder(builder).Encode(data)
	return builder.String()
}

func (this *BaseHandler) LookupClient(client proto_v2.AuthClient) (clientId string) {
	authClient, ok := config.Conf.AuthClientById[client.Id]
	if !ok {
		this.Logger.V(1).Info("Unable to find client.Id", "client.Id", client.Id)
		return ""
	}
	if authClient.Secret != client.Secret {
		this.Logger.V(1).Info("client.Secret mismatch", "client.Id", client.Id, "client.Secret", client.Secret, "server.Secret", authClient.Secret)
		return ""
	}
	this.Logger.V(1).Info("Found client", "client.Id", authClient.Id)
	return authClient.Id
}
