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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var groupbindinglog = logf.Log.WithName("groupbinding-resource")

func (r *GroupBinding) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-directory-koobind-io-v1alpha1-groupbinding,mutating=true,failurePolicy=fail,groups=directory.koobind.io,resources=groupbindings,verbs=create;update,versions=v1alpha1,name=mgroupbinding.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Defaulter = &GroupBinding{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *GroupBinding) Default() {
	groupbindinglog.V(1).Info("default", "name", r.Name)
	// Nothing to do for now
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-directory-koobind-io-v1alpha1-groupbinding,mutating=false,failurePolicy=fail,groups=directory.koobind.io,resources=groupbindings,versions=v1alpha1,name=vgroupbinding.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Validator = &GroupBinding{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *GroupBinding) ValidateCreate() error {
	groupbindinglog.V(1).Info("validate create", "name", r.Name)
	return r.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *GroupBinding) ValidateUpdate(old runtime.Object) error {
	groupbindinglog.V(1).Info("validate update", "name", r.Name)
	return r.validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *GroupBinding) ValidateDelete() error {
	groupbindinglog.V(1).Info("validate delete", "name", r.Name)
	return nil
}

func (this *GroupBinding) validate() error {
	//if !config.Conf.CrdNamespaces.Has(this.Namespace) {
	//	return fmt.Errorf("%s '%s': Invalid namespace '%s'. Should be one of '%v'", this.Kind, this.Name, this.Namespace, config.Conf.CrdNamespaces.AsList())
	//}
	return nil
}
