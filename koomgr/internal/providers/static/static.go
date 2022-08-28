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

package static

import (
	"fmt"
	tokenapi "github.com/koobind/koobind/koomgr/apis/tokens/v1alpha1"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"golang.org/x/crypto/bcrypt"
	ctrl "sigs.k8s.io/controller-runtime"
	"strconv"
)

var _ providers.Provider = &staticProvider{}

type staticProvider struct {
	*StaticProviderConfig
	userByLogin map[string]User
}

var spLog = ctrl.Log.WithName("static")

func (this *staticProvider) IsCritical() bool {
	return *this.Critical
}

func (this *staticProvider) GetUserStatus(login string, password string, checkPassword bool) (tokenapi.UserEntry, error) {
	userEntry := tokenapi.UserEntry{
		ProviderName:   this.Name,
		Authority:      *this.CredentialAuthority,
		Found:          false,
		PasswordStatus: tokenapi.PasswordStatusUnchecked,
		Uid:            "",
		Groups:         nil,
		Email:          "",
	}
	user, exists := this.userByLogin[login]
	if exists {
		userEntry.Found = true
		if *this.CredentialAuthority && user.PasswordHash != "" && checkPassword {
			err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
			if err == nil {
				userEntry.PasswordStatus = tokenapi.PasswordStatusChecked
				//this.logger.Debugf("User '%s' found. Login OK", login)
			} else {
				//this.logger.Debugf("User '%s' found. Login failed", login)
				spLog.V(1).Info("User found", "user", login)
				userEntry.PasswordStatus = tokenapi.PasswordStatusWrong
			}
		} else {
			if !*this.CredentialAuthority {
				spLog.V(1).Info("User found, but not CredentialAuthority!", "user", login)
			} else if user.PasswordHash == "" {
				spLog.V(1).Info("User found, but no password defined!", "user", login)
				userEntry.Authority = false
			} else {
				spLog.V(1).Info("User found, but no password check was required!", "user", login)
			}
			userEntry.PasswordStatus = tokenapi.PasswordStatusUnchecked
		}
		if user.Id != nil {
			userEntry.Uid = strconv.Itoa(*user.Id + this.UidOffet)
		}
		userEntry.Email = user.Email
		userEntry.CommonName = user.CommonName
		// Will not collect groups if auth failed
		if userEntry.PasswordStatus != tokenapi.PasswordStatusWrong && *this.GroupAuthority {
			userEntry.Groups = make([]string, len(user.Groups))
			for i := 0; i < len(user.Groups); i++ {
				userEntry.Groups[i] = fmt.Sprintf(this.GroupPattern, user.Groups[i])
			}
		}
	} else {
		//this.logger.Debugf("User '%s' NOT found!", login)
		spLog.V(1).Info("User NOT found!", "user", login)
		userEntry.Found = false
	}
	return userEntry, nil
}

func (this *staticProvider) ChangePassword(user string, oldPassword string, newPassword string) error {
	return providers.ErrorChangePasswordNotSupported
}
