// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package logstash

import (
	"context"

	//"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/certificates"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/events"

	//"reflect"
	"fmt"
	"hash/fnv"

	"go.elastic.co/apm/v2"

	//"sync/atomic"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	//"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	//logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	//ulog "github.com/elastic/cloud-on-k8s/v2/pkg/utils/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	logstashv1alpha1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/logstash/v1alpha1"

	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/association"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common"

	commonassociation "github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/association"
	//"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/certificates"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/defaults"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/deployment"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/driver"

	//"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/events"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/license"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/operator"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/reconciler"

	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/tracing"

	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/version"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/watches"
	"github.com/elastic/cloud-on-k8s/v2/pkg/utils/k8s"
	ulog "github.com/elastic/cloud-on-k8s/v2/pkg/utils/log"
	"github.com/elastic/cloud-on-k8s/v2/pkg/utils/maps"
)

const (
	controllerName = "logstash-controller"
)


// Add creates a new Logstash Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// // and Start it when the Manager is Started.
func Add(mgr manager.Manager, params operator.Parameters) error {
	reconciler := newReconciler(mgr, params)
	c, err := common.NewController(mgr, controllerName, reconciler, params)
	if err != nil {
		return err
	}
	return addWatches(c, reconciler)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, params operator.Parameters) *ReconcileLogstash {
	client := mgr.GetClient()
	return &ReconcileLogstash{
		Client:         client,
		recorder:       mgr.GetEventRecorderFor(controllerName),
		dynamicWatches: watches.NewDynamicWatches(),
		licenseChecker: license.NewLicenseChecker(client, params.OperatorNamespace),
		Parameters:     params,
	}
}

