// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package logstash

import (
	//"fmt"

	//"github.com/google/uuid"
	//"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/client-go/kubernetes/fake"
	//"k8s.io/client-go/testing"
	"k8s.io/apimachinery/pkg/util/uuid"
	"context"
	//"reflect"
	"testing"
	//k8stesting "k8s.io/client-go/testing"
	//"time"
	//ulog "github.com/elastic/cloud-on-k8s/v2/pkg/utils/log"
	//"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/reconciler"

	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	//"k8s.io/utils/pointer"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	//commonv1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/common/v1"
	logstashv1alpha1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/logstash/v1alpha1"
	//"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common"
	//"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/comparison"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/expectations"
	//"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/hash"
	"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/watches"
	//"github.com/elastic/cloud-on-k8s/v2/pkg/controller/logstash/labels"
	"github.com/elastic/cloud-on-k8s/v2/pkg/utils/k8s"
)

var (
	truePtr = true
	fixedStorageClass = storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{Name: "fixed"},
	}
	resizableStorageClass = storagev1.StorageClass{
		ObjectMeta:           metav1.ObjectMeta{Name: "resizable"},
		AllowVolumeExpansion: &truePtr,
	}
)
func newReconcileLogstash(objs ...client.Object) *ReconcileLogstash {
	client := k8s.NewFakeClient(objs...)
	//client := fake.NewSimpleClientset(objs...)
	//
	//client.PrependReactor("create", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
	//	// Generate a UUID
	//	uid := uuid.NewUUID()
	//
	//	// Get the object being created from the action
	//	obj := action.(k8stesting.CreateAction).GetObject()
	//
	//	// Set the UUID as an annotation or a label based on your needs
	//	// For example, obj.SetAnnotations(map[string]string{"uuid": uid})
	//
	//	fmt.Printf("Generated UUID: %s\n", uid)
	//	return false, obj, nil
	//})

	r := &ReconcileLogstash{
		Client:         client,
		recorder:       record.NewFakeRecorder(100),
		dynamicWatches: watches.NewDynamicWatches(),
		expectations:   expectations.NewExpectations(client),
	}
	return r
}


