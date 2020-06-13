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
	"fmt"
	"github.com/gorilla/mux"
	"github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koomgr/internal/authserver/handlers"
	"net/http"
)

type AdminV1Handler struct {
	handlers.AuthHandler
	AdminGroup  string
	handlerFunc handlerFunc
}

type handlerFunc func(handler *AdminV1Handler, usr common.User, response http.ResponseWriter, request *http.Request)

func describeUser(handler *AdminV1Handler, usr common.User, response http.ResponseWriter, request *http.Request) {
	user := mux.Vars(request)["user"]
	found, userDescribeResponse := handler.Providers.DescribeUser(user)
	if !found {
		http.Error(response, fmt.Sprintf("User %s not found", user), http.StatusNotFound)
	}
	handler.ServeJSON(response, userDescribeResponse)
}

func listToken(handler *AdminV1Handler, usr common.User, response http.ResponseWriter, request *http.Request) {
	list, err := handler.TokenBasket.GetAll()
	if err != nil {
		http.Error(response, "Server error. Check server logs", http.StatusInternalServerError)
		return
	}
	data := common.TokenListResponse{
		Tokens: list,
	}
	handler.ServeJSON(response, data)
}

func deleteToken(handler *AdminV1Handler, usr common.User, response http.ResponseWriter, request *http.Request) {
	token := mux.Vars(request)["token"]
	ok, err := handler.TokenBasket.Delete(token)
	if err != nil {
		http.Error(response, "Server error. Check server logs", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(response, "Not found", http.StatusNotFound)
	}
}

func (this *AdminV1Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	this.ServeAuthHTTP(response, request, func(usr common.User) {
		if this.AdminGroup != "" && stringInSlice(this.AdminGroup, usr.Groups) {
			this.Logger.V(1).Info(fmt.Sprintf("user '%s' allowed to access admin interface", usr.Username))
			this.handlerFunc(this, usr, response, request)
		} else {
			this.Logger.V(1).Info(fmt.Sprintf("user '%s': access to admin interface denied (Not in appropriate group)", usr.Username))
			http.Error(response, "Unallowed", http.StatusForbidden)
		}
	})
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

type TokensDataModel struct {
	Title  string             `json:"title"`
	Tokens []common.UserToken `json:"tokens"`
}
