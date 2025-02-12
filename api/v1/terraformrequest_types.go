/*
Copyright 2025.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TerraformRequestSpec defines the desired state of TerraformRequest.
type TerraformRequestSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Inline terraform
	Inline string `json:"inline,omitempty"`
	// credentials for cloud provider
	CredentialsSecretName string `json:"credentialSecretName,omitempty"`
	// Terraform apply or plab
	Operation string `json:"operation,omitempty"`
}

// TerraformRequestStatus defines the observed state of TerraformRequest.
type TerraformRequestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Applied string `json:"applied"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// TerraformRequest is the Schema for the terraformrequests API.
type TerraformRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TerraformRequestSpec   `json:"spec,omitempty"`
	Status TerraformRequestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TerraformRequestList contains a list of TerraformRequest.
type TerraformRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TerraformRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TerraformRequest{}, &TerraformRequestList{})
}