//func TestReconcileLogstash_Reconcile(t *testing.T) {
//	defaultLabels := (&logstashv1alpha1.Logstash{ObjectMeta: metav1.ObjectMeta{Name: "testLogstash"}}).GetIdentityLabels()
//	tests := []struct {
//		name            string
//		objs            []client.Object
//		request         reconcile.Request
//		want            reconcile.Result
//		expected        logstashv1alpha1.Logstash
//		expectedObjects expectedObjects
//		wantErr         bool
//	}{
//		{
//			name: "valid unmanaged Logstash does not increment observedGeneration",
//			objs: []client.Object{
//				&logstashv1alpha1.Logstash{
//					ObjectMeta: metav1.ObjectMeta{
//						Name:       "testLogstash",
//						Namespace:  "test",
//						Generation: 1,
//						Annotations: map[string]string{
//							common.ManagedAnnotation: "false",
//						},
//					},
//					Spec: logstashv1alpha1.LogstashSpec{
//						Version: "8.6.1",
//					},
//					Status: logstashv1alpha1.LogstashStatus{
//						ObservedGeneration: 1,
//					},
//				},
//			},
//			request: reconcile.Request{
//				NamespacedName: types.NamespacedName{
//					Namespace: "test",
//					Name:      "testLogstash",
//				},
//			},
//			want: reconcile.Result{},
//			expected: logstashv1alpha1.Logstash{
//				ObjectMeta: metav1.ObjectMeta{
//					Name:       "testLogstash",
//					Namespace:  "test",
//					Generation: 1,
//					Annotations: map[string]string{
//						common.ManagedAnnotation: "false",
//					},
//				},
//				Spec: logstashv1alpha1.LogstashSpec{
//					Version: "8.6.1",
//				},
//				Status: logstashv1alpha1.LogstashStatus{
//					ObservedGeneration: 1,
//				},
//			},
//			expectedObjects: []expectedObject{},
//			wantErr:         false,
//		},
//		{
//			name: "too long name fails validation, and updates observedGeneration",
//			objs: []client.Object{
//				&logstashv1alpha1.Logstash{
//					ObjectMeta: metav1.ObjectMeta{
//						Name:       "testLogstashwithtoolongofanamereallylongname",
//						Namespace:  "test",
//						Generation: 2,
//					},
//					Status: logstashv1alpha1.LogstashStatus{
//						ObservedGeneration: 1,
//					},
//				},
//			},
//			request: reconcile.Request{
//				NamespacedName: types.NamespacedName{
//					Namespace: "test",
//					Name:      "testLogstashwithtoolongofanamereallylongname",
//				},
//			},
//			want: reconcile.Result{},
//			expected: logstashv1alpha1.Logstash{
//				ObjectMeta: metav1.ObjectMeta{
//					Name:       "testLogstashwithtoolongofanamereallylongname",
//					Namespace:  "test",
//					Generation: 2,
//				},
//				Status: logstashv1alpha1.LogstashStatus{
//					ObservedGeneration: 2,
//				},
//			},
//			expectedObjects: []expectedObject{},
//			wantErr:         true,
//		},
//		{
//			name: "Logstash with ready StatefulSet and Pod updates status and creates secrets and service",
//			objs: []client.Object{
//				&logstashv1alpha1.Logstash{
//					ObjectMeta: metav1.ObjectMeta{
//						Name:       "testLogstash",
//						Namespace:  "test",
//						Generation: 2,
//					},
//					Spec: logstashv1alpha1.LogstashSpec{
//						Version: "8.6.1",
//						Count:   1,
//					},
//					Status: logstashv1alpha1.LogstashStatus{
//						ObservedGeneration: 1,
//					},
//				},
//				&appsv1.StatefulSet{
//					ObjectMeta: metav1.ObjectMeta{
//						Name:      "testLogstash-ls",
//						Namespace: "test",
//						Labels:    addLabel(defaultLabels, hash.TemplateHashLabelName, "3145706383"),
//					},
//					Status: appsv1.StatefulSetStatus{
//						AvailableReplicas: 1,
//						Replicas:          1,
//						ReadyReplicas:     1,
//					},
//					Spec: appsv1.StatefulSetSpec{
//						VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
//							{
//								ObjectMeta: metav1.ObjectMeta{
//									Name:      "logstash-data",
//									Namespace: "test",
//								},
//								Spec: corev1.PersistentVolumeClaimSpec{
//									StorageClassName: pointer.String(sampleStorageClass.Name),
//									Resources: corev1.ResourceRequirements{
//										Requests: corev1.ResourceList{
//											corev1.ResourceStorage: resource.MustParse("1Gi"),
//										},
//									},
//								},
//							},
//						},
//					},
//				},
//				&corev1.Pod{
//					ObjectMeta: metav1.ObjectMeta{
//						Name:       "testLogstash-ls",
//						Namespace:  "test",
//						Generation: 2,
//						Labels:     map[string]string{labels.NameLabelName: "testLogstash", VersionLabelName: "8.6.1"},
//					},
//					Status: corev1.PodStatus{
//						Phase: corev1.PodRunning,
//					},
//				},
//			},
//			request: reconcile.Request{
//				NamespacedName: types.NamespacedName{
//					Namespace: "test",
//					Name:      "testLogstash",
//				},
//			},
//			want: reconcile.Result{},
//			expectedObjects: []expectedObject{
//				{
//					t:    &corev1.Service{},
//					name: types.NamespacedName{Namespace: "test", Name: "testLogstash-ls-api"},
//				},
//				{
//					t:    &corev1.Secret{},
//					name: types.NamespacedName{Namespace: "test", Name: "testLogstash-ls-config"},
//				},
//				{
//					t:    &corev1.Secret{},
//					name: types.NamespacedName{Namespace: "test", Name: "testLogstash-ls-pipeline"},
//				},
//			},
//
//			expected: logstashv1alpha1.Logstash{
//				ObjectMeta: metav1.ObjectMeta{
//					Name:       "testLogstash",
//					Namespace:  "test",
//					Generation: 2,
//				},
//				Spec: logstashv1alpha1.LogstashSpec{
//					Version: "8.6.1",
//					Count:   1,
//				},
//				Status: logstashv1alpha1.LogstashStatus{
//					Version:            "8.6.1",
//					ExpectedNodes:      1,
//					AvailableNodes:     1,
//					ObservedGeneration: 2,
//					Selector:           "common.k8s.elastic.co/type=logstash,logstash.k8s.elastic.co/name=testLogstash",
//				},
//			},
//			wantErr: false,
//		},
//		{
//			name: "Logstash with a custom service creates secrets and service",
//			objs: []client.Object{
//				&logstashv1alpha1.Logstash{
//					ObjectMeta: metav1.ObjectMeta{
//						Name:       "testLogstash",
//						Namespace:  "test",
//						Generation: 2,
//					},
//					Spec: logstashv1alpha1.LogstashSpec{
//						Version: "8.6.1",
//						Count:   1,
//						Services: []logstashv1alpha1.LogstashService{{
//							Name: "test",
//							Service: commonv1.ServiceTemplate{
//								Spec: corev1.ServiceSpec{
//									Ports: []corev1.ServicePort{
//										{Protocol: "TCP", Port: 9500},
//									},
//								},
//							},
//						}},
//					},
//					Status: logstashv1alpha1.LogstashStatus{
//						ObservedGeneration: 1,
//					},
//				},
//				&appsv1.StatefulSet{
//					ObjectMeta: metav1.ObjectMeta{
//						Name:      "testLogstash-ls",
//						Namespace: "test",
//						Labels:    addLabel(defaultLabels, hash.TemplateHashLabelName, "3145706383"),
//					},
//					Status: appsv1.StatefulSetStatus{
//						AvailableReplicas: 1,
//						Replicas:          1,
//						ReadyReplicas:     1,
//					},
//					Spec: appsv1.StatefulSetSpec{
//						VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
//							{
//								ObjectMeta: metav1.ObjectMeta{
//									Name:      "logstash-data",
//									Namespace: "test",
//								},
//								Spec: corev1.PersistentVolumeClaimSpec{
//									StorageClassName: pointer.String(sampleStorageClass.Name),
//									Resources: corev1.ResourceRequirements{
//										Requests: corev1.ResourceList{
//											corev1.ResourceStorage: resource.MustParse("1Gi"),
//										},
//									},
//								},
//							},
//						},
//					},
//				},
//				&corev1.Pod{
//					ObjectMeta: metav1.ObjectMeta{
//						Name:       "testLogstash-ls",
//						Namespace:  "test",
//						Generation: 2,
//						Labels:     map[string]string{labels.NameLabelName: "testLogstash", VersionLabelName: "8.6.1"},
//					},
//					Status: corev1.PodStatus{
//						Phase: corev1.PodRunning,
//					},
//				},
//			},
//			request: reconcile.Request{
//				NamespacedName: types.NamespacedName{
//					Namespace: "test",
//					Name:      "testLogstash",
//				},
//			},
//			want: reconcile.Result{},
//			expectedObjects: []expectedObject{
//				{
//					t:    &corev1.Service{},
//					name: types.NamespacedName{Namespace: "test", Name: "testLogstash-ls-api"},
//				},
//				{
//					t:    &corev1.Service{},
//					name: types.NamespacedName{Namespace: "test", Name: "testLogstash-ls-test"},
//				},
//				{
//					t:    &corev1.Secret{},
//					name: types.NamespacedName{Namespace: "test", Name: "testLogstash-ls-config"},
//				},
//				{
//					t:    &corev1.Secret{},
//					name: types.NamespacedName{Namespace: "test", Name: "testLogstash-ls-pipeline"},
//				},
//			},
//
//			expected: logstashv1alpha1.Logstash{
//				ObjectMeta: metav1.ObjectMeta{
//					Name:       "testLogstash",
//					Namespace:  "test",
//					Generation: 2,
//				},
//				Spec: logstashv1alpha1.LogstashSpec{
//					Version: "8.6.1",
//					Count:   1,
//					Services: []logstashv1alpha1.LogstashService{{
//						Name: "test",
//						Service: commonv1.ServiceTemplate{
//							Spec: corev1.ServiceSpec{
//								Ports: []corev1.ServicePort{
//									{Protocol: "TCP", Port: 9500},
//								},
//							},
//						},
//					}},
//				},
//				Status: logstashv1alpha1.LogstashStatus{
//					Version:            "8.6.1",
//					ExpectedNodes:      1,
//					AvailableNodes:     1,
//					ObservedGeneration: 2,
//					Selector:           "common.k8s.elastic.co/type=logstash,logstash.k8s.elastic.co/name=testLogstash",
//				},
//			},
//			wantErr: false,
//		},
//		{
//			name: "Logstash with a service with no port creates secrets and service",
//			objs: []client.Object{
//				&logstashv1alpha1.Logstash{
//					ObjectMeta: metav1.ObjectMeta{
//						Name:       "testLogstash",
//						Namespace:  "test",
//						Generation: 2,
//					},
//					Spec: logstashv1alpha1.LogstashSpec{
//						Version: "8.6.1",
//						Count:   1,
//						Services: []logstashv1alpha1.LogstashService{{
//							Name: "api",
//							Service: commonv1.ServiceTemplate{
//								Spec: corev1.ServiceSpec{
//									Ports: nil,
//								},
//							},
//						}},
//					},
//					Status: logstashv1alpha1.LogstashStatus{
//						ObservedGeneration: 1,
//					},
//				},
//				&storagev1.StorageClass{
//					ObjectMeta: metav1.ObjectMeta{
//						Name: "default-sc",
//					},
//				},
//				&appsv1.StatefulSet{
//					ObjectMeta: metav1.ObjectMeta{
//						Name:      "testLogstash-ls",
//						Namespace: "test",
//						Labels:    addLabel(defaultLabels, hash.TemplateHashLabelName, "3145706383"),
//					},
//					Status: appsv1.StatefulSetStatus{
//						AvailableReplicas: 1,
//						Replicas:          1,
//						ReadyReplicas:     1,
//					},
//					Spec: appsv1.StatefulSetSpec{
//						VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
//							{
//								ObjectMeta: metav1.ObjectMeta{
//									Name:      "logstash-data",
//									Namespace: "test",
//								},
//								Spec: corev1.PersistentVolumeClaimSpec{
//									StorageClassName: pointer.String(sampleStorageClass.Name),
//									Resources: corev1.ResourceRequirements{
//										Requests: corev1.ResourceList{
//											corev1.ResourceStorage: resource.MustParse("1Gi"),
//										},
//									},
//								},
//							},
//						},
//					},
//				},
//				&corev1.Pod{
//					ObjectMeta: metav1.ObjectMeta{
//						Name:       "testLogstash-ls",
//						Namespace:  "test",
//						Generation: 2,
//						Labels:     map[string]string{labels.NameLabelName: "testLogstash", VersionLabelName: "8.6.1"},
//					},
//					Status: corev1.PodStatus{
//						Phase: corev1.PodRunning,
//					},
//				},
//			},
//			request: reconcile.Request{
//				NamespacedName: types.NamespacedName{
//					Namespace: "test",
//					Name:      "testLogstash",
//				},
//			},
//			want: reconcile.Result{},
//			expectedObjects: []expectedObject{
//				{
//					t:    &corev1.Service{},
//					name: types.NamespacedName{Namespace: "test", Name: "testLogstash-ls-api"},
//				},
//				{
//					t:    &corev1.Secret{},
//					name: types.NamespacedName{Namespace: "test", Name: "testLogstash-ls-config"},
//				},
//				{
//					t:    &corev1.Secret{},
//					name: types.NamespacedName{Namespace: "test", Name: "testLogstash-ls-pipeline"},
//				},
//			},
//
//			expected: logstashv1alpha1.Logstash{
//				ObjectMeta: metav1.ObjectMeta{
//					Name:       "testLogstash",
//					Namespace:  "test",
//					Generation: 2,
//				},
//				Spec: logstashv1alpha1.LogstashSpec{
//					Version: "8.6.1",
//					Count:   1,
//					Services: []logstashv1alpha1.LogstashService{{
//						Name: "api",
//						Service: commonv1.ServiceTemplate{
//							Spec: corev1.ServiceSpec{
//								Ports: nil,
//							},
//						},
//					}},
//				},
//				Status: logstashv1alpha1.LogstashStatus{
//					Version:            "8.6.1",
//					ExpectedNodes:      1,
//					AvailableNodes:     1,
//					ObservedGeneration: 2,
//					Selector:           "common.k8s.elastic.co/type=logstash,logstash.k8s.elastic.co/name=testLogstash",
//				},
//			},
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			r := newReconcileLogstash(tt.objs...)
//			got, err := r.Reconcile(context.Background(), tt.request)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("ReconcileLogstash.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("ReconcileLogstash.Reconcile() = %v, want %v", got, tt.want)
//			}
//
//			var Logstash logstashv1alpha1.Logstash
//			if err := r.Client.Get(context.Background(), tt.request.NamespacedName, &Logstash); err != nil {
//				t.Error(err)
//				return
//			}
//			tt.expectedObjects.assertExist(t, r.Client)
//			comparison.AssertEqual(t, &Logstash, &tt.expected)
//		})
//	}
//}

