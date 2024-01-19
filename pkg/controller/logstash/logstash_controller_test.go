// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package logstash

import (
	"context"
	//"reflect"
	"testing"
	"time"
	ulog "github.com/elastic/cloud-on-k8s/v2/pkg/utils/log"
	//"github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/reconciler"

	"k8s.io/apimachinery/pkg/util/uuid"
	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
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

func newReconcileLogstash(objs ...client.Object) *ReconcileLogstash {
	client := k8s.NewFakeClient(objs...)
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


//func TestReconcileLogstash_ReconcileVolumeExpansion(t *testing.T) {
//	defaultLabels := (&logstashv1alpha1.Logstash{ObjectMeta: metav1.ObjectMeta{Name: "testLogstash"}}).GetIdentityLabels()
//	tests := []struct {
//		name            string
//		objs            []client.Object
//		secondObjs            []client.Object
//		request         reconcile.Request
//		want            reconcile.Result
//		expected        logstashv1alpha1.Logstash
//		expectedObjects expectedObjects
//		wantErr         bool
//	}{
//
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
//			secondObjs: []client.Object{
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
//											corev1.ResourceStorage: resource.MustParse("2Gi"),
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
//					ObservedGeneration: 4,
//					Selector:           "common.k8s.elastic.co/type=logstash,logstash.k8s.elastic.co/name=testLogstash",
//				},
//			},
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			r := newReconcileLogstash(tt.objs...)
//			r2 := newReconcileLogstash(tt.objs...)
//
//			got, err := r.Reconcile(context.Background(), tt.request)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("ReconcileLogstash.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			got, err = r2.Reconcile(context.Background(), tt.request)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("ReconcileLogstash.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//
//
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


func Test_PVCResize(t *testing.T) {
	// focus on the special case of handling PVC resize
	//lsuid := uuid.NewUUID()

	truePtr := true
	storageClass := storagev1.StorageClass{
		ObjectMeta:           metav1.ObjectMeta{Name: "resizeable"},
		AllowVolumeExpansion: &truePtr,
	}
	ls := logstashv1alpha1.Logstash{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "testLogstash",
		},
		Spec: logstashv1alpha1.LogstashSpec{
			Version: "8.11.0",
			Count: 1,
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{ObjectMeta: metav1.ObjectMeta{Name: "logstash-data"},
					Spec: corev1.PersistentVolumeClaimSpec{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
						StorageClassName: &storageClass.Name,
					},
				},
			},

		},

	}

	//ls2 := logstashv1alpha1.Logstash{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Namespace: "test",
	//		Name:      "testLogstash",
	//	},
	//	Spec: logstashv1alpha1.LogstashSpec{
	//		Version: "8.11.0",
	//		Count: 1,
	//		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
	//			{ObjectMeta: metav1.ObjectMeta{Name: "logstash-data-testLogstash-ls-0"},
	//				Spec: corev1.PersistentVolumeClaimSpec{
	//					Resources: corev1.ResourceRequirements{
	//						Requests: corev1.ResourceList{
	//							corev1.ResourceStorage: resource.MustParse("3Gi"),
	//						},
	//					},
	//					StorageClassName: &storageClass.Name,
	//				},
	//			},
	//		},
	//
	//	},
	//
	//}


	// 1 nodes x 1Gi storage
	actualStatefulSet := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "testLogstash-ls",
			UID: 		uuid.NewUUID(),
			Labels: map[string]string{
				"common.k8s.elastic.co/type":   "logstash",
				"logstash.k8s.elastic.co/name": "testLogstash",
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: pointer.Int32(1),
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{ObjectMeta: metav1.ObjectMeta{Name: "logstash-data"},
					Spec: corev1.PersistentVolumeClaimSpec{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
						StorageClassName: &storageClass.Name,
					},
				},
			},
		},
	}
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "testLogstash-ls-0",
			Namespace:  "test",
			Generation: 1,
			Labels: map[string]string{
				"common.k8s.elastic.co/type":"logstash",
				"logstash.k8s.elastic.co/name":"testLogstash",
				"logstash.k8s.elastic.co/statefulset-name":"testLogstash-ls",
				"logstash.k8s.elastic.co/version":"8.11.0",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	ctx := context.Background()
	r := newReconcileLogstash(&ls, &pod, &storageClass,  &actualStatefulSet)
	request := reconcile.Request{
						NamespacedName: types.NamespacedName{
							Namespace: "test",
							Name:      "testLogstash",
						},
					}

	result, err := r.Reconcile(ctx, request)
	require.NoError(t, err)
	updatedLs := logstashv1alpha1.Logstash{}
	r.Client.Get(ctx, k8s.ExtractNamespacedName(&ls.ObjectMeta), &updatedLs)
	ulog.FromContext(ctx).V(1).Info("ls aftger reconcile", "ls", ls)
	//ulog.FromContext(ctx).V(1).Info("ls aftger reconcile", "us", updatedLs)


	//params := newParams(r, updatedLs)
	//ulog.FromContext(ctx).V(1).Info("ls aftger params", "params", params)
	//results, _ := funcName(params, dataResized)

	logstashResized := *updatedLs.DeepCopy()
	logstashResized.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse("3Gi")

	r.Client.Update(ctx, &logstashResized)


	//results, err := internalReconcile(params)
	result, err = r.Reconcile(ctx, request)

	//ulog.FromContext(ctx).V(1).Info("ls aftger func name", "ls", ls)
	require.Equal(t, reconcile.Result{Requeue: true}, result)
	//require.Equal(t, reconciler.NewResult(ctx).WithResult(reconcile.Result{Requeue: true}), results)
	updatedls := logstashv1alpha1.Logstash{}
	r.Client.Get(ctx, request.NamespacedName, &updatedls)
	require.Equal(t, 1, len(updatedls.Annotations))
	ulog.FromContext(ctx).V(1).Info("* First Pass - marked with requeue, and Annotations correctly set")

	//params = newParams(r, updatedls)
	//ulog.FromContext(ctx).V(1).Info("ls aftger params", "ls", ls)


	result, err = r.Reconcile(ctx, request)
	require.Equal(t, reconcile.Result{RequeueAfter: 30 * time.Second}, result)
	require.NoError(t, err)

	ulog.FromContext(ctx).V(1).Info("______ First reconcile stateful set should be deleted --------")
	//results, _ = funcName(params, dataResized)
	ss := appsv1.StatefulSet{}
	require.Error(t, r.Client.Get(ctx, k8s.ExtractNamespacedName(&actualStatefulSet.ObjectMeta), &ss))

	//updatedls = logstashv1alpha1.Logstash{}
	//ulog.FromContext(ctx).V(1).Info("Updated LS1", "ls", updatedls)
	//r.Client.Get(ctx, k8s.ExtractNamespacedName(&ls.ObjectMeta), &updatedls)
	//r.Client.Get(ctx, request.NamespacedName, &ls)
	//ulog.FromContext(ctx).V(1).Info("original", "requestName", request.NamespacedName, "extracted", k8s.ExtractNamespacedName(&ls.ObjectMeta), "ls", ls)


	//r = newReconcileLogstash(&updatedls, &pod, &storageClass,  &actualStatefulSet)
	//result, err = r.Reconcile(ctx, request)
	//require.Equal(t, reconcile.Result{RequeueAfter: 30 * time.Second}, result)
	//require.NoError(t, err)

	//params = newParams(r, updatedls)
	//results, _ = funcName(params, dataResized)
	//require.Equal(t, reconciler.NewResult(ctx).WithResult(reconcile.Result{RequeueAfter: 30 * time.Second}), results)
	//ss = appsv1.StatefulSet{}
	//require.NoError(t, r.Client.Get(ctx, k8s.ExtractNamespacedName(&actualStatefulSet.ObjectMeta), &ss))
	//updatedls = logstashv1alpha1.Logstash{}
	//r.Client.Get(ctx, k8s.ExtractNamespacedName(&ls.ObjectMeta), &updatedls)
	//ulog.FromContext(ctx).V(1).Info("Updated LS2", "ls", updatedls)

	//result, err = r.Reconcile(ctx, request)
	//require.Equal(t, reconcile.Result{Requeue: false}, result)
	//require.NoError(t, err)

	result, err = r.Reconcile(ctx, request)
	require.NoError(t, err)

	ulog.FromContext(ctx).V(1).Info("______ Second reconcile stateful set should be recreated --------")

	ss = appsv1.StatefulSet{}
	require.NoError(t, r.Client.Get(ctx, k8s.ExtractNamespacedName(&actualStatefulSet.ObjectMeta), &ss))
	require.NoError(t, r.Client.Get(ctx, k8s.ExtractNamespacedName(&actualStatefulSet.ObjectMeta), &ss))
	require.Equal(t, resource.MustParse("3Gi"), ss.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage])

	require.Equal(t, reconcile.Result{RequeueAfter: 30 * time.Second}, result)

	result, err = r.Reconcile(ctx, request)
	require.NoError(t, err)

	ulog.FromContext(ctx).V(1).Info("______ Third reconcile should be all set --------")
	updatedls = logstashv1alpha1.Logstash{}
	r.Client.Get(ctx, request.NamespacedName, &updatedls)
	require.Equal(t, 0, len(updatedls.Annotations))

	require.Equal(t, reconcile.Result{Requeue: false}, result)


	//r = newReconcileLogstash(&ls, &pod, &storageClass,  &dataResized)
	//result, err = r.Reconcile(ctx, request)
	//require.Equal(t, reconcile.Result{Requeue: false}, result)
	//require.NoError(t, err)

	//params = newParams(r, updatedls)
	//results, _ = funcName(params, ss)
	//require.Equal(t, reconciler.NewResult(ctx).WithResult(reconcile.Result{Requeue: false}), results)

	//
	//require.Equal(t, reconciler.NewResult(ctx).WithResult(reconcile.Result{Requeue: true}), results)
	//ss = appsv1.StatefulSet{}
	//require.NoError(t, r.Client.Get(ctx, k8s.ExtractNamespacedName(&actualStatefulSet.ObjectMeta), &ss))

	//updatedls = logstashv1alpha1.Logstash{}
	//r.Client.Get(ctx, k8s.ExtractNamespacedName(&ls.ObjectMeta), &updatedls)
	//require.Equal(t, "", updatedls.Annotations)


	//results, _ = funcName(params, dataResized)
	//require.Equal(t, reconciler.NewResult(ctx).WithResult(reconcile.Result{Requeue: false}), results)
	//
	//params = newParams(r, ls, &storageClass, &actualStatefulSet)
	//ulog.FromContext(ctx).V(1).Info("ls aftger params 2", "ls", ls)
	//results, _ = funcName(params, resizeStatefulSet)
	//ulog.FromContext(ctx).V(1).Info("ls aftger func name 2", "ls", ls)
	//require.Equal(t, reconciler.NewResult(ctx).WithResult(reconcile.Result{Requeue: true}), results)
	//
	//require.Error(t, err)
	//require.NoError(t, r.Client.Get(ctx, k8s.ExtractNamespacedName(&ls.ObjectMeta), &ls))

	require.Error(t, r.Client.Get(ctx, k8s.ExtractNamespacedName(&actualStatefulSet.ObjectMeta), &ss))
	require.Equal(t, resource.MustParse("3Gi"), ss.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage])

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

func newParams(reconciler *ReconcileLogstash, ls logstashv1alpha1.Logstash) Params {
	//client := k8s.NewFakeClient(objs...)
	r := Params{
		Client:         reconciler.Client,
		Context: 		context.Background(),
		Logstash: ls,
		EventRecorder:       reconciler.recorder,
		Watches: reconciler.dynamicWatches,
		Expectations:   reconciler.expectations,
	}
	return r
}
