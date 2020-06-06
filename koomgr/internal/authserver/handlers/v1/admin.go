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
	"github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koomgr/internal/authserver/handlers"
	"html/template"
	"net/http"
	"path"
	"strings"
)

type AdminV1Handler struct {
	handlers.AuthHandler
	AdminGroup   string
	TemplatePath string
	subEntries   []subEntry
}

type handlerFunc func(handler *AdminV1Handler, usr common.User, remainingPath string, response http.ResponseWriter, request *http.Request)

type subEntry struct {
	method      string
	path        string
	handlerFunc handlerFunc
}

func (this *AdminV1Handler) init() {
	this.subEntries = make([]subEntry, 0, 2)
	this.subEntries = append(this.subEntries, subEntry{
		method:      "DELETE",
		path:        "tokens/",
		handlerFunc: deleteToken,
	})
	this.subEntries = append(this.subEntries, subEntry{
		method:      "GET",
		path:        "tokens",
		handlerFunc: listToken,
	})
	this.subEntries = append(this.subEntries, subEntry{
		method:      "GET",
		path:        "users/",
		handlerFunc: describeUser,
	})
}

func describeUser(handler *AdminV1Handler, usr common.User, remainingPath string, response http.ResponseWriter, request *http.Request) {
	userDescribeResponse, err := handler.Providers.DescribeUser(remainingPath)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
	handler.ServeJSON(response, userDescribeResponse)
}

func listToken(handler *AdminV1Handler, usr common.User, remainingPath string, response http.ResponseWriter, request *http.Request) {
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

func deleteToken(handler *AdminV1Handler, usr common.User, remainingPath string, response http.ResponseWriter, request *http.Request) {
	ok, err := handler.TokenBasket.Delete(remainingPath)
	if err != nil {
		http.Error(response, "Server error. Check server logs", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(response, "Not found", http.StatusNotFound)
	}
}

func (this *AdminV1Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if this.subEntries == nil {
		this.init()
	}
	this.ServeAuthHTTP(response, request, func(usr common.User) {
		if this.AdminGroup != "" && stringInSlice(this.AdminGroup, usr.Groups) {
			this.Logger.V(1).Info(fmt.Sprintf("user '%s' allowed to access admin interface", usr.Username))
			urlpath := request.URL.Path[this.PrefixLength:]
			for _, entry := range this.subEntries {
				if entry.method == request.Method && strings.HasPrefix(urlpath, entry.path) {
					remainingPath := urlpath[len(entry.path):]
					entry.handlerFunc(this, usr, strings.TrimSpace(remainingPath), response, request)
					return
				}
			}
			http.Error(response, "Not found", http.StatusNotFound)
		} else {
			this.Logger.V(1).Info(fmt.Sprintf("user '%s': access to admin interface denied (Not in appropriate group)", usr.Username))
			http.Error(response, "Unallowed", http.StatusForbidden)
		}
	})
}

func (this *AdminV1Handler) serveHTML(response http.ResponseWriter, data interface{}, tmplName string) {
	tmpl := template.Must(template.ParseFiles(path.Join(this.TemplatePath, tmplName)))
	err := tmpl.Execute(response, data)
	if err != nil {
		panic(err)
	}
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
