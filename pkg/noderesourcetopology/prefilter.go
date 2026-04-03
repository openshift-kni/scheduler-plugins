/*
Copyright 2026 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package noderesourcetopology

import (
	"context"

	v1 "k8s.io/api/core/v1"
	fwk "k8s.io/kube-scheduler/framework"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// The extension point for is implemented only to be able to use AddPod/RemovePod functions during preemption flow

func (tm *TopologyMatch) PreFilter(ctx context.Context, state fwk.CycleState, pod *v1.Pod) (*framework.PreFilterResult, *fwk.Status) {
	return nil, fwk.NewStatus(fwk.Success, "")
}

func (tm *TopologyMatch) PreFilterExtensions() framework.PreFilterExtensions {
	// // AddPod is called by the framework while trying to evaluate the impact
	// // of adding podToAdd to the node while scheduling podToSchedule.
	//AddPod(ctx context.Context, state fwk.CycleState, podToSchedule *v1.Pod, podInfoToAdd fwk.PodInfo, nodeInfo fwk.NodeInfo) *fwk.Status
	// // RemovePod is called by the framework while trying to evaluate the impact
	// // of removing podToRemove from the node while scheduling podToSchedule.
	// RemovePod(ctx context.Context, state fwk.CycleState, podToSchedule *v1.Pod, podInfoToRemove fwk.PodInfo, nodeInfo fwk.NodeInfo) *fwk.Status

	return tm
}

func (tm *TopologyMatch) AddPod(ctx context.Context, state fwk.CycleState, podToSchedule *v1.Pod, podInfoToAdd fwk.PodInfo, nodeInfo fwk.NodeInfo) *fwk.Status {
	return nil //considered success
}

func (tm *TopologyMatch) RemovePod(ctx context.Context, state fwk.CycleState, podToSchedule *v1.Pod, podInfoToRemove fwk.PodInfo, nodeInfo fwk.NodeInfo) *fwk.Status {
	return nil //considered success
}
