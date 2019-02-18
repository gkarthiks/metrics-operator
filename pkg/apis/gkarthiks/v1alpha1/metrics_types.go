package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MetricsSpec defines the desired state of Metrics
// +k8s:openapi-gen=true
type MetricsSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Size int32 `json:"size"`
}

// MetricsStatus defines the observed state of Metrics
// +k8s:openapi-gen=true
type MetricsStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Nodes []string `json:"nodes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Metrics is the Schema for the metrics API
// +k8s:openapi-gen=true
type Metrics struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MetricsSpec   `json:"spec,omitempty"`
	Status MetricsStatus `json:"status,omitempty"`

	Prometheus struct {
		Image           string `json:"image,required"`
		ImagePullPolicy string `json:"imagePullPolicy,omitempty"`
		Resources       struct {
			Limits struct {
				CPU    string `json:"cpu,omitempty"`
				Memory string `json:"memory,omitempty"`
			}
			Requests struct {
				CPU    string `json:"cpu,omitempty"`
				Memory string `json:"memory,omitempty"`
			}
		}
	}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MetricsList contains a list of Metrics
type MetricsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Metrics `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Metrics{}, &MetricsList{})
}