func Test_PVCResizeAllowed(t *testing.T) {
	ctx := context.Background()
	request := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "test",Name: "testLogstash"}}
	r := setupFixtures("1Gi", resizableStorageClass)
	result, err := r.Reconcile(ctx, request)
	require.NoError(t, err)
	require.Equal(t, reconcile.Result{}, result)


	updatedLs := logstashv1alpha1.Logstash{}
	r.Client.Get(ctx, request.NamespacedName, &updatedLs)


	// Update sizing requirements
	logstashResized := *updatedLs.DeepCopy()
	logstashResized.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse("3Gi")
	r.Client.Update(ctx, &logstashResized)

	// Run reconciler: First pass of the reconciler will add annotations to Logstash including details of the stateful
	//                 set to be replaced.
	result, err = r.Reconcile(ctx, request)
	require.Equal(t, reconcile.Result{Requeue: true}, result)

	updatedls := logstashv1alpha1.Logstash{}
	r.Client.Get(ctx, request.NamespacedName, &updatedls)
	require.Equal(t, 1, len(updatedls.Annotations))

	// Run reconciler: Second pass of the reconciler removes the stateful set, and should requeue the request
	result, err = r.Reconcile(ctx, request)
	require.NoError(t, err)
	require.NotZero(t, result.RequeueAfter)

	ssNamespacedName := types.NamespacedName{Namespace: "test", Name: "testLogstash-ls"}
	ss := appsv1.StatefulSet{}
	require.Error(t, r.Client.Get(ctx, ssNamespacedName, &ss))

	// Run reconciler: third pass of the reconciler should recreate the stateful set
	result, err = r.Reconcile(ctx, request)
	require.NoError(t, err)
	require.NotZero(t, result.RequeueAfter)

	ss = appsv1.StatefulSet{}
	require.NoError(t, r.Client.Get(ctx, ssNamespacedName, &ss))
	require.Equal(t, resource.MustParse("3Gi"), ss.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage])

	ss.UID = uuid.NewUUID()
	r.Client.Update(ctx, &ss)

	// Run reconciler: Final pass of the reconciler should delete the logstash annotations
	result, err = r.Reconcile(ctx, request)
	require.NoError(t, err)

	require.Equal(t, reconcile.Result{Requeue: false}, result)
	updatedls = logstashv1alpha1.Logstash{}
	r.Client.Get(ctx, request.NamespacedName, &updatedls)
	require.Equal(t, 0, len(updatedls.Annotations))
	require.NoError(t, r.Client.Get(ctx, ssNamespacedName, &ss))
	require.Equal(t, resource.MustParse("3Gi"), ss.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage])
}

