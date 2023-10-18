/*
Copyright 2021 The Kubernetes Authors.

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

package logging

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	topologyv1alpha2 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"
)

func PodRef(pod *corev1.Pod) string {
	if pod == nil {
		return ""
	}
	return pod.Namespace + "/" + pod.Name
}

func NodeResourceTopology(logID, desc string, nrtObj *topologyv1alpha2.NodeResourceTopology) {
	if !klog.V(6).Enabled() {
		// avoid the expensive marshal operation
		return
	}

	ntrJson, err := json.MarshalIndent(nrtObj, "", " ")
	if err != nil {
		klog.V(6).ErrorS(err, "failed to marshal noderesourcetopology object")
		return
	}
	klog.V(6).InfoS(desc, "logID", logID, "noderesourcetopology", string(ntrJson))
}
