// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package logstash

import (
	"context"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/certificates"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/labels"
	ulog "github.com/elastic/cloud-on-k8s/v2/pkg/utils/log"
	"hash"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"path/filepath"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	commonv1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/common/v1"
	logstashv1alpha1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/logstash/v1alpha1"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/association"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/reconciler"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/settings"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/tracing"
	"github.com/elastic/cloud-on-k8s/v2/pkg/utils/k8s"
)

func reconcileConfig(params Params, configHash hash.Hash) *reconciler.Results {
	defer tracing.Span(&params.Context)()
	results := reconciler.NewResult(params.Context)

	cfgBytes, err := buildConfig(params)
	if err != nil {
		return results.WithError(err)
	}

	expected := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: params.Logstash.Namespace,
			Name:      ConfigSecretName(params.Logstash.Name),
			Labels:    labels.AddCredentialsLabel(NewLabels(params.Logstash)),
		},
		Data: map[string][]byte{
			ConfigFileName: cfgBytes,
		},
	}

	if _, err = reconciler.ReconcileSecret(params.Context, params.Client, expected, &params.Logstash); err != nil {
		return results.WithError(err)
	}

	_, _ = configHash.Write(cfgBytes)

	return results
}

func buildConfig(params Params) ([]byte, error) {
	existingCfg, err := getExistingConfig(params.Context, params.Client, params.Logstash)
	if err != nil {
		return nil, err
	}

	userProvidedCfg, err := getUserConfig(params)
	if err != nil {
		return nil, err
	}

	cfg, err := defaultConfig()
	if err != nil {
		return nil, err
	}

	associationCfg, err := associationConfig(params)
	if err != nil {
		return nil, err
	}

	// merge with user settings last so they take precedence
	if err = cfg.MergeWith(existingCfg, associationCfg, userProvidedCfg); err != nil {
		return nil, err
	}

	return cfg.Render()
}

// getExistingConfig retrieves the canonical config, if one exists
func getExistingConfig(ctx context.Context, client k8s.Client, logstash logstashv1alpha1.Logstash) (*settings.CanonicalConfig, error) {
	var secret corev1.Secret
	key := types.NamespacedName{
		Namespace: logstash.Namespace,
		Name:      ConfigName(logstash.Name),
	}
	err := client.Get(context.Background(), key, &secret)
	if err != nil && apierrors.IsNotFound(err) {
		ulog.FromContext(ctx).V(1).Info("Logstash config secret does not exist", "namespace", logstash.Namespace, "name", logstash.Name)
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	rawCfg, exists := secret.Data[ConfigFileName]
	if !exists {
		return nil, nil
	}

	cfg, err := settings.ParseConfig(rawCfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// getUserConfig extracts the config either from the spec `config` field or from the Secret referenced by spec
// `configRef` field.
func getUserConfig(params Params) (*settings.CanonicalConfig, error) {
	if params.Logstash.Spec.Config != nil {
		return settings.NewCanonicalConfigFrom(params.Logstash.Spec.Config.Data)
	}
	return common.ParseConfigRef(params, &params.Logstash, params.Logstash.Spec.ConfigRef, ConfigFileName)
}

// TODO: remove testing value
func defaultConfig() (*settings.CanonicalConfig, error) {
	settingsMap := map[string]interface{}{
		"node.name": "test",
	}

	return settings.MustCanonicalConfig(settingsMap), nil
}

func associationConfig(params Params) (*settings.CanonicalConfig, error) {

	allAssociations := params.Logstash.GetAssociations()

	var esAssociations []commonv1.Association
	for _, assoc := range allAssociations {
		if assoc.AssociationType() == commonv1.ElasticsearchAssociationType {
			esAssociations = append(esAssociations, assoc)
		}
	}

	cfg := settings.NewCanonicalConfig()
	for _, assoc := range esAssociations {
		assocConf, err := assoc.AssociationConf()
		if err != nil {
			return settings.NewCanonicalConfig(), err
		}
		if !assocConf.IsConfigured() {
			return settings.NewCanonicalConfig(), nil
		}

		credentials, err := association.ElasticsearchAuthSettings(params.Context, params.Client, assoc)
		if err != nil {
			return settings.NewCanonicalConfig(), err
		}

		if err := cfg.MergeWith(settings.MustCanonicalConfig(map[string]interface{}{
			"xpack.management.enabled":                "true",
			"xpack.management.elasticsearch.hosts":    assocConf.GetURL(),
			"xpack.management.elasticsearch.username": credentials.Username,
			"xpack.management.elasticsearch.password": credentials.Password,
		})); err != nil {
			return nil, err
		}

		if assocConf.GetCACertProvided() {
			if err := cfg.MergeWith(settings.MustCanonicalConfig(map[string]interface{}{
				"xpack.management.elasticsearch.ssl.certificate_authority": filepath.Join(certificatesDir(assoc), certificates.CAFileName),
			})); err != nil {
				return nil, err
			}
		}

	}

	return cfg, nil
}

func ReconcileConfigMap(
	ctx context.Context,
	c k8s.Client,
	expected corev1.ConfigMap,
	logstash *logstashv1alpha1.Logstash,
) error {
	reconciled := &corev1.ConfigMap{}
	return reconciler.ReconcileResource(
		reconciler.Params{
			Context:    ctx,
			Client:     c,
			Owner:      logstash,
			Expected:   &expected,
			Reconciled: reconciled,
			NeedsUpdate: func() bool {
				return !reflect.DeepEqual(expected.Data, reconciled.Data)
			},
			UpdateReconciled: func() {
				reconciled.Data = expected.Data
			},
		},
	)
}
