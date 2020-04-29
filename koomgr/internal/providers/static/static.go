package static

import (
	"fmt"
	"github.com/koobind/koobind/common"
	"golang.org/x/crypto/bcrypt"
	ctrl "sigs.k8s.io/controller-runtime"
	"strconv"
)

type staticProvider struct {
	*StaticProviderConfig
	userByLogin map[string]User
}

var spLog = ctrl.Log.WithName("static")

func (this *staticProvider) GetName() string {
	return this.Name
}

// As we have small chance to fail, we can be critical
func (this *staticProvider) IsCritical() bool {
	return true
}

func (this *staticProvider) GetUserStatus(login string, password string, checkPassword bool) (common.UserStatus, error) {
	userStatus := common.UserStatus{
		ProviderName:   this.Name,
		Found:          false,
		PasswordStatus: common.Unchecked,
		Uid:            "",
		Groups:         nil,
		Email:          "",
	}
	user, exists := this.userByLogin[login]
	if exists {
		userStatus.Found = true
		if *this.CredentialAuthority && user.PasswordHash != "" && checkPassword {
			err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
			if err == nil {
				userStatus.PasswordStatus = common.Checked
				//this.logger.Debugf("User '%s' found. Login OK", login)

			} else {
				//this.logger.Debugf("User '%s' found. Login failed", login)
				spLog.V(1).Info("User found", "user", login)
				userStatus.PasswordStatus = common.Wrong
			}
		} else {
			if !*this.CredentialAuthority {
				//this.logger.Debugf("User '%s' found, but not CredentialAuthority!", login)
				spLog.V(1).Info("User found, but not CredentialAuthority!", "user", login)
			} else if user.PasswordHash == "" {
				//this.logger.Debugf("User '%s' found, but no password defined!", login)
				spLog.V(1).Info("User found, but no password defined!", "user", login)
			} else {
				//this.logger.Debugf("User '%s' found, but no password check was required!", login)
				spLog.V(1).Info("User found, but no password check was required!", "user", login)
			}
			userStatus.PasswordStatus = common.Unchecked
		}
		if user.Id != nil {
			userStatus.Uid = strconv.Itoa(*user.Id + this.UidOffet)
		}
		userStatus.Email = user.Email
		userStatus.Groups = make([]string, len(user.Groups))
		for i := 0; i < len(user.Groups); i++ {
			userStatus.Groups[i] = fmt.Sprintf(this.GroupPattern, user.Groups[i])
		}
	} else {
		//this.logger.Debugf("User '%s' NOT found!", login)
		spLog.V(1).Info("User NOT found!", "user", login)
		userStatus.Found = false
	}
	return userStatus, nil
}
