/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"skas/sk-common/proto/v1/proto"
)

type TokenLifecycle struct {
	InactivityTimeout metav1.Duration `json:"inactivityTimeout"`
	MaxTTL            metav1.Duration `json:"maxTTL"`
	ClientTTL         metav1.Duration `json:"clientTTL"`
}

type TokenSpec struct {

	// +required
	Client string `json:"client"`

	// +required
	User proto.User `json:"user"`

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
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=skastoken
// +kubebuilder:printcolumn:name="Client",type=string,JSONPath=`.spec.client`
// +kubebuilder:printcolumn:name="User login",type=string,JSONPath=`.spec.user.login`
// +kubebuilder:printcolumn:name="User ID",type=string,JSONPath=`.spec.user.uid`
// +kubebuilder:printcolumn:name="User Groups",type=string,JSONPath=`.spec.user.groups`
// +kubebuilder:printcolumn:name="Last hit",type=string,JSONPath=`.status.lastHit`
type Token struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TokenSpec   `json:"spec,omitempty"`
	Status TokenStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type TokenList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Token `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Token{}, &TokenList{})
}
