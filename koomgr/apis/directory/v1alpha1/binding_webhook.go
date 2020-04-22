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
var bindinglog = logf.Log.WithName("binding-resource")

func (r *Binding) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-directory-koobind-io-v1alpha1-binding,mutating=true,failurePolicy=fail,groups=directory.koobind.io,resources=bindings,verbs=create;update,versions=v1alpha1,name=mbinding.kb.io

var _ webhook.Defaulter = &Binding{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Binding) Default() {
	bindinglog.Info("default", "name", r.Name)
	// Nothing to do for now
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-directory-koobind-io-v1alpha1-binding,mutating=false,failurePolicy=fail,groups=directory.koobind.io,resources=bindings,versions=v1alpha1,name=vbinding.kb.io

var _ webhook.Validator = &Binding{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Binding) ValidateCreate() error {
	bindinglog.Info("validate create", "name", r.Name)
	return r.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Binding) ValidateUpdate(old runtime.Object) error {
	bindinglog.Info("validate update", "name", r.Name)
	return r.validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Binding) ValidateDelete() error {
	bindinglog.Info("validate delete", "name", r.Name)
	return nil
}

func (this *Binding) validate() error {
	if this.Namespace != config.Conf.Namespace {
		return fmt.Errorf("%s '%s': Invalid namespace '%s'. Should be '%s'", this.Kind, this.Name, this.Namespace, config.Conf.Namespace)
	}
	return nil
}