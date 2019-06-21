package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AzureAgentPoolSpec defines the desired state of AzureAgentPool
// +k8s:openapi-gen=true
type AzureAgentPoolSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	Account     string `json:"account"`
	Project     string `json:"project"`
	AccessToken string `json:"accessToken"`
	AgentPool   string `json:"agentPool"`
}

// AzureAgentPoolStatus defines the observed state of AzureAgentPool
// +k8s:openapi-gen=true
type AzureAgentPoolStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AzureAgentPool is the Schema for the azureagentpools API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type AzureAgentPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureAgentPoolSpec   `json:"spec,omitempty"`
	Status AzureAgentPoolStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AzureAgentPoolList contains a list of AzureAgentPool
type AzureAgentPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureAgentPool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AzureAgentPool{}, &AzureAgentPoolList{})
}
