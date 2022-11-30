// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	commonv1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/common/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
const (
	LogstashContainerName = "logstash"
	// Kind is inferred from the struct name using reflection in SchemeBuilder.Register()
	// we duplicate it as a constant here for practical purposes.
	Kind = "Logstash"
)

// LogstashSpec defines the desired state of Logstash
type LogstashSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Version of Logstash
	Version string `json:"version"`

	// Image is the Logstash Docker image to deploy.
	Image string `json:"image,omitempty"`

	// Count of Logstash instances to deploy.
	Count int32 `json:"count,omitempty"`

	// ElasticsearchRef is a reference to an Elasticsearch cluster running in the same Kubernetes cluster.
	ElasticsearchRef commonv1.ObjectSelector `json:"elasticsearchRef,omitempty"`

	// Config holds the Logstash configuration.
	// +kubebuilder:pruning:PreserveUnknownFields
	Config *commonv1.Config `json:"config,omitempty"`

	// ConfigRef contains a reference to an existing Kubernetes Secret holding the Logstash configuration.
	// Configuration settings are merged and have precedence over settings specified in `config`.
	// +kubebuilder:validation:Optional
	ConfigRef *commonv1.ConfigSource `json:"configRef,omitempty"`

}

// LogstashStatus defines the observed state of Logstash
type LogstashStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Logstash is the Schema for the logstashes API
// +k8s:openapi-gen=true
type Logstash struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LogstashSpec   `json:"spec,omitempty"`
	Status LogstashStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LogstashList contains a list of Logstash
type LogstashList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Logstash `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Logstash{}, &LogstashList{})
}