func addWatches(c controller.Controller, r *ReconcileLogstash) error {
	// Watch for changes to Logstash
	if err := c.Watch(&source.Kind{Type: &logstashv1alpha1.Logstash{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}

	// Watch deployments
	if err := c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &logstashv1alpha1.Logstash{},
	}); err != nil {
		return err
	}

	// Watch Pods, to ensure `status.version` and version upgrades are correctly reconciled on any change.
	// Watching Deployments only may lead to missing some events.
	if err := watches.WatchPods(c, NameLabelName); err != nil {
		return err
	}

	// Watch services
	if err := c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &logstashv1alpha1.Logstash{},
	}); err != nil {
		return err
	}

	// Watch owned and soft-owned secrets
	if err := c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &logstashv1alpha1.Logstash{},
	}); err != nil {
		return err
	}
	if err := watches.WatchSoftOwnedSecrets(c, logstashv1alpha1.Kind); err != nil {
		return err
	}

	// Dynamically watch referenced secrets to connect to Elasticsearch
	return c.Watch(&source.Kind{Type: &corev1.Secret{}}, r.dynamicWatches.Secrets)
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("logstash-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Logstash
	err = c.Watch(&source.Kind{Type: &logstashv1alpha1.Logstash{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by Logstash - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &logstashv1alpha1.Logstash{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileLogstash{}

// ReconcileLogstash reconciles a Logstash object
type ReconcileLogstash struct {
	k8s.Client
	operator.Parameters
	recorder       record.EventRecorder
	dynamicWatches watches.DynamicWatches
	licenseChecker license.Checker
	// iteration is the number of times this controller has run its Reconcile method
	iteration uint64
}

func (r *ReconcileLogstash) K8sClient() k8s.Client {
	return r.Client
}

func (r *ReconcileLogstash) DynamicWatches() watches.DynamicWatches {
	return r.dynamicWatches
}

func (r *ReconcileLogstash) Recorder() record.EventRecorder {
	return r.recorder
}

var _ driver.Interface = &ReconcileLogstash{}

// Reconcile reads that state of the cluster for a Logstash object and makes changes based on the state read and what is
// in the Logstash.Spec
func (r *ReconcileLogstash) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := ulog.FromContext(ctx)
	ctx = common.NewReconciliationContext(ctx, &r.iteration, r.Tracer, controllerName, "logstash_name", request)
	defer common.LogReconciliationRun(ulog.FromContext(ctx))()
	defer tracing.EndContextTransaction(ctx)
	//
	//// retrieve the Logstash object
	var logstash logstashv1alpha1.Logstash
	if err := r.Client.Get(ctx, request.NamespacedName, &logstash); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, r.onDelete(ctx,
				types.NamespacedName{
					Namespace: request.Namespace,
					Name:      request.Name,
				})
		}
		return reconcile.Result{}, tracing.CaptureError(ctx, err)
	}
	//
	log.Info("Retrieved", "logstash", logstash)
	if common.IsUnmanaged(ctx, &logstash) {
		log.Info("Object is currently not managed by this controller. Skipping reconciliation", "namespace", logstash.Namespace, "logstash_name", logstash.Name)
		return reconcile.Result{}, nil
	}
	//
	//// Logstash will be deleted nothing to do other than remove the watches
	if logstash.IsMarkedForDeletion() {
		return reconcile.Result{}, r.onDelete(ctx, k8s.ExtractNamespacedName(&logstash))
	}
	//
	//// main reconciliation logic
	results, status := r.doReconcile(ctx, logstash)
	if err := r.updateStatus(ctx, logstash, status); err != nil {
		ulog.FromContext(ctx).Info("ERROR updating status", "err", err)
		if apierrors.IsConflict(err) {
			return results.WithResult(reconcile.Result{Requeue: true}).Aggregate()
		}
		results.WithError(err)
	}
	return results.Aggregate()
}

func (r *ReconcileLogstash) doReconcile(ctx context.Context, logstash logstashv1alpha1.Logstash) (*reconciler.Results, logstashv1alpha1.LogstashStatus) {
	log := ulog.FromContext(ctx)
	results := reconciler.NewResult(ctx)
	status := newStatus(logstash)

	// TODO: If Logstash requires licensing reinstate this section.
	//enabled, err := r.licenseChecker.EnterpriseFeaturesEnabled(ctx)
	//if err != nil {
	//	return results.WithError(err), status
	//}

	//if !enabled {
	//	msg := "Logstash on ECK is an enterprise feature. Enterprise features are disabled"
	//	log.Info(msg, "namespace", logstash.Namespace, "logstash_name", ems.Name)
	//	r.recorder.Eventf(&logstash, corev1.EventTypeWarning, events.EventReconciliationError, msg)
	//	// we don't have a good way of watching for the license level to change so just requeue with a reasonably long delay
	//	return results.WithResult(reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Minute}), status
	//}


	isEsAssocConfigured, err := association.IsConfiguredIfSet(ctx, &logstash, r.recorder)
	if err != nil {
		return results.WithError(err), status
	}
	if !isEsAssocConfigured {
		return results, status
	}

	// Run validation in case the webhook is disabled
	if err := r.validate(ctx, logstash); err != nil {
		return results.WithError(err), status
	}

	svc, err := common.ReconcileService(ctx, r.Client, NewService(logstash), &logstash)
	if err != nil {
		return results.WithError(err), status
	}

	//TODO: THis is a hack to use svc
	ulog.FromContext(ctx).Info("This is the service!", "service", svc)

	//_, results = certificates.Reconciler{
	//	K8sClient:             r.K8sClient(),
	//	DynamicWatches:        r.DynamicWatches(),
	//	Owner:                 &logstash,
	//	TLSOptions:            logstash.Spec.HTTP.TLS,
	//	Namer:                 LogstashNamer,
	//	Labels:                labels(logstash),
	//	Services:              []corev1.Service{*svc},
	//	GlobalCA:              r.GlobalCA,
	//	CACertRotation:        r.CACertRotation,
	//	CertRotation:          r.CertRotation,
	//	GarbageCollectSecrets: true,
	//}.ReconcileCAAndHTTPCerts(ctx)
	//if results.HasError() {
	//	_, err := results.Aggregate()
	//	k8s.EmitErrorEvent(r.recorder, err, &logstash, events.EventReconciliationError, "Certificate reconciliation error: %v", err)
	//	return results, status
	//}

	logstashVersion, err := version.Parse(logstash.Spec.Version)
	if err != nil {
		return results.WithError(err), status
	}
	assocAllowed, err := association.AllowVersion(logstashVersion, logstash.Associated(), log, r.recorder)
	if err != nil {
		return results.WithError(err), status
	}
	if !assocAllowed {
		// will eventually retry once updated, along with the results
		// from the certificate reconciliation having a retry after a time period
		return results, status
	}

	configSecret, err := reconcileConfig(ctx, r, logstash, r.IPFamily)
	if err != nil {
		return results.WithError(err), status
	}

	// build a hash of various inputs to rotate Pods on any change
	configHash, err := buildConfigHash(r.K8sClient(), logstash, configSecret)
	if err != nil {
		return results.WithError(fmt.Errorf("build config hash: %w", err)), status
	}

	deploy, err := r.reconcileDeployment(ctx, logstash, configHash)
	if err != nil {
		return results.WithError(fmt.Errorf("reconcile deployment: %w", err)), status
	}

	status, err = r.getStatus(ctx, logstash, deploy)
	if err != nil {
		return results.WithError(fmt.Errorf("calculating status: %w", err)), status
	}

	return results, status
}

func newStatus(logstash logstashv1alpha1.Logstash) logstashv1alpha1.LogstashStatus {
	status := logstash.Status
	status.ObservedGeneration = logstash.Generation
	return status
}

func (r *ReconcileLogstash) validate(ctx context.Context, logstash logstashv1alpha1.Logstash) error {

	span, vctx := apm.StartSpan(ctx, "validate", tracing.SpanTypeApp)
	defer span.End()
	//
	if err := logstash.ValidateCreate(); err != nil {
		ulog.FromContext(ctx).Error(err, "Validation failed")
		k8s.EmitErrorEvent(r.recorder, err, &logstash, events.EventReasonValidation, err.Error())
		return tracing.CaptureError(vctx, err)
	}

	return nil
}

// Create new Service
func NewService(logstash logstashv1alpha1.Logstash) *corev1.Service {

	svc := corev1.Service{}
	svc.ObjectMeta.Namespace = logstash.Namespace
	svc.ObjectMeta.Name = HTTPService(logstash.Name)

	labels := labels(logstash)

	// TODO: Make this a setting, rather than hard coding each ingress requirement
	ports := []corev1.ServicePort{
		{
			Name:     "http",
			Protocol: corev1.ProtocolTCP,
			Port:     9600,
		},
		{
			Name:     "beats",
			Protocol: corev1.ProtocolTCP,
			Port:     5044,
		},
	}

	return defaults.SetServiceDefaults(&svc, labels, labels, ports)
}

func buildConfigHash(c k8s.Client, logstash logstashv1alpha1.Logstash, configSecret corev1.Secret) (string, error) {
	// build a hash of various settings to rotate the Pod on any change
	configHash := fnv.New32a()

	// Write logstash yaml
	_, _ = configHash.Write(configSecret.Data[ConfigFilename])
	//// - in the associated Elasticsearch TLS certificates
	if err := commonassociation.WriteAssocsToConfigHash(c, logstash.GetAssociations(), configHash); err != nil {
		return "", err
	}

	return fmt.Sprint(configHash.Sum32()), nil
}

func (r *ReconcileLogstash) reconcileDeployment(
	ctx context.Context,
	logstash logstashv1alpha1.Logstash,
	configHash string,
) (appsv1.Deployment, error) {
	span, _ := apm.StartSpan(ctx, "reconcile_deployment", tracing.SpanTypeApp)
	defer span.End()

	deployParams, err := r.deploymentParams(logstash, configHash)

	if err != nil {
		return appsv1.Deployment{}, err
	}
	deploy := deployment.New(deployParams)

	return deployment.Reconcile(ctx, r.K8sClient(), deploy, &logstash)
}

func (r *ReconcileLogstash) deploymentParams(logstash logstashv1alpha1.Logstash, configHash string) (deployment.Params, error) {
	podSpec, err := newPodSpec(logstash, configHash)
	if err != nil {
		return deployment.Params{}, err
	}

	deploymentLabels := labels(logstash)

	podLabels := maps.Merge(labels(logstash), versionLabels(logstash))
	//// merge with user-provided labels
	podSpec.Labels = maps.MergePreservingExistingKeys(podSpec.Labels, podLabels)

	return deployment.Params{
		Name:            Deployment(logstash.Name),
		Namespace:       logstash.Namespace,
		Replicas:        logstash.Spec.Count,
		Selector:        deploymentLabels,
		Labels:          deploymentLabels,
		PodTemplateSpec: podSpec,
		Strategy:        appsv1.DeploymentStrategy{Type: appsv1.RollingUpdateDeploymentStrategyType},
	}, nil
}

func (r *ReconcileLogstash) getStatus(ctx context.Context, logstash logstashv1alpha1.Logstash, deploy appsv1.Deployment) (logstashv1alpha1.LogstashStatus, error) {
	status := newStatus(logstash)
	pods, err := k8s.PodsMatchingLabels(r.K8sClient(), logstash.Namespace, map[string]string{NameLabelName: logstash.Name})
	if err != nil {
		return status, err
	}
	deploymentStatus, err := common.DeploymentStatus(ctx, logstash.Status.DeploymentStatus, deploy, pods, versionLabelName)
	if err != nil {
		return status, err
	}
	status.DeploymentStatus = deploymentStatus
	status.AssociationStatus = logstash.Status.AssociationStatus

	return status, nil
}

//TODO: Look at MapsStatus

func (r *ReconcileLogstash) updateStatus(ctx context.Context, logstash logstashv1alpha1.Logstash, status logstashv1alpha1.LogstashStatus) error {
	//if reflect.DeepEqual(status, logstash.Status) {
	//	return nil // nothing to do
	//}
	//if status.IsDegraded(logstash.Status.DeploymentStatus) {
	//	r.recorder.Event(&logstash, corev1.EventTypeWarning, events.EventReasonUnhealthy, "Logstash health degraded")
	//}
	//ulog.FromContext(ctx).V(1).Info("Updating status",
	//	"iteration", atomic.LoadUint64(&r.iteration),
	//	"namespace", logstash.Namespace,
	//	"logstash_name", logstash.Name,
	//	"status", status,
	//)
	logstash.Status = status
	return common.UpdateStatus(ctx, r.Client, &logstash)
}

func (r *ReconcileLogstash) onDelete(ctx context.Context, obj types.NamespacedName) error {
	// Clean up watches set on custom http tls certificates
	// TODO:
	//r.dynamicWatches.Secrets.RemoveHandlerForKey(certificates.CertificateWatchKey(EMSNamer, obj.Name))

	// same for the configRef secret
	r.dynamicWatches.Secrets.RemoveHandlerForKey(common.ConfigRefWatchName(obj))
	return reconciler.GarbageCollectSoftOwnedSecrets(ctx, r.Client, obj, logstashv1alpha1.Kind)
}

//// Reconcile reads that state of the cluster for a Logstash object and makes changes based on the state read
//// and what is in the Logstash.Spec
//// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
//// a Deployment as an example
//// Automatically generate RBAC rules to allow the Controller to read and write Deployments
//// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
//// +kubebuilder:rbac:groups=logstash.k8s.elastic.co,resources=logstashes,verbs=get;list;watch;create;update;patch;delete
//// +kubebuilder:rbac:groups=logstash.k8s.elastic.co,resources=logstashes/status,verbs=get;update;patch
//func (r *ReconcileLogstash) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
//	// Fetch the Logstash instance
//	instance := &logstashv1alpha1.Logstash{}
//	err := r.Get(context.TODO(), request.NamespacedName, instance)
//	if err != nil {
//		if errors.IsNotFound(err) {
//			// Object not found, return.  Created objects are automatically garbage collected.
//			// For additional cleanup logic use finalizers.
//			return reconcile.Result{}, nil
//		}
//		// Error reading the object - requeue the request.
//		return reconcile.Result{}, err
//	}
//
//	// TODO(user): Change this to be the object type created by your controller
//	// Define the desired Deployment object
//	deploy := &appsv1.Deployment{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      instance.Name + "-deployment",
//			Namespace: instance.Namespace,
//		},
//		Spec: appsv1.DeploymentSpec{
//			Selector: &metav1.LabelSelector{
//				MatchLabels: map[string]string{"deployment": instance.Name + "-deployment"},
//			},
//			Template: corev1.PodTemplateSpec{
//				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"deployment": instance.Name + "-deployment"}},
//				Spec: corev1.PodSpec{
//					Containers: []corev1.Container{
//						{
//							Name:  "nginx",
//							Image: "nginx",
//						},
//					},
//				},
//			},
//		},
//	}
//	if err := controllerutil.SetControllerReference(instance, deploy, r.scheme); err != nil {
//		return reconcile.Result{}, err
//	}
//
//	// TODO(user): Change this for the object type created by your controller
//	// Check if the Deployment already exists
//	found := &appsv1.Deployment{}
//	err = r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
//	if err != nil && errors.IsNotFound(err) {
//		//log.Info("Creating Deployment", "namespace", deploy.Namespace, "name", deploy.Name)
//		err = r.Create(context.TODO(), deploy)
//		return reconcile.Result{}, err
//	} else if err != nil {
//		return reconcile.Result{}, err
//	}
//
//	// TODO(user): Change this for the object type created by your controller
//	// Update the found object and write the result back if there are any changes
//	if !reflect.DeepEqual(deploy.Spec, found.Spec) {
//		found.Spec = deploy.Spec
//		//log.Info("Updating Deployment", "namespace", deploy.Namespace, "name", deploy.Name)
//		err = r.Update(context.TODO(), found)
//		if err != nil {
//			return reconcile.Result{}, err
//		}
//	}
//	return reconcile.Result{}, nil
//}
