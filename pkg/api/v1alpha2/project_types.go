package v1alpha2

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProjectSpec struct {
	Namespaces []string `json:"namespaces,omitempty"`
}

type OwnedGlobalResource struct {
	Ref *v1.TypedObjectReference `json:"ref"`
	// +kubebuilder:validation:Enum=created;failed;alreadyExists;unchanged;updated;changePending;deleted
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

type ProjectStatus struct {
	ObservedGeneration int64                 `json:"observedGeneration,omitempty"`
	OwnedResources     []OwnedGlobalResource `json:"ownedResources,omitempty"`
	Errors             []string              `json:"errors,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion
//+kubebuilder:resource:path=projects,scope=Cluster

type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the specification of the Project.
	Spec ProjectSpec `json:"spec,omitempty"`

	// Status holds the status of the Project
	Status *ProjectStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:storageversion

// ProjectList contains a list of Projects
type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Project `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Project{}, &ProjectList{})
}
