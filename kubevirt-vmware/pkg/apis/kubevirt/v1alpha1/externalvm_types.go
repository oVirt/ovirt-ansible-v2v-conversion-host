package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ExternalVmSpec defines the desired state of ExternalVm
// +k8s:openapi-gen=true
type ExternalVmSpec struct {
	VmName        string         `json:"vmname"`
	Memory        int32          `json:"memory"`
	CPUs          int32          `json:"cpus"`
	CpuSockets    int32          `json:"cpusockets"`
	GuestFullName string         `json:"guestname"`
	PowerState    string         `json:"powerstate"`
	Disks         []ExternalDisk `json:"disks,omitempty"`
	Nics          []ExternalNic  `json:"nics,omitempty"`
}

// ExternalDisk defines the desired state of ExternalDisk
// +k8s:openapi-gen=true
type ExternalDisk struct {
	Label    string `json:"label,omitempty"`
	Filename string `json:"filename,omitempty"`
	Capacity int64  `json:"capacity,omitempty"`
}

// ExternalVmStatus defines the desired state of ExternalNic
// +k8s:openapi-gen=true
type ExternalNic struct {
	Label string `json:"label,omitempty"`
	Mac   string `json:"mac,omitempty"`
}

// ExternalVmStatus defines the observed state of ExternalVm
// +k8s:openapi-gen=true
type ExternalVmStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExternalVm is the Schema for the externalvms API
// +k8s:openapi-gen=true
type ExternalVm struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExternalVmSpec   `json:"spec,omitempty"`
	Status ExternalVmStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExternalVmList contains a list of ExternalVm
type ExternalVmList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExternalVm `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ExternalVm{}, &ExternalVmList{})
}
