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
)

type GroupBindingSpec struct {
	// +kubebuilder:validation:MinLength=1
	// +required
	User string `json:"user"`

	// +kubebuilder:validation:MinLength=1
	// +required
	Group string `json:"group"`
}

// GroupBindingStatus defines the observed state of GroupBinding
type GroupBindingStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=gb;skgroupbinding;skgb;skasgb
// +kubebuilder:printcolumn:name="User",type=string,JSONPath=`.spec.user`
// +kubebuilder:printcolumn:name="Group",type=string,JSONPath=`.spec.group`

// GroupBinding is the Schema for the groupbindings API
type GroupBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupBindingSpec   `json:"spec,omitempty"`
	Status GroupBindingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GroupBindingList contains a list of GroupBinding
type GroupBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GroupBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GroupBinding{}, &GroupBindingList{})
}
