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

package customization

import (
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	apiconfig "sigs.k8s.io/scheduler-plugins/apis/config"
)

func RejectNode(nodeName string, scoreStrategyType apiconfig.ScoringStrategyType, reason string) *framework.Status {
	if scoreStrategyType == apiconfig.LeastNUMANodes {
		klog.V(4).InfoS("allowed by scoring strategy", "node", nodeName, "scoreStrategy", scoreStrategyType)
		return nil
	}
	klog.V(2).InfoS(reason, "node", nodeName)
	return framework.NewStatus(framework.Unschedulable, reason)
}
