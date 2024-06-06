/*
Copyright 2024.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NamespaceCleanerSpec defines the desired state of NamespaceCleaner
type NamespaceCleanerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// How long should namespace with no resource to be considered as stale
	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:Minimum=180
	// +kubebuilder:example=190
	// +kubebuilder:validation:Required
	DeleteDeadLineSeconds *int `json:"deletedeadlineseconds"`

	// List of Excluded Namespace from stale deletion
	// +kubebuilder:validation:Optional
	// kubebuilder:example=[default,foo,bar]
	ExcludeNamespace []string `json:"excludenamespace,omitempty"`
}

// NamespaceCleanerStatus defines the observed state of NamespaceCleaner
type NamespaceCleanerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	DeleteCount int `json:"deletecount"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NamespaceCleaner is the Schema for the namespacecleaners API
type NamespaceCleaner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NamespaceCleanerSpec   `json:"spec,omitempty"`
	Status NamespaceCleanerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NamespaceCleanerList contains a list of NamespaceCleaner
type NamespaceCleanerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NamespaceCleaner `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NamespaceCleaner{}, &NamespaceCleanerList{})
}
