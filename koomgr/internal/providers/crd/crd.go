package crd

import (
	"context"
	"fmt"
	"github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koomgr/apis/directory/v1alpha1"
	"github.com/koobind/koobind/koomgr/internal/config"
	"golang.org/x/crypto/bcrypt"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

type crdProvider struct {
	*CrdProviderConfig
	kubeClient client.Client
}

var crdLog = ctrl.Log.WithName("crd")

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
		Found:          false,
		PasswordStatus: common.Unchecked,
		Uid:            "",
		Groups:         nil,
		Email:          "",
	}
	usr := v1alpha1.User{}
	err := this.kubeClient.Get(context.TODO(), client.ObjectKey{
		Namespace: config.Conf.Namespace,
		Name:      login,
	}, &usr)
	if client.IgnoreNotFound(err) != nil {
		return userStatus, err
	}
	if err != nil {
		crdLog.V(1).Info("User NOT found", "user", login)
		return userStatus, nil // User not found. Not an error
	}
	if usr.Spec.Disabled {
		crdLog.V(1).Info("User found but disabled", "user", login)
		return userStatus, nil // Act as if user was not found
	}
	userStatus.Found = true
	if usr.Spec.Uid != nil {
		userStatus.Uid = strconv.Itoa(*usr.Spec.Uid + this.UidOffet)
	}
	userStatus.Email = usr.Spec.Email
	if *this.CredentialAuthority && checkPassword && usr.Spec.PasswordHash != "" {
		err := bcrypt.CompareHashAndPassword([]byte(usr.Spec.PasswordHash), []byte(password))
		if err == nil {
			crdLog.V(1).Info("User found and password OK", "user", login)
			userStatus.PasswordStatus = common.Checked
		} else {
			crdLog.V(1).Info("User found but password failed", "user", login, "password", false)
			userStatus.PasswordStatus = common.Wrong
		}
	} else {
		if !*this.CredentialAuthority {
			crdLog.V(1).Info("User found, but not CredentialAuthority!", "user", login)
		} else if usr.Spec.PasswordHash == "" {
			crdLog.V(1).Info("User found, but no password defined!", "user", login)
		} else {
			crdLog.V(1).Info("User found, but password check was not required!", "user", login)
		}
		userStatus.PasswordStatus = common.Unchecked
	}
	list := v1alpha1.GroupBindingList{}
	err = this.kubeClient.List(context.TODO(), &list, client.MatchingFields{"userkey": login})
	if err != nil {
		return userStatus, err
	}
	userStatus.Groups = make([]string, 0, len(list.Items))
	for i := 0; i < len(list.Items); i++ {
		binding := list.Items[i]
		crdLog.V(1).Info("lookup", "binding", binding.Name)
		if binding.Spec.Disabled {
			continue
		}
		grp := v1alpha1.Group{}
		err := this.kubeClient.Get(context.TODO(), client.ObjectKey{
			Namespace: config.Conf.Namespace,
			Name:      binding.Spec.Group,
		}, &grp)
		if client.IgnoreNotFound(err) != nil {
			return userStatus, err
		}
		if err != nil {
			// Not found. Broken link
			crdLog.Error(nil, "Broken GroupBinding link (No matching group)", "groupBinding", binding.Name, "group", binding.Spec.Group)
			continue
		}
		if grp.Spec.Disabled {
			continue
		}
		userStatus.Groups = append(userStatus.Groups, fmt.Sprintf(this.GroupPattern, grp.Name))
	}
	return userStatus, nil
}
