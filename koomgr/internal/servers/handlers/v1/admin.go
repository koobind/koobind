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
	tokenapi "github.com/koobind/koobind/koomgr/apis/tokens/v1alpha1"
	"github.com/koobind/koobind/koomgr/internal/servers/handlers"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AdminV1Handler struct {
	handlers.AuthHandler
	AdminGroup  string
	KubeClient  client.Client
	HandlerFunc HandlerFunc
}

type HandlerFunc func(handler *AdminV1Handler, usr tokenapi.UserDesc, response http.ResponseWriter, request *http.Request)

func (this *AdminV1Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	this.ServeAuthenticatedHTTP(response, request, func(usr tokenapi.UserDesc) {
		if this.AdminGroup != "" && stringInSlice(this.AdminGroup, usr.Groups) {
			this.Logger.V(1).Info(fmt.Sprintf("user '%s' allowed to access admin interface", usr.Name))
			this.HandlerFunc(this, usr, response, request)
		} else {
			this.Logger.V(1).Info(fmt.Sprintf("user '%s': access to admin interface denied (Not in appropriate group)", usr.Name))
			this.HttpError(response, "Unallowed", http.StatusForbidden)
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
