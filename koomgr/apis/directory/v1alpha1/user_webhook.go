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

package v1alpha1

import (
	"fmt"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"k8s.io/apimachinery/pkg/runtime"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var userlog = logf.Log.WithName("user-resource")

func (r *User) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-directory-koobind-io-v1alpha1-user,mutating=true,failurePolicy=fail,groups=directory.koobind.io,resources=users,verbs=create;update,versions=v1alpha1,name=muser.kb.io

var _ webhook.Defaulter = &User{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (this *User) Default() {
	userlog.Info("default", "name", this.Name, "namespace", this.Namespace)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-directory-koobind-io-v1alpha1-user,mutating=false,failurePolicy=fail,groups=directory.koobind.io,resources=users,versions=v1alpha1,name=vuser.kb.io

var _ webhook.Validator = &User{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *User) ValidateCreate() error {
	userlog.Info("validate create", "name", r.Name)
	return r.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *User) ValidateUpdate(old runtime.Object) error {
	userlog.Info("validate update", "name", r.Name)
	return r.validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *User) ValidateDelete() error {
	userlog.Info("validate delete", "name", r.Name)
	return nil
}

var emailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func (this *User) validate() error {
	if !config.Conf.CrdNamespaces.Has(this.Namespace) {
		return fmt.Errorf("%s '%s': Invalid namespace '%s'. Should be one of '%v'", this.Kind, this.Name, this.Namespace, utils.Set2stringSlice(config.Conf.CrdNamespaces))
	}
	if this.Spec.PasswordHash != "" {
		err := bcrypt.CompareHashAndPassword([]byte(this.Spec.PasswordHash), []byte("xxxxx"))
		if err != nil && err != bcrypt.ErrMismatchedHashAndPassword {
			return fmt.Errorf("Invalid passwordHash!")
		}
	}
	if this.Spec.Email != "" {
		if !emailRegexp.MatchString(this.Spec.Email) {
			return fmt.Errorf("Invalid Email")
		}
	}
	return nil
}