func Test_PVCResizeNotAllowedStorageClass(t *testing.T) {
	ctx := context.Background()
	request := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "test",Name: "testLogstash"}}
	r := setupFixtures("5Gi", fixedStorageClass)
	result, err := r.Reconcile(ctx, request)
	require.NoError(t, err)
	require.Equal(t, reconcile.Result{}, result)

	updatedLs := logstashv1alpha1.Logstash{}
	r.Client.Get(ctx, request.NamespacedName, &updatedLs)

	// Update sizing requirements
	logstashResized := *updatedLs.DeepCopy()
	logstashResized.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse("3Gi")
	r.Client.Update(ctx, &logstashResized)

	// ---- end prepare section
	_, err = r.Reconcile(ctx, request)
	require.Error(t, err)
}

func Test_PVCResizeNotAllowedShrinking(t *testing.T) {
	ctx := context.Background()
	request := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "test",Name: "testLogstash"}}
	r := setupFixtures("5Gi", resizableStorageClass)
	result, err := r.Reconcile(ctx, request)
	require.NoError(t, err)
	require.Equal(t, reconcile.Result{}, result)


	updatedLs := logstashv1alpha1.Logstash{}
	r.Client.Get(ctx, request.NamespacedName, &updatedLs)

	// Update sizing requirements
	logstashResized := *updatedLs.DeepCopy()
	logstashResized.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse("3Gi")
	r.Client.Update(ctx, &logstashResized)

	_, err = r.Reconcile(ctx, request)
	require.Error(t, err)
}


