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
	"github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koomgr/internal/token"
	"net/http"
)

type BaseHandler struct {
	Logger       logr.Logger
	TokenBasket  token.TokenBasket
	PrefixLength int
	RequestId    int
}

func (this *BaseHandler) ServeJSON(response http.ResponseWriter, data interface{}) {
	response.Header().Set("Content-Type", "application/json")
	if this.Logger.V(1).Enabled() {
		this.Logger.V(1).Info(fmt.Sprintf("Emit JSON:%s", common.JSON2String(data)))

	}
	err := json.NewEncoder(response).Encode(data)
	if err != nil {
		panic(err)
	}
}
