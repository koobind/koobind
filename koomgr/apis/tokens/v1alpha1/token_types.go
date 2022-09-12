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

type PasswordStatus string

const PasswordStatusUnchecked PasswordStatus = "unchecked"
const PasswordStatusChecked PasswordStatus = "checked"
const PasswordStatusWrong PasswordStatus = "wrong"

// UserEntry is the user definition provided by a given provider
type UserEntry struct {
	ProviderName   string         `json:"provider"`  // Used for 'describe' command
	Authority      bool           `json:"authority"` // Is this provider Authority for authentication (password) for this user (A password is defined)
	Found          bool           `json:"found"`
	PasswordStatus PasswordStatus `json:"passwordStatus"` // For describe, always 'unchecked'
	Uid            string         `json:"uid"`            // Issued from the authoritative server (The first one which checked the password).
	Groups         []string       `json:"groups"`
	Email          string         `json:"email"`
	CommonName     string         `json:"commonName"`
	Messages       []string       `json:"messages"` // To report error or explanation i.e broken link in crd provider, or disabled link
}

// User is the consolidated description of a user
type UserDesc struct {
	Name        string      `json:"name"`
	Uid         string      `json:"uid"`
	Groups      []string    `json:"groups"`
	Emails      []string    `json:"emails"`
	CommonNames []string    `json:"commonNames"`
	Authority   string      `json:"authority"` // The provider who validated the user password
	Entries     []UserEntry `json:"userEntries"`
}

type TokenLifecycle struct {
	InactivityTimeout metav1.Duration `json:"inactivityTimeout"`
	MaxTTL            metav1.Duration `json:"maxTTL"`
	ClientTTL         metav1.Duration `json:"clientTTL"`
}

// K8s Name will be the token itself
type TokenSpec struct {

	// +required
	User UserDesc `json:"user"`

	// +required
	Creation metav1.Time `json:"creation"`

	// +required
	Lifecycle TokenLifecycle `json:"lifecycle"`
}

// TokenStatus defines the observed state of Token
type TokenStatus struct {
	LastHit metav1.Time `json:"lastHit"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced,shortName=ktoken;kootoken
// +kubebuilder:printcolumn:name="User name",type=string,JSONPath=`.spec.user.name`
// +kubebuilder:printcolumn:name="User ID",type=string,JSONPath=`.spec.user.uid`
// +kubebuilder:printcolumn:name="User Groups",type=string,JSONPath=`.spec.user.groups`
// +kubebuilder:printcolumn:name="Last hit",type=string,JSONPath=`.status.lastHit`
// Token is the Schema for the tokens API
type Token struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TokenSpec   `json:"spec,omitempty"`
	Status TokenStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// TokenList contains a list of Token
type TokenList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Token `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Token{}, &TokenList{})
}
