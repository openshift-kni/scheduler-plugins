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
	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	resourcehelper "k8s.io/kubernetes/pkg/api/v1/resource"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	knifeatures "sigs.k8s.io/scheduler-plugins/pkg-kni/features"
)

type KNIDebug struct{}

var _ framework.FilterPlugin = &KNIDebug{}

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name     string = "KNIDebug"
	LogLevel int    = 6
)

// Name returns name of the plugin. It is used in logs, etc.
func (kd *KNIDebug) Name() string {
	return Name
}

// New initializes a new plugin and returns it.
func New(_ context.Context, args runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	klog.V(6).InfoS("Creating new KNIDebug plugin")
	knifeatures.LogState(Name, 2, knifeatures.Names())
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
	lh := klog.FromContext(ctx)
	node := nodeInfo.Node()
	if node == nil {
		// should never happen
		return framework.NewStatus(framework.Error, "node not found")
	}

	// note the fit.go plugin computes this in the prefilter stage. Does this make any practical difference in our context?
	checkRequest(lh.V(LogLevel), pod, nodeInfo)
	return nil // must never fail
}

func frameworkResourceToLoggable(req *framework.Resource) []interface{} {
	items := []interface{}{
		"cpu", humanCPU(req.MilliCPU),
		"memory", humanMemory(req.Memory),
	}

	if req.EphemeralStorage > 0 {
		items = append(items,
			"ephemeral-storage", humanMemory(req.EphemeralStorage),
		)
	}

	if len(req.ScalarResources) == 0 {
		return items
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
// # The resources defined for Overhead should be added to the calculated Resource request sum
//
// Example:
//
// Pod:
//
//	InitContainers
//	  IC1:
//	    CPU: 2
//	    Memory: 1G
//	  IC2:
//	    CPU: 2
//	    Memory: 3G
//	Containers
//	  C1:
//	    CPU: 2
//	    Memory: 1G
//	  C2:
//	    CPU: 1
//	    Memory: 1G
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
func checkRequest(lh logr.Logger, pod *corev1.Pod, nodeInfo *framework.NodeInfo) {
	lh = lh.WithValues("pod", klog.KObj(pod), "podUID", podGetUID(pod), "node", nodeInfo.Node().Name)
	req := computePodResourceRequest(pod)

	if req.MilliCPU == 0 && req.Memory == 0 && req.EphemeralStorage == 0 && len(req.ScalarResources) == 0 {
		lh.Info("target resource requests none")
		return
	}
	lh.Info("target resource requests", frameworkResourceToLoggable(req)...)

	violations := 0
	if availCPU := (nodeInfo.Allocatable.MilliCPU - nodeInfo.Requested.MilliCPU); req.MilliCPU > availCPU {
		lh.Info("insufficient node resources", "resource", "cpu", "request", humanCPU(req.MilliCPU), "available", humanCPU(availCPU))
		violations++
	}
	if availMemory := (nodeInfo.Allocatable.Memory - nodeInfo.Requested.Memory); req.Memory > availMemory {
		lh.Info("insufficient node resources", "resource", "memory", "request", humanMemory(req.Memory), "available", humanMemory(availMemory))
		violations++
	}
	for rName, rQuant := range req.ScalarResources {
		if availQuant := (nodeInfo.Allocatable.ScalarResources[rName] - nodeInfo.Requested.ScalarResources[rName]); rQuant > availQuant {
			lh.Info("insufficient node resources", "resource", rName, "request", rQuant, "available", availQuant)
			violations++
		}
	}

	if violations > 0 {
		return
	}
	lh.Info("enough node resources")
}

func podGetUID(pod *corev1.Pod) string {
	if pod == nil {
		return "<nil>"
	}
	return string(pod.GetUID())
}
