package crd

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koomgr/apis/directory/v1alpha1"
	"golang.org/x/crypto/bcrypt"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

type crdProvider struct {
	*CrdProviderConfig
	kubeClient client.Client
	logger     logr.Logger
}

func (this *crdProvider) GetName() string {
	return this.Name
}

// As we have small chance to fail, we can be critical
func (this *crdProvider) IsCritical() bool {
	return true
}

func (this *crdProvider) GetUserStatus(login string, password string, checkPassword bool) (common.UserStatus, error) {
	userStatus := common.UserStatus{
		ProviderName:   this.Name,
		Authority:      *this.CredentialAuthority,
		Found:          false,
		PasswordStatus: common.Unchecked,
		Uid:            "",
		Groups:         []string{},
		Email:          "",
	}
	usr := v1alpha1.User{}
	err := this.kubeClient.Get(context.TODO(), client.ObjectKey{
		Namespace: this.Namespace,
		Name:      login,
	}, &usr)
	if client.IgnoreNotFound(err) != nil {
		return userStatus, err
	}
	if err != nil {
		this.logger.V(1).Info("User NOT found", "user", login)
		return userStatus, nil // User not found. Not an error
	}
	if usr.Spec.Disabled {
		this.logger.V(1).Info("User found but disabled", "user", login)
		return userStatus, nil // Act as if user was not found
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
		this.logger.Info(fmt.Sprintf("************ namespace:%s", this.Namespace))
		err = this.kubeClient.List(context.TODO(), &list, client.MatchingFields{"userkey": login}, client.InNamespace(this.Namespace))
		if err != nil {
			return userStatus, err
		}
		userStatus.Groups = make([]string, 0, len(list.Items))
		for i := 0; i < len(list.Items); i++ {
			binding := list.Items[i]
			this.logger.V(1).Info("lookup", "binding", binding.Name)
			if binding.Spec.Disabled {
				continue
			}
			grp := v1alpha1.Group{}
			this.logger.Info(fmt.Sprintf("************ namespace:%s   name:%s", this.Namespace, binding.Spec.Group))
			err := this.kubeClient.Get(context.TODO(), client.ObjectKey{
				Namespace: this.Namespace,
				Name:      binding.Spec.Group,
			}, &grp)
			if client.IgnoreNotFound(err) != nil {
				return userStatus, err
			}
			if err != nil {
				// Not found. Broken link
				//crdLog.Error(fmt.Errorf("Broken GroupBinding link "), "(No matching group)", "groupbinding", binding.Name, "group", binding.Spec.Group)
				this.logger.V(-1).Info("Broken GroupBinding link (No matching group)", "groupbinding", binding.Name, "group", binding.Spec.Group)
				continue
			}
			if grp.Spec.Disabled {
				continue
			}
			userStatus.Groups = append(userStatus.Groups, fmt.Sprintf(this.GroupPattern, grp.Name))
		}
	}
	return userStatus, nil
}
