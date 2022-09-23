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

// UserSpec defines the desired state of User
type UserSpec struct {
	// The user login is the Name of the resource.

	// The user common name.
	// +optional
	CommonName string `json:"commonName,omitempty"`

	// The user email.
	// +optional
	Email string `json:"email,omitempty"`

	// The user password, Hashed. Using golang.org/x/crypto/bcrypt.GenerateFromPassword()
	// Is optional, in case we only enrich a user from another directory
	// +optional
	PasswordHash string `json:"passwordHash,omitempty"`

	// Numerical user id
	// +optional
	Uid *int `json:"uid,omitempty"`

	// Whatever extra information related to this user.
	// +optional
	Comment string `json:"comment,omitempty"`

	// Prevent this user to login. Even if this user is managed by an external provider (i.e LDAP)
	// +optional
	Disabled *bool `json:"disabled,omitempty"`
}

// UserStatus defines the observed state of User
type UserStatus struct {
}

// +kubebuilder:object:root=true

// +kubebuilder:resource:scope=Namespaced,shortName=koouser;kuser;koousers;kusers
// +kubebuilder:printcolumn:name="Common name",type=string,JSONPath=`.spec.commonName`
// +kubebuilder:printcolumn:name="Email",type=string,JSONPath=`.spec.email`
// +kubebuilder:printcolumn:name="Uid",type=integer,JSONPath=`.spec.uid`
// +kubebuilder:printcolumn:name="Comment",type=string,JSONPath=`.spec.comment`
// +kubebuilder:printcolumn:name="Disabled",type=boolean,JSONPath=`.spec.disabled`
// User is the Schema for the users API
type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserSpec   `json:"spec,omitempty"`
	Status UserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserList contains a list of User
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}

func init() {
	SchemeBuilder.Register(&User{}, &UserList{})
}
