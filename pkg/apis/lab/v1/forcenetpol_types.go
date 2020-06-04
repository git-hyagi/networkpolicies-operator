package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ForceNetPolSpec defines the desired state of ForceNetPol
type ForceNetPolSpec struct {
	Projects []string `json:"projects"`
}

// ForceNetPolStatus defines the observed state of ForceNetPol
type ForceNetPolStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ForceNetPol is the Schema for the forcenetpols API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=forcenetpols,scope=Namespaced
// +genclient:nonNamespaced
type ForceNetPol struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ForceNetPolSpec   `json:"spec,omitempty"`
	Status ForceNetPolStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ForceNetPolList contains a list of ForceNetPol
type ForceNetPolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ForceNetPol `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ForceNetPol{}, &ForceNetPolList{})
}
