// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package logstash

import (
	"context"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"

	logstashv1alpha1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/logstash/v1alpha1"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/expectations"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/keystore"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/operator"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/reconciler"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/tracing"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/watches"

	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/logstash/sset"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/logstash/stackmon"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/logstash/volume"
	"github.com/elastic/cloud-on-k8s/v2/pkg/utils/k8s"
	"github.com/elastic/cloud-on-k8s/v2/pkg/utils/log"

	ulog "github.com/elastic/cloud-on-k8s/v2/pkg/utils/log"
)

// Params are a set of parameters used during internal reconciliation of Logstash.
type Params struct {
	Context context.Context

	Client        k8s.Client
	EventRecorder record.EventRecorder
	Watches       watches.DynamicWatches

	Logstash logstashv1alpha1.Logstash
	Status   logstashv1alpha1.LogstashStatus

	OperatorParams    operator.Parameters
	KeystoreResources *keystore.Resources

	// Expectations control some expectations set on resources in the cache, in order to
	// avoid doing certain operations if the cache hasn't seen an up-to-date resource yet.
	Expectations *expectations.Expectations
}

// K8sClient returns the Kubernetes client.
func (p Params) K8sClient() k8s.Client {
	return p.Client
}

// Recorder returns the Kubernetes event recorder.
func (p Params) Recorder() record.EventRecorder {
	return p.EventRecorder
}

// DynamicWatches returns the set of stateful dynamic watches used during reconciliation.
func (p Params) DynamicWatches() watches.DynamicWatches {
	return p.Watches
}

// GetPodTemplate returns the configured pod template for the associated Elastic Logstash.
func (p *Params) GetPodTemplate() corev1.PodTemplateSpec {
	return p.Logstash.Spec.PodTemplate
}

// Logger returns the configured logger for use during reconciliation.
func (p *Params) Logger() logr.Logger {
	return log.FromContext(p.Context)
}

func newStatus(logstash logstashv1alpha1.Logstash) logstashv1alpha1.LogstashStatus {
	status := logstash.Status
	status.ObservedGeneration = logstash.Generation
	return status
}

func internalReconcile(params Params) (*reconciler.Results, logstashv1alpha1.LogstashStatus) {
	defer tracing.Span(&params.Context)()
	results := reconciler.NewResult(params.Context)

	_, err := reconcileServices(params)
	if err != nil {
		return results.WithError(err), params.Status
	}

	configHash := fnv.New32a()

	// reconcile beats config secrets if Stack Monitoring is defined
	if err := stackmon.ReconcileConfigSecrets(params.Context, params.Client, params.Logstash); err != nil {
		return results.WithError(err), params.Status
	}

	if err := reconcileConfig(params, configHash); err != nil {
		return results.WithError(err), params.Status
	}

	// We intentionally DO NOT pass the configHash here. We don't want to consider the pipeline definitions in the
	// hash of the config to ensure that a pipeline change does not automatically trigger a restart
	// of the pod, but allows Logstash's automatic reload of pipelines to take place
	if err := reconcilePipeline(params); err != nil {
		return results.WithError(err), params.Status
	}

	params.Logstash.Spec.VolumeClaimTemplates = volume.AppendDefaultPVCs(params.Logstash.Spec.VolumeClaimTemplates,
		params.Logstash.Spec.PodTemplate.Spec)

	if keystoreResources, err := reconcileKeystore(params, configHash); err != nil {
		return results.WithError(err), params.Status
	} else if keystoreResources != nil {
		params.KeystoreResources = keystoreResources
	}

	podTemplate, err := buildPodTemplate(params, configHash)
	if err != nil {
		return results.WithError(err), params.Status
	}
	return reconcileStatefulSet(params, podTemplate)
}

// expectationsSatisfied checks that resources in our local cache match what we expect.
// If not, it's safer to not move on with StatefulSets and Pods reconciliation.
// Continuing with the reconciliation at this point may lead to:
// - calling ES orchestration settings (zen1/zen2/allocation excludes) with wrong assumptions
// (eg. incorrect number of nodes or master-eligible nodes topology)
// - create or delete more than one master node at once
func (p *Params) expectationsSatisfied(ctx context.Context) (bool, string, error) {
	log := ulog.FromContext(ctx)
	// make sure the cache is up-to-date
	expectationsOK, reason, err := p.Expectations.Satisfied()
	if err != nil {
		return false, "Cache is not up to date", err
	}
	if !expectationsOK {
		log.V(1).Info("Cache expectations are not satisfied yet, re-queueing", "namespace", p.Logstash.Namespace, "ls_name", p.Logstash.Name, "reason", reason)
		return false, reason, nil
	}
	actualStatefulSets, err := sset.RetrieveActualStatefulSets(p.Client, k8s.ExtractNamespacedName(&p.Logstash))
	if err != nil {
		return false, "Cannot retrieve actual stateful sets", err
	}
	// make sure StatefulSet statuses have been reconciled by the StatefulSet controller
	pendingStatefulSetReconciliation := actualStatefulSets.PendingReconciliation()
	if len(pendingStatefulSetReconciliation) > 0 {
		log.V(1).Info("StatefulSets observedGeneration is not reconciled yet, re-queueing", "namespace", p.Logstash.Namespace, "ls_name", p.Logstash.Name)
		return false, fmt.Sprintf("observedGeneration is not reconciled yet for StatefulSets %s", strings.Join(pendingStatefulSetReconciliation.Names().AsSlice(), ",")), nil
	}
	return actualStatefulSets.PodReconciliationDone(ctx, p.Client)
}
