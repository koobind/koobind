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

package token

import (
	"github.com/koobind/koobind/koomgr/apis/proto"
	tokenapi "github.com/koobind/koobind/koomgr/apis/tokens/v1alpha1"
)

type TokenBasket interface {
	NewUserToken(clientId string, user tokenapi.UserDesc) (proto.UserToken, error)
	Get(token string) (*proto.UserToken, error)
	GetAll() ([]proto.UserToken, error)
	Clean() error
	Delete(token string) (bool, error)
}
