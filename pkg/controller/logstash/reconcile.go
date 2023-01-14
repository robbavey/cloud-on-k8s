// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package logstash

import (
	"context"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/pkg/errors"

	logstashv1alpha1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/logstash/v1alpha1"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/deployment"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/reconciler"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/tracing"
	"github.com/elastic/cloud-on-k8s/v2/pkg/utils/k8s"
	"github.com/elastic/cloud-on-k8s/v2/pkg/utils/pointer"
)

func reconcilePodVehicle(params Params, podTemplate corev1.PodTemplateSpec) (*reconciler.Results, logstashv1alpha1.LogstashStatus) {
	defer tracing.Span(&params.Context)()
	results := reconciler.NewResult(params.Context)

	spec := params.Logstash.Spec
	name := DeploymentName(params.Logstash.Name)

	var toDelete client.Object
	var reconciliationFunc func(params ReconciliationParams) (int32, int32, error)
	switch {
	case spec.Deployment != nil:
		reconciliationFunc = reconcileDeployment
	}

	ready, desired, err := reconciliationFunc(ReconciliationParams{
		ctx:         params.Context,
		client:      params.Client,
		logstash:    params.Logstash,
		podTemplate: podTemplate,
	})

	if err != nil {
		return results.WithError(err), params.Status
	}

	// clean up the other one
	if err := params.Client.Get(params.Context, types.NamespacedName{
		Namespace: params.Logstash.Namespace,
		Name:      name,
	}, toDelete); err == nil {
		results.WithError(params.Client.Delete(params.Context, toDelete))
	} else if !apierrors.IsNotFound(err) {
		results.WithError(err)
	}

	var status logstashv1alpha1.LogstashStatus
	if status, err = calculateStatus(&params, ready, desired); err != nil {
		err = errors.Wrap(err, "while calculating status")
	}

	return results.WithError(err), status
}

func reconcileDeployment(rp ReconciliationParams) (int32, int32, error) {
	d := deployment.New(deployment.Params{
		Name:                 DeploymentName(rp.logstash.Name),
		Namespace:            rp.logstash.Namespace,
		Selector:             NewLabels(rp.logstash),
		Labels:               NewLabels(rp.logstash),
		PodTemplateSpec:      rp.podTemplate,
		Replicas:             pointer.Int32OrDefault(rp.logstash.Spec.Deployment.Replicas, int32(1)),
		RevisionHistoryLimit: rp.logstash.Spec.RevisionHistoryLimit,
		Strategy:             rp.logstash.Spec.Deployment.Strategy,
	})
	if err := controllerutil.SetControllerReference(&rp.logstash, &d, scheme.Scheme); err != nil {
		return 0, 0, err
	}

	reconciled, err := deployment.Reconcile(rp.ctx, rp.client, d, &rp.logstash)
	if err != nil {
		return 0, 0, err
	}

	return reconciled.Status.ReadyReplicas, reconciled.Status.Replicas, nil
}

// ReconciliationParams are the parameters used during an Elastic Logstash's reconciliation.
type ReconciliationParams struct {
	ctx         context.Context
	client      k8s.Client
	logstash    logstashv1alpha1.Logstash
	podTemplate corev1.PodTemplateSpec
}

// calculateStatus will calculate a new status from the state of the pods within the k8s cluster
// and will return any error encountered.
func calculateStatus(params *Params, ready, desired int32) (logstashv1alpha1.LogstashStatus, error) {
	logstash := params.Logstash
	status := params.Status

	pods, err := k8s.PodsMatchingLabels(params.Client, logstash.Namespace, map[string]string{NameLabelName: logstash.Name})
	if err != nil {
		return status, err
	}

	status.Version = common.LowestVersionFromPods(params.Context, status.Version, pods, VersionLabelName)
	status.AvailableNodes = ready
	status.ExpectedNodes = desired
	return status, nil
}

// updateStatus will update the Elastic Logstash's status within the k8s cluster, using the given Elastic Logstash and status.
func updateStatus(ctx context.Context, logstash logstashv1alpha1.Logstash, client client.Client, status logstashv1alpha1.LogstashStatus) error {
	if reflect.DeepEqual(logstash.Status, status) {
		return nil
	}
	logstash.Status = status
	return common.UpdateStatus(ctx, client, &logstash)
}
