/*
  Copyright (C) 2023 Serge ALEXANDRE

  This file is part of Skas project

  Skas is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  Skas is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with Skas.  If not, see <http://www.gnu.org/licenses/>.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type UserSpec struct {
	// The user login is the Name of the resource.

	// The user common name(s).
	// +optional
	CommonNames []string `json:"commonNames,omitempty"`

	// The user email(s).
	// +optional
	Emails []string `json:"emails,omitempty"`

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
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=skuser;skasuser
// +kubebuilder:printcolumn:name="Common names",type=string,JSONPath=`.spec.commonNames`
// +kubebuilder:printcolumn:name="Emails",type=string,JSONPath=`.spec.emails`
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

//+kubebuilder:object:root=true

// UserList contains a list of User
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}

func init() {
	SchemeBuilder.Register(&User{}, &UserList{})
}
