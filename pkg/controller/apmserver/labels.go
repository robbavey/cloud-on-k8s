// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package apmserver

import "github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/labels"

const (
	// ApmServerNameLabelName used to represent an ApmServer in k8s resources
	ApmServerNameLabelName = "apm.k8s.elastic.co/name"
	// Type represents the apm server type
	Type = "apm-server"
	// APMVersionLabelName used to propagate APMServer version from the spec to the pods
	APMVersionLabelName = "apm.k8s.elastic.co/version"
)

// NewLabels constructs a new set of labels for an ApmServer pod
func NewLabels(apmServerName string) map[string]string {
	return map[string]string{
		ApmServerNameLabelName: apmServerName,
		labels.TypeLabelName:   Type,
	}
}
