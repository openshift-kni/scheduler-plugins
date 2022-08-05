/*
 * Copyright 2022 Red Hat, Inc.
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

package knidebug

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/dustin/go-humanize"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	resourcehelper "k8s.io/kubernetes/pkg/api/v1/resource"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

type KNIDebug struct{}

var _ framework.FilterPlugin = &KNIDebug{}

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name     string     = "KNIDebug"
	LogLevel klog.Level = 6
)

// Name returns name of the plugin. It is used in logs, etc.
func (kd *KNIDebug) Name() string {
	return Name
}

// New initializes a new plugin and returns it.
func New(args runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	klog.V(6).InfoS("Creating new KNIDebug plugin")
	return &KNIDebug{}, nil
}

func (kd *KNIDebug) EventsToRegister() []framework.ClusterEvent {
	// this can actually be empty - this plugin never fails, but we keep the same
	// (simple and safe) events noderesourcesfit registered
	return []framework.ClusterEvent{
		{Resource: framework.Pod, ActionType: framework.Delete},
		{Resource: framework.Node, ActionType: framework.Add | framework.UpdateNodeAllocatable},
	}
}

func (kd *KNIDebug) Filter(ctx context.Context, cycleState *framework.CycleState, pod *corev1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	node := nodeInfo.Node()
	if node == nil {
		// should never happen
		return framework.NewStatus(framework.Error, "node not found")
	}

	logKey := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
	// note the fit.go plugin computes this in the prefilter stage. Does this make any practical difference in our context?
	req := computePodResourceRequest(pod)
	checkRequest(LogLevel, logKey, req, nodeInfo)
	return nil // must never fail
}

func frameworkResourceToLoggable(logKey string, req *framework.Resource) []interface{} {
	items := []interface{}{
		"logKey", logKey,
		"cpu", humanCPU(req.MilliCPU),
		"memory", humanMemory(req.Memory),
	}

	resNames := []string{}
	for resName := range req.ScalarResources {
		resNames = append(resNames, string(resName))
	}
	sort.Strings(resNames)

	for _, resName := range resNames {
		quan := req.ScalarResources[corev1.ResourceName(resName)]
		if resourcehelper.IsHugePageResourceName(corev1.ResourceName(resName)) {
			items = append(items, resName, humanMemory(quan))
		} else {
			items = append(items, resName, strconv.FormatInt(quan, 10))
		}
	}
	return items
}

type humanMemory int64

func (hi humanMemory) String() string {
	return fmt.Sprintf("%d (%s)", hi, humanize.IBytes(uint64(hi)))
}

type humanCPU int64

func (hc humanCPU) String() string {
	return fmt.Sprintf("%d (%d)", hc, hc/1000)
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// logic taken from fit.go, changing the return value to use a plain *framework.Resource
// https://github.com/kubernetes/kubernetes/blob/v1.24.0/pkg/scheduler/framework/plugins/noderesources/fit.go#L133-L175

// computePodResourceRequest returns a framework.Resource that covers the largest
// width in each resource dimension. Because init-containers run sequentially, we collect
// the max in each dimension iteratively. In contrast, we sum the resource vectors for
// regular containers since they run simultaneously.
//
// The resources defined for Overhead should be added to the calculated Resource request sum
//
// Example:
//
// Pod:
//   InitContainers
//     IC1:
//       CPU: 2
//       Memory: 1G
//     IC2:
//       CPU: 2
//       Memory: 3G
//   Containers
//     C1:
//       CPU: 2
//       Memory: 1G
//     C2:
//       CPU: 1
//       Memory: 1G
//
// Result: CPU: 3, Memory: 3G
func computePodResourceRequest(pod *corev1.Pod) *framework.Resource {
	result := &framework.Resource{}
	for _, container := range pod.Spec.Containers {
		result.Add(container.Resources.Requests)
	}

	// take max_resource(sum_pod, any_init_container)
	for _, container := range pod.Spec.InitContainers {
		result.SetMaxResource(container.Resources.Requests)
	}

	if pod.Spec.Overhead != nil {
		result.Add(pod.Spec.Overhead)
	}
	return result
}

// see again fit.go for the skeleton code. Here we intentionally only log
func checkRequest(logLevel klog.Level, logKey string, podRequest *framework.Resource, nodeInfo *framework.NodeInfo) {
	if podRequest.MilliCPU == 0 && podRequest.Memory == 0 && podRequest.EphemeralStorage == 0 && len(podRequest.ScalarResources) == 0 {
		klog.V(logLevel).InfoS("target resource requests none", "logKey", logKey)
		return
	}
	klog.V(logLevel).InfoS("target resource requests", frameworkResourceToLoggable(logKey, podRequest)...)

	nodeName := nodeInfo.Node().Name // shortcut

	violations := 0
	if availCPU := (nodeInfo.Allocatable.MilliCPU - nodeInfo.Requested.MilliCPU); podRequest.MilliCPU > availCPU {
		klog.V(logLevel).InfoS("insufficient node resources", "logKey", logKey, "node", nodeName, "resource", "CPU", "request", humanCPU(podRequest.MilliCPU), "available", humanCPU(availCPU))
		violations++
	}
	if availMemory := (nodeInfo.Allocatable.Memory - nodeInfo.Requested.Memory); podRequest.Memory > availMemory {
		klog.V(logLevel).InfoS("insufficient node resources", "logKey", logKey, "node", nodeName, "resource", "memory", "request", humanMemory(podRequest.Memory), "available", humanMemory(availMemory))
		violations++
	}
	for rName, rQuant := range podRequest.ScalarResources {
		if availQuant := (nodeInfo.Allocatable.ScalarResources[rName] - nodeInfo.Requested.ScalarResources[rName]); rQuant > availQuant {
			klog.V(logLevel).InfoS("insufficient node resources", "logKey", logKey, "node", nodeName, "resource", rName, "request", rQuant, "available", availQuant)
			violations++
		}
	}

	if violations > 0 {
		return
	}
	klog.V(logLevel).InfoS("enough node resources", "logKey", logKey, "node", nodeName)
}
