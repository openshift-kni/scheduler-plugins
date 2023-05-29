/*
 * Copyright 2023 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package podinformer

import (
	"context"
	"os"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	podlisterv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	k8scache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	nrtInformerEnvVar string = "NRT_ENABLE_INFORMER"
)

var (
	enabled bool
)

func IsEnabled() bool {
	return enabled
}

func Setup() {
	hasNRTInf, ok := os.LookupEnv(nrtInformerEnvVar)
	if !ok || hasNRTInf == "" {
		klog.InfoS("NRT specific informer disabled", "variableFound", ok, "valueGiven", hasNRTInf != "")
		return
	}
	val, err := strconv.ParseBool(hasNRTInf)
	if err != nil {
		klog.Error(err, "NRT specific informer disabled")
		return
	}
	klog.InfoS("NRT specific informer status", "value", val)
	enabled = val
}

func FromHandle(handle framework.Handle) (k8scache.SharedIndexInformer, podlisterv1.PodLister) {
	if !IsEnabled() {
		podHandle := handle.SharedInformerFactory().Core().V1().Pods() // shortcut
		return podHandle.Informer(), podHandle.Lister()
	}

	podInformer := coreinformers.NewFilteredPodInformer(handle.ClientSet(), metav1.NamespaceAll, 0, cache.Indexers{}, nil)
	podLister := podlisterv1.NewPodLister(podInformer.GetIndexer())

	klog.V(5).InfoS("Start custom pod informer")
	ctx := context.Background()
	go podInformer.Run(ctx.Done())

	klog.V(5).InfoS("Syncing custom pod informer")
	cache.WaitForCacheSync(ctx.Done(), podInformer.HasSynced)
	klog.V(5).InfoS("Synced custom pod informer")

	return podInformer, podLister
}

func IsPodRelevantForState(pod *corev1.Pod) bool {
	if pod == nil {
		return false // should never happen
	}
	if IsEnabled() {
		return true // consider all pods including ones in terminal phase
	}
	return pod.Status.Phase == corev1.PodRunning // we are interested only about nodes which consume resources
}
