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

package crd

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	directoryapi "github.com/koobind/koobind/koomgr/apis/directory/v1alpha1"
	tokenapi "github.com/koobind/koobind/koomgr/apis/tokens/v1alpha1"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"golang.org/x/crypto/bcrypt"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

var _ providers.Provider = &CrdProvider{}

type CrdProvider struct {
	*CrdProviderConfig
	logger logr.Logger
}

func (this *CrdProvider) IsCritical() bool {
	return *this.Critical
}

func (this *CrdProvider) GetUserStatus(login string, password string, checkPassword bool) (tokenapi.UserEntry, error) {
	userEntry := tokenapi.UserEntry{
		ProviderName:   this.Name,
		Authority:      *this.CredentialAuthority,
		Found:          false,
		PasswordStatus: tokenapi.PasswordStatusUnchecked,
		Uid:            "",
		Groups:         []string{},
		Email:          "",
		Messages:       make([]string, 0, 0),
	}
	usr := directoryapi.User{}
	err := config.KubeClient.Get(context.TODO(), client.ObjectKey{
		Namespace: this.Namespace,
		Name:      login,
	}, &usr)
	if client.IgnoreNotFound(err) != nil {
		return userEntry, err
	}
	if err != nil {
		this.logger.V(1).Info("User NOT found", "user", login)
		// Check if there is some orphean GroupBindings
		list := directoryapi.GroupBindingList{}
		err = config.KubeClient.List(context.TODO(), &list, client.MatchingFields{"userkey": login}, client.InNamespace(this.Namespace))
		if err != nil {
			return userEntry, err
		}
		for i := 0; i < len(list.Items); i++ {
			userEntry.Messages = append(userEntry.Messages, fmt.Sprintf("Orphean GroupBinding '%s' to '%s'", list.Items[i].Name, list.Items[i].Spec.Group))
		}
		return userEntry, nil // User not found. Not an error
	}
	if usr.Spec.Disabled != nil && *usr.Spec.Disabled {
		this.logger.V(1).Info("User found but disabled", "user", login)
		userEntry.Messages = append(userEntry.Messages, "User disabled")
		return userEntry, nil // User Disabled. Not an error
	}
	userEntry.Found = true
	if usr.Spec.Uid != nil {
		userEntry.Uid = strconv.Itoa(*usr.Spec.Uid + this.UidOffet)
	}
	userEntry.CommonName = usr.Spec.CommonName
	userEntry.Email = usr.Spec.Email
	if *this.CredentialAuthority && checkPassword && usr.Spec.PasswordHash != "" {
		err := bcrypt.CompareHashAndPassword([]byte(usr.Spec.PasswordHash), []byte(password))
		if err == nil {
			this.logger.V(1).Info("User found and password OK", "user", login)
			userEntry.PasswordStatus = tokenapi.PasswordStatusChecked
		} else {
			this.logger.V(1).Info("User found but password failed", "user", login, "password", false)
			userEntry.PasswordStatus = tokenapi.PasswordStatusWrong
		}
	} else {
		if !*this.CredentialAuthority {
			this.logger.V(1).Info("User found, but not CredentialAuthority!", "user", login)
		} else if usr.Spec.PasswordHash == "" {
			userEntry.Messages = append(userEntry.Messages, "No password")
			this.logger.V(1).Info("User found, but no password defined!", "user", login)
			userEntry.Authority = false
		} else {
			this.logger.V(1).Info("User found, but password check was not required!", "user", login)
		}
		userEntry.PasswordStatus = tokenapi.PasswordStatusUnchecked
	}
	// Will not collect groups if auth failed.
	if userEntry.PasswordStatus != tokenapi.PasswordStatusWrong && *this.GroupAuthority {
		list := directoryapi.GroupBindingList{}
		err = config.KubeClient.List(context.TODO(), &list, client.MatchingFields{"userkey": login}, client.InNamespace(this.Namespace))
		if err != nil {
			return userEntry, err
		}
		userEntry.Groups = make([]string, 0, len(list.Items))
		for i := 0; i < len(list.Items); i++ {
			binding := list.Items[i]
			this.logger.V(1).Info("lookup", "binding", binding.Name)
			if binding.Spec.Disabled != nil && *binding.Spec.Disabled {
				userEntry.Messages = append(userEntry.Messages, fmt.Sprintf("GroupBinding '%s' disabled", binding.Name))
				continue
			}
			grp := directoryapi.Group{}
			err := config.KubeClient.Get(context.TODO(), client.ObjectKey{
				Namespace: this.Namespace,
				Name:      binding.Spec.Group,
			}, &grp)
			if client.IgnoreNotFound(err) != nil {
				return userEntry, err
			}
			if err != nil {
				// Not found. Broken link
				this.logger.V(-1).Info("Broken GroupBinding link (No matching group)", "groupbinding", binding.Name, "group", binding.Spec.Group)
				userEntry.Messages = append(userEntry.Messages, fmt.Sprintf("No matching group '%s' for GroupBinding '%s'", binding.Spec.Group, binding.Name))
				continue
			}
			if grp.Spec.Disabled != nil && *grp.Spec.Disabled {
				userEntry.Messages = append(userEntry.Messages, fmt.Sprintf("Group '%s' disabled", grp.Name))
				continue
			}
			userEntry.Groups = append(userEntry.Groups, fmt.Sprintf(this.GroupPattern, grp.Name))
		}
	}
	return userEntry, nil
}

func (this *CrdProvider) ChangePassword(user string, oldPassword string, newPassword string) error {
	crdUser := &directoryapi.User{}
	err := config.KubeClient.Get(context.TODO(), client.ObjectKey{
		Namespace: this.Namespace,
		Name:      user,
	}, crdUser)
	if err != nil {
		return err
	}
	err = bcrypt.CompareHashAndPassword([]byte(crdUser.Spec.PasswordHash), []byte(oldPassword))
	if err != nil {
		return providers.ErrorInvalidOldPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	crdUser.Spec.PasswordHash = string(hash)
	err = config.KubeClient.Update(context.TODO(), crdUser)
	if err != nil {
		return err
	}
	return nil
}
