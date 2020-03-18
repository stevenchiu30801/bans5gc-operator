package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PLMN ID
type PlmnId struct {
	Mcc string `json:"mcc"`
	Mnc string `json:"mnc"`
}

// TAI
type Tai struct {
	PlmnId PlmnId `json:"plmnId"`
	Tac    string `json:"tac"`
}

// BansSliceSpec defines the desired state of BansSlice
type BansSliceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// S-NSSAI list
	SnssaiList []Snssai `json:"snssaiList"`

	// Minimum rate of bandiwdth in Mbps
	MinRate uint `json:"minRate"`

	// Maximum rate of bandiwdth in Mbps
	MaxRate uint `json:"maxRate"`

	// gNodeB address
	GNBAddr string `json:"gNBAddr"`

	// TAI
	Tai Tai `json:"tai"`
}

// BansSliceStatus defines the observed state of BansSlice
type BansSliceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BansSlice is the Schema for the bansslice API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=bansslice,scope=Namespaced
type BansSlice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BansSliceSpec   `json:"spec,omitempty"`
	Status BansSliceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BansSliceList contains a list of BansSlice
type BansSliceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BansSlice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BansSlice{}, &BansSliceList{})
}
