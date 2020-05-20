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
package providers

import (
	"github.com/koobind/koobind/common"
)

type Provider interface {
	// If checkPassword == true && password == "", then UserStatus will be Wrong if this Provider is CredentialAuthority. Unchecked if not.
	// For non ldap provider, if this provider is CredentialAuthority, but password is not defined for a user, then password will be unchecked.
	// For LDAP provider defined as CredentialAuthority, password is assumed to be always defined.
	GetUserStatus(login string, password string, checkPassword bool) (common.UserStatus, error)
	GetName() string
	// If critical, a failure will induce 'Invalid login'. Otherwhise, other providers will be used
	IsCritical() bool
}

type ProviderChain interface {
	Login(login, password string) (user common.User, loginOk bool, authenticator string, err error) // authenticator is the name of the provider who authenticate the user
	DescribeUser(login string) ([]common.UserStatus, error)
	String() string
}
