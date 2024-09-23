package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    corev1 "k8s.io/api/core/v1"
)

// ClusterLimitRangeSpec defines the desired state of ClusterLimitRange
type ClusterLimitRangeSpec struct {
    IgnoredNamespaces []string             `json:"ignoredNamespaces,omitempty"`
    ApplyNamespaces   []string             `json:"applyNamespaces,omitempty"`
    Limits            []LimitRangeItem     `json:"limits"`
}

// LimitRangeItem defines the limits for the ClusterLimitRange
type LimitRangeItem struct {
    Type               corev1.LimitType           `json:"type"`
    Max                corev1.ResourceList        `json:"max,omitempty"`
    Min                corev1.ResourceList        `json:"min,omitempty"`
    Default            corev1.ResourceList        `json:"default,omitempty"`
    DefaultRequest     corev1.ResourceList        `json:"defaultRequest,omitempty"`
    MaxLimitRequestRatio corev1.ResourceList      `json:"maxLimitRequestRatio,omitempty"`
}

// ClusterLimitRangeStatus defines the observed state of ClusterLimitRange
type ClusterLimitRangeStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ClusterLimitRange is the Schema for the clusterlimitranges API
type ClusterLimitRange struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   ClusterLimitRangeSpec   `json:"spec,omitempty"`
    Status ClusterLimitRangeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterLimitRangeList contains a list of ClusterLimitRange
type ClusterLimitRangeList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []ClusterLimitRange `json:"items"`
}

func init() {
    SchemeBuilder.Register(&ClusterLimitRange{}, &ClusterLimitRangeList{})
}
