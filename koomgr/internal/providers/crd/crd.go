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
	"github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koomgr/apis/directory/v1alpha1"
	"github.com/koobind/koobind/koomgr/internal/config"
	"golang.org/x/crypto/bcrypt"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

type CrdProvider struct {
	*CrdProviderConfig
	logger logr.Logger
}

func (this *CrdProvider) IsCritical() bool {
	return *this.Critical
}

func (this *CrdProvider) GetUserStatus(login string, password string, checkPassword bool) (common.UserStatus, error) {
	userStatus := common.UserStatus{
		ProviderName:   this.Name,
		Authority:      *this.CredentialAuthority,
		Found:          false,
		PasswordStatus: common.Unchecked,
		Uid:            "",
		Groups:         []string{},
		Email:          "",
		Messages:       make([]string, 0, 0),
	}
	usr := v1alpha1.User{}
	err := config.KubeClient.Get(context.TODO(), client.ObjectKey{
		Namespace: this.Namespace,
		Name:      login,
	}, &usr)
	if client.IgnoreNotFound(err) != nil {
		return userStatus, err
	}
	if err != nil {
		this.logger.V(1).Info("User NOT found", "user", login)
		// Check if there is some orphean GroupBindings
		list := v1alpha1.GroupBindingList{}
		err = config.KubeClient.List(context.TODO(), &list, client.MatchingFields{"userkey": login}, client.InNamespace(this.Namespace))
		if err != nil {
			return userStatus, err
		}
		for i := 0; i < len(list.Items); i++ {
			userStatus.Messages = append(userStatus.Messages, fmt.Sprintf("Orphean GroupBinding '%s' to '%s'", list.Items[i].Name, list.Items[i].Spec.Group))
		}
		return userStatus, nil // User not found. Not an error
	}
	if usr.Spec.Disabled {
		this.logger.V(1).Info("User found but disabled", "user", login)
		userStatus.Messages = append(userStatus.Messages, "User disabled")
		return userStatus, nil // User Disabled. Not an error
	}
	userStatus.Found = true
	if usr.Spec.Uid != nil {
		userStatus.Uid = strconv.Itoa(*usr.Spec.Uid + this.UidOffet)
	}
	userStatus.CommonName = usr.Spec.CommonName
	userStatus.Email = usr.Spec.Email
	if *this.CredentialAuthority && checkPassword && usr.Spec.PasswordHash != "" {
		err := bcrypt.CompareHashAndPassword([]byte(usr.Spec.PasswordHash), []byte(password))
		if err == nil {
			this.logger.V(1).Info("User found and password OK", "user", login)
			userStatus.PasswordStatus = common.Checked
		} else {
			this.logger.V(1).Info("User found but password failed", "user", login, "password", false)
			userStatus.PasswordStatus = common.Wrong
		}
	} else {
		if !*this.CredentialAuthority {
			this.logger.V(1).Info("User found, but not CredentialAuthority!", "user", login)
		} else if usr.Spec.PasswordHash == "" {
			userStatus.Messages = append(userStatus.Messages, "No password")
			this.logger.V(1).Info("User found, but no password defined!", "user", login)
			userStatus.Authority = false
		} else {
			this.logger.V(1).Info("User found, but password check was not required!", "user", login)
		}
		userStatus.PasswordStatus = common.Unchecked
	}
	// Will not collect groups if auth failed.
	if userStatus.PasswordStatus != common.Wrong && *this.GroupAuthority {
		list := v1alpha1.GroupBindingList{}
		err = config.KubeClient.List(context.TODO(), &list, client.MatchingFields{"userkey": login}, client.InNamespace(this.Namespace))
		if err != nil {
			return userStatus, err
		}
		userStatus.Groups = make([]string, 0, len(list.Items))
		for i := 0; i < len(list.Items); i++ {
			binding := list.Items[i]
			this.logger.V(1).Info("lookup", "binding", binding.Name)
			if binding.Spec.Disabled {
				userStatus.Messages = append(userStatus.Messages, fmt.Sprintf("GroupBinding '%s' disabled", binding.Name))
				continue
			}
			grp := v1alpha1.Group{}
			err := config.KubeClient.Get(context.TODO(), client.ObjectKey{
				Namespace: this.Namespace,
				Name:      binding.Spec.Group,
			}, &grp)
			if client.IgnoreNotFound(err) != nil {
				return userStatus, err
			}
			if err != nil {
				// Not found. Broken link
				this.logger.V(-1).Info("Broken GroupBinding link (No matching group)", "groupbinding", binding.Name, "group", binding.Spec.Group)
				userStatus.Messages = append(userStatus.Messages, fmt.Sprintf("No matching group '%s' for GroupBinding '%s'", binding.Spec.Group, binding.Name))
				continue
			}
			if grp.Spec.Disabled {
				userStatus.Messages = append(userStatus.Messages, fmt.Sprintf("Group '%s' disabled", grp.Name))
				continue
			}
			userStatus.Groups = append(userStatus.Groups, fmt.Sprintf(this.GroupPattern, grp.Name))
		}
	}
	return userStatus, nil
}
