// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package v1alpha1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
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

	// ServiceAccountName is used to check access from the current resource to a resource (for ex. Elasticsearch) in a different namespace.
	// Can only be used if ECK is enforcing RBAC on references.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// PodTemplate provides customisation options (labels, annotations, affinity rules, resource requests, and so on) for the Elastic Maps Server pods
	// +kubebuilder:validation:Optional
	// +kubebuilder:pruning:PreserveUnknownFields
	PodTemplate corev1.PodTemplateSpec `json:"podTemplate,omitempty"`

}

// LogstashStatus defines the observed state of Logstash
type LogstashStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	commonv1.DeploymentStatus `json:",inline"`

	AssociationStatus commonv1.AssociationStatus `json:"associationStatus,omitempty"`

	// ObservedGeneration is the most recent generation observed for this Elastic Maps Server.
	// It corresponds to the metadata generation, which is updated on mutation by the API Server.
	// If the generation observed in status diverges from the generation in metadata, the Elastic
	// Maps controller has not yet processed the changes contained in the Elastic Maps specification.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

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
	assocConf *commonv1.AssociationConf `json:"-"`
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

// IsMarkedForDeletion returns true if the Elastic Maps Server instance is going to be deleted
func (m *Logstash) IsMarkedForDeletion() bool {
	return !m.DeletionTimestamp.IsZero()
}

func (m *Logstash) Associated() commonv1.Associated {
	return m
}

func (m *Logstash) AssociationConfAnnotationName() string {
	return commonv1.ElasticsearchConfigAnnotationNameBase
}

func (m *Logstash) AssociationType() commonv1.AssociationType {
	return commonv1.ElasticsearchAssociationType
}

func (m *Logstash) AssociationRef() commonv1.ObjectSelector {
	return m.Spec.ElasticsearchRef.WithDefaultNamespace(m.Namespace)
}

func (m *Logstash) ServiceAccountName() string {
	return m.Spec.ServiceAccountName
}

func (m *Logstash) AssociationConf() (*commonv1.AssociationConf, error) {
	return commonv1.GetAndSetAssociationConf(m, m.assocConf)
}

func (m *Logstash) SetAssociationConf(assocConf *commonv1.AssociationConf) {
	m.assocConf = assocConf
}

// RequiresAssociation returns true if the spec specifies an Elasticsearch reference.
func (m *Logstash) RequiresAssociation() bool {
	return m.Spec.ElasticsearchRef.IsDefined()
}

func (m *Logstash) AssociationStatusMap(typ commonv1.AssociationType) commonv1.AssociationStatusMap {
	if typ == commonv1.ElasticsearchAssociationType && m.Spec.ElasticsearchRef.IsDefined() {
		return commonv1.NewSingleAssociationStatusMap(m.Status.AssociationStatus)
	}

	return commonv1.AssociationStatusMap{}
}

func (m *Logstash) SetAssociationStatusMap(typ commonv1.AssociationType, status commonv1.AssociationStatusMap) error {
	single, err := status.Single()
	if err != nil {
		return err
	}

	if typ != commonv1.ElasticsearchAssociationType {
		return fmt.Errorf("association type %s not known", typ)
	}

	m.Status.AssociationStatus = single
	return nil
}

func (m *Logstash) ElasticServiceAccount() (commonv1.ServiceAccountName, error) {
	return "", nil
}

func (m *Logstash) GetAssociations() []commonv1.Association {
	associations := make([]commonv1.Association, 0)
	if m.Spec.ElasticsearchRef.IsDefined() {
		associations = append(associations, m)
	}
	return associations
}

func (m *Logstash) AssociationID() string {
	return commonv1.SingletonAssociationID
}

var _ commonv1.Associated = &Logstash{}
var _ commonv1.Association = &Logstash{}

// GetObservedGeneration will return the observed generation from the Elastic Maps status.
func (m *Logstash) GetObservedGeneration() int64 {
	return m.Status.ObservedGeneration
}
