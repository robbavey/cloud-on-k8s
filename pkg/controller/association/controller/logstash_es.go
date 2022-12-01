// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package controller

import (
	//"fmt"
	//"strings"

	//pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	logstashv1alpha1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/logstash/v1alpha1"
	commonv1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/common/v1"
	esv1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/elasticsearch/v1"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/association"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/operator"
	//"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/version"
	eslabel "github.com/elastic/cloud-on-k8s/v2/pkg/controller/elasticsearch/label"
	esuser "github.com/elastic/cloud-on-k8s/v2/pkg/controller/elasticsearch/user"
	"github.com/elastic/cloud-on-k8s/v2/pkg/utils/k8s"
	"github.com/elastic/cloud-on-k8s/v2/pkg/utils/rbac"
)

const (
	// LogstashAssociationLabelName marks resources created for an association originating from Logstash with the
	// Logstash name.
	LogstashAssociationLabelName = "logstashassociation.k8s.elastic.co/name"
	// LogstashAssociationLabelNamespace marks resources created for an association originating from Logstash with the
	// Logstash namespace.
	LogstashAssociationLabelNamespace = "logstashassociation.k8s.elastic.co/namespace"
	// AgentAssociationLabelType marks resources created for an association originating from Logstash
	// with the target resource type (e.g. "elasticsearch" or "kibana").
	LogstashAssociationLabelType = "logstashassociation.k8s.elastic.co/type"
)

func AddLogstashES(mgr manager.Manager, accessReviewer rbac.AccessReviewer, params operator.Parameters) error {
	return association.AddAssociationController(mgr, accessReviewer, params, association.AssociationInfo{
		AssociatedObjTemplate:     func() commonv1.Associated { return &logstashv1alpha1.Logstash{} },
		ReferencedObjTemplate:     func() client.Object { return &esv1.Elasticsearch{} },
		AssociationType:           commonv1.ElasticsearchAssociationType,
		ReferencedResourceVersion: referencedElasticsearchStatusVersion,
		ExternalServiceURL:        getElasticsearchExternalURL,
		ReferencedResourceNamer:   esv1.ESNamer,
		AssociationName:           "logstash-es",
		AssociatedShortName:       "logstash",
		Labels: func(associated types.NamespacedName) map[string]string {
			return map[string]string{
				LogstashAssociationLabelName:      associated.Name,
				LogstashAssociationLabelNamespace: associated.Namespace,
				LogstashAssociationLabelType:      commonv1.ElasticsearchAssociationType,
			}
		},
		AssociationConfAnnotationNameBase:     commonv1.ElasticsearchConfigAnnotationNameBase,
		AssociationResourceNameLabelName:      eslabel.ClusterNameLabelName,
		AssociationResourceNamespaceLabelName: eslabel.ClusterNamespaceLabelName,

		ElasticsearchUserCreation: &association.ElasticsearchUserCreation{
			ElasticsearchRef: func(c k8s.Client, association commonv1.Association) (bool, commonv1.ObjectSelector, error) {
				return true, association.AssociationRef(), nil
			},
			UserSecretSuffix: "logstash-user",
			ESUserRole:       getLogstashRoles,
		},
	})
}

func getLogstashRoles(assoc commonv1.Associated) (string, error) {
	//logstash, ok := assoc.(*logstashv1alpha1.Logstash)
	//if !ok {
	//	return "", pkgerrors.Errorf(
	//		"Logstash expected, got %s/%s",
	//		assoc.GetObjectKind().GroupVersionKind().Group,
	//		assoc.GetObjectKind().GroupVersionKind().Kind,
	//	)
	//}
	//
	////if strings.Contains(beat.Spec.Type, ",") {
	////	return "", fmt.Errorf("beat type %s should not contain a comma", beat.Spec.Type)
	////}
	////if _, ok := beatv1beta1.KnownTypes[beat.Spec.Type]; !ok {
	////	return fmt.Sprintf("eck_beat_es_%s_role", beat.Spec.Type), nil
	////}
	//
	//v, err := version.Parse(logstash.Spec.Version)
	//if err != nil {
	//	return "", err
	//}
	//if v.GTE(version.MinFor(8, 0, 0)){
	//
	//}
	// TODO: THis may be nonsense
	return esuser.LogstashUserRole, nil

	// Roles for supported Logstashs are based on:
	// TODO:
	// https://www.elastic.co/guide/en/beats/filebeat/current/feature-roles.html
	// Docs are the same for all Logstashs. For a specific version docs change "current" to major.minor, eg:
	// https://www.elastic.co/guide/en/beats/filebeat/7.1/feature-roles.html
	//switch {
	////TODO: Figure out logstashesrolename
	//case v.GTE(version.From(7, 7, 0)):
	//	return strings.Join([]string{
	//		"kibana_admin",
	//		"ingest_admin",
	//		"logstash_admin",
	//		"remote_monitoring_agent",
	//		esuser.LogstashEsRoleName(esuser.V77, beat.Spec.Type),
	//	}, ","), nil
	//case v.GTE(version.From(7, 5, 0)):
	//	return strings.Join([]string{
	//		"kibana_user",
	//		"ingest_admin",
	//		"logstash_admin",
	//		"remote_monitoring_agent",
	//		esuser.LogstashEsRoleName(esuser.V75, beat.Spec.Type),
	//	}, ","), nil
	//case v.GTE(version.From(7, 3, 0)):
	//	return strings.Join([]string{
	//		"kibana_user",
	//		"ingest_admin",
	//		"logstash_admin",
	//		"remote_monitoring_agent",
	//		esuser.LogstashEsRoleName(esuser.V73, beat.Spec.Type),
	//	}, ","), nil
	//default:
	//	return strings.Join([]string{
	//		"kibana_user",
	//		"ingest_admin",
	//		"logstash_admin",
	//		"monitoring_user",
	//		esuser.LogstashEsRoleName(esuser.V70, beat.Spec.Type),
	//	}, ","), nil
	//}
}
