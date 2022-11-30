// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package logstash

import (
	"context"
	//"path"
	//"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonv1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/common/v1"
	logstashv1alpha1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/logstash/v1alpha1"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/association"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common"
	//"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/certificates"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/driver"
	commonlabels "github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/labels"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/reconciler"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/settings"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/volume"
	"github.com/elastic/cloud-on-k8s/v2/pkg/utils/k8s"
	"github.com/elastic/cloud-on-k8s/v2/pkg/utils/net"
)

const (
	//ESCertsPath     = "/mnt/elastic-internal/es-certs"
	ConfigFilename  = "logstash.yml"
	ConfigMountPath = "/usr/src/app/server/config/logstash.yml"
)

func configSecretVolume(logstash logstashv1alpha1.Logstash) volume.SecretVolume {
	return volume.NewSecretVolume(Config(logstash.Name), "config", ConfigMountPath, ConfigFilename, 0444)
}

func reconcileConfig(ctx context.Context, driver driver.Interface, logstash logstashv1alpha1.Logstash, ipFamily corev1.IPFamily) (corev1.Secret, error) {
	cfg, err := newConfig(ctx, driver, logstash, ipFamily)
	if err != nil {
		return corev1.Secret{}, err
	}

	cfgBytes, err := cfg.Render()
	if err != nil {
		return corev1.Secret{}, err
	}

	// Reconcile the configuration in a secret
	expectedConfigSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: logstash.Namespace,
			Name:      Config(logstash.Name),
			Labels:    commonlabels.AddCredentialsLabel(labels(logstash)),
		},
		Data: map[string][]byte{
			ConfigFilename: cfgBytes,
		},
	}

	return reconciler.ReconcileSecret(ctx, driver.K8sClient(), expectedConfigSecret, &logstash)
}

func newConfig(ctx context.Context, d driver.Interface, logstash logstashv1alpha1.Logstash, ipFamily corev1.IPFamily) (*settings.CanonicalConfig, error) {
	cfg := settings.NewCanonicalConfig()

	inlineUserCfg, err := inlineUserConfig(logstash.Spec.Config)
	if err != nil {
		return cfg, err
	}

	refUserCfg, err := common.ParseConfigRef(d, &logstash, logstash.Spec.ConfigRef, ConfigFilename)
	if err != nil {
		return cfg, err
	}

	defaults := defaultConfig(ipFamily)
	tls := tlsConfig(logstash)
	assocCfg, err := associationConfig(ctx, d.K8sClient(), logstash)
	if err != nil {
		return cfg, err
	}
	err = cfg.MergeWith(inlineUserCfg, refUserCfg, defaults, tls, assocCfg)
	return cfg, err
}

func inlineUserConfig(cfg *commonv1.Config) (*settings.CanonicalConfig, error) {
	if cfg == nil {
		cfg = &commonv1.Config{}
	}
	return settings.NewCanonicalConfigFrom(cfg.Data)
}

func defaultConfig(ipFamily corev1.IPFamily) *settings.CanonicalConfig {
	return settings.MustCanonicalConfig(map[string]interface{}{
		"host": net.InAddrAnyFor(ipFamily).String(),
	})
}

func tlsConfig(logstash logstashv1alpha1.Logstash) *settings.CanonicalConfig {
	return settings.NewCanonicalConfig()
	//if !ems.Spec.HTTP.TLS.Enabled() {
	//	return settings.NewCanonicalConfig()
	//}
	//return settings.MustCanonicalConfig(map[string]interface{}{
	//	"ssl.enabled":     true,
	//	"ssl.certificate": path.Join(certificates.HTTPCertificatesSecretVolumeMountPath, certificates.CertFileName),
	//	"ssl.key":         path.Join(certificates.HTTPCertificatesSecretVolumeMountPath, certificates.KeyFileName),
	//})
}

func associationConfig(ctx context.Context, c k8s.Client, logstash logstashv1alpha1.Logstash) (*settings.CanonicalConfig, error) {
	cfg := settings.NewCanonicalConfig()
	assocConf, err := logstash.AssociationConf()
	if err != nil {
		return nil, err
	}
	if !assocConf.IsConfigured() {
		return cfg, nil
	}
	credentials, err := association.ElasticsearchAuthSettings(ctx, c, &logstash)
	if err != nil {
		return nil, err
	}
	// TODO: Make this the value in logstash.xml
	if err := cfg.MergeWith(settings.MustCanonicalConfig(map[string]string{
		"elasticsearch.host":     assocConf.URL,
		"elasticsearch.username": credentials.Username,
		"elasticsearch.password": credentials.Password,
	})); err != nil {
		return nil, err
	}

	// TODO: Setup CA
	//if assocConf.GetCACertProvided() {
	//	if err := cfg.MergeWith(settings.MustCanonicalConfig(map[string]interface{}{
	//		"elasticsearch.ssl.verificationMode":       "certificate",
	//		"elasticsearch.ssl.certificateAuthorities": filepath.Join(ESCertsPath, certificates.CAFileName),
	//	})); err != nil {
	//		return nil, err
	//	}
	//}
	return cfg, nil
}
