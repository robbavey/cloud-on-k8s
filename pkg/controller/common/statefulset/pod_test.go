// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package statefulset

//import (
//	"testing"
//
//	//"github.com/stretchr/testify/assert"
//	//"github.com/stretchr/testify/require"
//	//appsv1 "k8s.io/api/apps/v1"
//	//corev1 "k8s.io/api/core/v1"
//	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	//"k8s.io/utils/ptr"
//	//"sigs.k8s.io/controller-runtime/pkg/client"
//
//	//esv1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/elasticsearch/v1"
//	//"github.com/elastic/cloud-on-k8s/v2/pkg/utils/k8s"
//)

//func TestPodName(t *testing.T) {
//	require.Equal(t, "sset-2", PodName("sset", 2))
//}
//
//func TestPodNames(t *testing.T) {
//	require.Equal(t,
//		[]string{"sset-0", "sset-1", "sset-2"},
//		PodNames(appsv1.StatefulSet{
//			ObjectMeta: metav1.ObjectMeta{
//				Namespace: "ns",
//				Name:      "sset",
//			},
//			Spec: appsv1.StatefulSetSpec{
//				Replicas: ptr.To[int32](3),
//			},
//		}),
//	)
//}
