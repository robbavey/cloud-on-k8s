// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package logstash

import (
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/certificates"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/volume"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"

	logstashv1alpha1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/logstash/v1alpha1"
	//"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/certificates"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/container"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/defaults"
	//"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/volume"
)

const (
	HTTPPort                 = 9600
	configHashAnnotationName = "logstash.k8s.elastic.co/config-hash"
	ConfigFilename           = "logstash.yml"
	ConfigMountPath          = "/usr/share/logstash/config/logstash.yml"
)

var (
	DefaultMemoryLimits = resource.MustParse("1Gi")
	DefaultResources    = corev1.ResourceRequirements{
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceMemory: DefaultMemoryLimits,
		},
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceMemory: DefaultMemoryLimits,
		},
	}
)

// readinessProbe is the readiness probe for the maps container
func readinessProbe(useTLS bool) corev1.Probe {
	scheme := corev1.URISchemeHTTP
	//if useTLS {
	//	scheme = corev1.URISchemeHTTPS
	//}
	return corev1.Probe{
		FailureThreshold:    3,
		InitialDelaySeconds: 10,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		TimeoutSeconds:      5,
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Port:   intstr.FromInt(HTTPPort),
				Path:   "/",
				Scheme: scheme,
			},
		},
	}
}

func newPodSpec(logstash logstashv1alpha1.Logstash, configHash string) (corev1.PodTemplateSpec, error) {
	// ensure the Pod gets rotated on config change
	annotations := map[string]string{configHashAnnotationName: configHash}

	//defaultContainerPorts := []corev1.ContainerPort{
	//	{Name: logstash.Spec.HTTP.Protocol(), ContainerPort: int32(HTTPPort), Protocol: corev1.ProtocolTCP},
	//}

	// TODO: Make
	defaultContainerPorts := []corev1.ContainerPort{
		{Name: "http", ContainerPort: 9600, Protocol: corev1.ProtocolTCP},
	}

	// TODO: Add volumes for other logstash parts
	cfgVolume := configSecretVolume(logstash)

	builder := defaults.NewPodTemplateBuilder(logstash.Spec.PodTemplate, logstashv1alpha1.LogstashContainerName).
		WithAnnotations(annotations).
		WithResources(DefaultResources).
		WithDockerImage(logstash.Spec.Image, container.ImageRepository(container.LogstashImage, logstash.Spec.Version)).
		//WithReadinessProbe(readinessProbe(logstash.Spec.HTTP.TLS.Enabled())).
		WithPorts(defaultContainerPorts).
		WithVolumes(cfgVolume.Volume()).
		WithVolumeMounts(cfgVolume.VolumeMount()).
		WithInitContainerDefaults()

	builder, err := withESCertsVolume(builder, logstash)
	if err != nil {
		return corev1.PodTemplateSpec{}, err
	}

	//builder = withHTTPCertsVolume(builder, logstash)

	esAssocConf, err := logstash.AssociationConf()
	if err != nil {
		return corev1.PodTemplateSpec{}, err
	}
	if !esAssocConf.IsConfigured() {
		// supported as of 7.14, harmless on prior versions, but both Elasticsearch connection and this must not be specified
		builder = builder.WithEnv(corev1.EnvVar{Name: "ELASTICSEARCH_PREVALIDATED", Value: "true"})
	}

	return builder.PodTemplate, nil
}

func withESCertsVolume(builder *defaults.PodTemplateBuilder, logstash logstashv1alpha1.Logstash) (*defaults.PodTemplateBuilder, error) {
	esAssocConf, err := logstash.AssociationConf()
	if err != nil {
		return nil, err
	}
	if !esAssocConf.CAIsConfigured() {
		return builder, nil
	}
	vol := volume.NewSecretVolumeWithMountPath(
		esAssocConf.GetCASecretName(),
		"es-certs",
		ESCertsPath,
	)
	return builder.
		WithVolumes(vol.Volume()).
		WithVolumeMounts(vol.VolumeMount()), nil
}

func withHTTPCertsVolume(builder *defaults.PodTemplateBuilder, logstash logstashv1alpha1.Logstash) *defaults.PodTemplateBuilder {
	if !logstash.Spec.HTTP.TLS.Enabled() {
		return builder
	}
	vol := certificates.HTTPCertSecretVolume(LogstashNamer, logstash.Name)
	return builder.WithVolumes(vol.Volume()).WithVolumeMounts(vol.VolumeMount())
}
