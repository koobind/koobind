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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GroupBindingSpec defines the desired state of GroupBinding
type GroupBindingSpec struct {
	// +kubebuilder:validation:MinLength=1
	// +required
	User string `json:"user"`

	// +kubebuilder:validation:MinLength=1
	// +required
	Group string `json:"group"`

	// Allow to 'hide' this groupBinding. Even if defined in another provider
	// +optional
	Disabled bool `json:"disabled,omitempty"`
}

// GroupBindingStatus defines the observed state of GroupBinding
type GroupBindingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced,shortName=gb;ugb;usergroupbinding
// +kubebuilder:printcolumn:name="User",type=string,JSONPath=`.spec.user`
// +kubebuilder:printcolumn:name="Group",type=string,JSONPath=`.spec.group`
// +kubebuilder:printcolumn:name="Disabled",type=boolean,JSONPath=`.spec.disabled`
// GroupBinding is the Schema for the groupBindings API
type GroupBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupBindingSpec   `json:"spec,omitempty"`
	Status GroupBindingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GroupBindingList contains a list of GroupBinding
type GroupBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GroupBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GroupBinding{}, &GroupBindingList{})
}