func setupFixtures(capacity string, storage storagev1.StorageClass) (*ReconcileLogstash) {
	ls := createLogstash(capacity, storage.Name)
	pod := createPod()
	return newReconcileLogstash(&ls, &pod, &storage)
}

func createPod() corev1.Pod {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "testLogstash-ls-0",
			Namespace:  "test",
			UID:        uuid.NewUUID(),
			Generation: 1,
			Labels: map[string]string{
				"common.k8s.elastic.co/type":               "logstash",
				"logstash.k8s.elastic.co/name":             "testLogstash",
				"logstash.k8s.elastic.co/statefulset-name": "testLogstash-ls",
				"logstash.k8s.elastic.co/version":          "8.11.0",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
	return pod
}

func createLogstash(capacity string, storageClassName string) logstashv1alpha1.Logstash {
	ls := logstashv1alpha1.Logstash{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "testLogstash",
			UID:       uuid.NewUUID(),
		},
		Spec: logstashv1alpha1.LogstashSpec{
			Version: "8.11.0",
			Count:   1,
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{ObjectMeta: metav1.ObjectMeta{
					Name: "logstash-data",
				},
					Spec: corev1.PersistentVolumeClaimSpec{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse(capacity),
							},
						},
						StorageClassName: &storageClassName,
					},
				},
			},
		},
	}
	return ls
}


func addLabel(labels map[string]string, key, value string) map[string]string {
	newLabels := make(map[string]string, len(labels))
	for k, v := range labels {
		newLabels[k] = v
	}
	newLabels[key] = value
	return newLabels
}

type expectedObject struct {
	t    client.Object
	name types.NamespacedName
}

type expectedObjects []expectedObject

func (e expectedObjects) assertExist(t *testing.T, k8s client.Client) {
	t.Helper()
	for _, o := range e {
		obj := o.t.DeepCopyObject().(client.Object) //nolint:forcetypeassert
		assert.NoError(t, k8s.Get(context.Background(), o.name, obj), "Expected object not found: %s", o.name)
	}
}
