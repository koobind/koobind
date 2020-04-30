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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var grouplog = logf.Log.WithName("group-resource")

func (r *Group) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-directory-koobind-io-v1alpha1-group,mutating=true,failurePolicy=fail,groups=directory.koobind.io,resources=groups,verbs=create;update,versions=v1alpha1,name=mgroup.kb.io

var _ webhook.Defaulter = &Group{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Group) Default() {
	grouplog.Info("default", "name", r.Name)
	// Nothing to do for now
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-directory-koobind-io-v1alpha1-group,mutating=false,failurePolicy=fail,groups=directory.koobind.io,resources=groups,versions=v1alpha1,name=vgroup.kb.io

var _ webhook.Validator = &Group{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Group) ValidateCreate() error {
	grouplog.Info("validate create", "name", r.Name)
	return r.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Group) ValidateUpdate(old runtime.Object) error {
	grouplog.Info("validate update", "name", r.Name)
	return r.validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Group) ValidateDelete() error {
	grouplog.Info("validate delete", "name", r.Name)
	return nil
}

func (this *Group) validate() error {
	if !config.Conf.Namespaces[this.Namespace] {
		return fmt.Errorf("%s '%s': Invalid namespace '%s'. Should be one of '%v'", this.Kind, this.Name, this.Namespace, mapkeys2slice(config.Conf.Namespaces))
	}
	return nil
}
