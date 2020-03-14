package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Bandwidth Slice
type Slice struct {
	// Minimum rate of bandiwdth in Mbps
	MinRate *uint `json:"minRate,omitempty"`

	// Maximum rate of bandiwdth in Mbps
	MaxRate *uint `json:"maxRate,omitempty"`

	// Priority
	Priority *uint `json:"priority,omitempty"`

	// Matching Flows
	Flows []Flow `json:"flows"`
}

// Matching Flow
type Flow struct {
	SrcAddr  string `json:"srcAddr,omitempty"`
	DstAddr  string `json:"dstAddr,omitempty"`
	SrcPort  uint16 `json:"srcPort,omitempty"`
	DstPort  uint16 `json:"dstPort,omitempty"`
	Protocol uint8  `json:"protocol,omitempty"`
}

// BandwidthSliceSpec defines the desired state of BandwidthSlice
type BandwidthSliceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	Slices []Slice `json:"slices"`
}

// BandwidthSliceStatus defines the observed state of BandwidthSlice
type BandwidthSliceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BandwidthSlice is the Schema for the bandwidthslice API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=bandwidthslice,scope=Namespaced
type BandwidthSlice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BandwidthSliceSpec   `json:"spec,omitempty"`
	Status BandwidthSliceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BandwidthSliceList contains a list of BandwidthSlice
type BandwidthSliceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BandwidthSlice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BandwidthSlice{}, &BandwidthSliceList{})
}
