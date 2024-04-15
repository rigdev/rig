package v1alpha2

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProjectEnvironmentSpec struct {
	Project     string `json:"project"`
	Environment string `json:"environment"`
}

type OwnedGlobalResource struct {
	Ref *v1.TypedObjectReference `json:"ref"`
	// +kubebuilder:validation:Enum=created;failed;alreadyExists;unchanged;updated;changePending;deleted
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

type ProjectEnvironmentStatus struct {
	ObservedGeneration int64                 `json:"observedGeneration,omitempty"`
	OwnedResources     []OwnedGlobalResource `json:"ownedResources,omitempty"`
	Errors             []string              `json:"errors,omitempty"`
	CreatedNamespace   bool                  `json:"createdNamespace"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion
//+kubebuilder:resource:path=projectenvironments,scope=Cluster

type ProjectEnvironment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the specification of the Project.
	Spec ProjectEnvironmentSpec `json:"spec,omitempty"`

	// Status holds the status of the Project
	Status *ProjectEnvironmentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:storageversion

// ProjectEnvironmentList contains a list of Projects
type ProjectEnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ProjectEnvironment `json:"items"`
}

// func init() {
// 	SchemeBuilder.Register(&ProjectEnvironment{}, &ProjectEnvironmentList{})
// }
