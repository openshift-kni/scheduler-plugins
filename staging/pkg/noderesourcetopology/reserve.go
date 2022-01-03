/*
Copyright 2022 The Kubernetes Authors.

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
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	topologyv1alpha1 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha1"
	informerv1alpha1 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/generated/informers/externalversions/topology/v1alpha1"
)

// Reserve keeps track of the resources allocated by the plugin, but not yet reported by the updaters running on the node.

// Reserve is the functions invoked by the framework at "reserve" extension point.
func (tm *TopologyMatch) Reserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	tm.pendingResources.AddFromPod(nodeName, pod)
	return framework.NewStatus(framework.Success, "")
}

// Unreserve rejects all other Pods in the PodGroup when one of the pods in the group times out.
func (tm *TopologyMatch) Unreserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) {
	errs := tm.pendingResources.RemoveFromPod(nodeName, pod)
	for _, err := range errs {
		klog.V(3).ErrorS(err, "cannot update resources", "nodeName", nodeName, "podNamespace", pod.Namespace, "podName", pod.Name)
	}
}

func (tm *TopologyMatch) getNodeTopology(nodeName string) *topologyv1alpha1.NodeResourceTopology {
	nrt := findNodeTopology(nodeName, tm.lister)
	if nrt == nil {
		return nil
	}
	pa := tm.pendingResources.Get(nodeName)
	if len(pa) == 0 {
		return nrt
	}
	errs := adjustWithResourceCounters(nrt, pa)
	for _, err := range errs {
		klog.V(3).ErrorS(err, "error adjusting node resource topology info", "node", nodeName)
	}
	return nrt
}

// From the noderesourcetopology-aware scheduling perspective:
//
// A "False negative" is when the scheduler falsely believes there are NOT enough available resources on a compute *node* (e.g. across
// all the NUMA zones found on that node). This will cause the node to be ruled out possibly causing scheduling delays if _all_ the nodes
// are mistakenly ruled out. Note this can also cause surprising workload placement on all the nodes, but, as long as all the other
// scheduling requirements are satisfied, this is still correct.
//
// A "False positive" is when the scheduler falsely believes there are MORE available resources on a compute *node* (e.g. on at least
// one NUMA zone found on that node) that is actually the case. This is likely to cause a TopologyAffinityError because the kubelet
// (the Topology Manager in the kubelet) running on that node can actually reject the workload, having a accurate picture of the available
// resources.
//
// In the current architecture, a "false positive" is more severe and hard to recover than a "false negative", because the downsides of
// the latter is a scheduling delay, which can be mitigated with cluster tuning, while the former require more work (descheduling, leftovers
// cleanup...).
// Hence, we try hard to avoid "false positives", to the extent to accept the risk of some "false negatives" if this makes the "false positives"
// much less likely - or impossible.

func setupNodeResourceCountersWithInformer(nodeTopologyInformer informerv1alpha1.NodeResourceTopologyInformer, nrc *nodeResourceCounters) {
	updateHelper := func(obj interface{}, nrc *nodeResourceCounters) {
		nrt, ok := obj.(*topologyv1alpha1.NodeResourceTopology)
		if !ok {
			// TODO: more informative message
			klog.V(5).InfoS("Unexpected object")
			return
		}
		nrc.Flush(nrt.Name)
	}

	informer := nodeTopologyInformer.Informer()
	// After every update to NRTs, we are getting fresher information from the node topology updaters;
	// this means we can safely discarc the pending data we stored, because the information is now consolidated
	// in the NRT objects, and it is automatically available when we run the lister.
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			updateHelper(obj, nrc)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			updateHelper(newObj, nrc)
		},
		DeleteFunc: func(obj interface{}) {
			// this is to make sure we don't leave leftovers around
			updateHelper(obj, nrc)
		},
	})

}

type nodeResourceCounters struct {
	rwlock sync.RWMutex
	nodes  map[string]resourceCounter
}

func newNodeResourceCounters() *nodeResourceCounters {
	return &nodeResourceCounters{
		nodes: make(map[string]resourceCounter),
	}
}

func (nrc *nodeResourceCounters) Get(nodeName string) resourceCounter {
	nrc.rwlock.RLock()
	defer nrc.rwlock.RUnlock()
	return nrc.nodes[nodeName]
}

func (nrc *nodeResourceCounters) Flush(nodeName string) {
	nrc.rwlock.Lock()
	defer nrc.rwlock.Unlock()
	delete(nrc.nodes, nodeName)
}

func (nrc *nodeResourceCounters) AddFromPod(nodeName string, pod *v1.Pod) {
	nrc.rwlock.Lock()
	defer nrc.rwlock.Unlock()
	cur := nrc.nodes[nodeName]
	cur.Add(resourceCounterFromPod(pod))
	nrc.nodes[nodeName] = cur
}

func (nrc *nodeResourceCounters) RemoveFromPod(nodeName string, pod *v1.Pod) []error {
	nrc.rwlock.Lock()
	defer nrc.rwlock.Unlock()
	cur := nrc.nodes[nodeName]
	errs := cur.SubWithChecks(resourceCounterFromPod(pod))
	nrc.nodes[nodeName] = cur
	return errs
}
