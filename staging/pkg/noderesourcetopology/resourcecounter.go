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
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	topologyv1alpha1 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha1"
)

type resourceCounter v1.ResourceList

func resourceCounterFromPod(pod *v1.Pod) resourceCounter {
	rc := make(resourceCounter)
	if pod == nil {
		return rc
	}
	for idx := 0; idx < len(pod.Spec.Containers); idx++ {
		rc.Add(resourceCounter(pod.Spec.Containers[idx].Resources.Limits))
	}
	return rc
}

// Add adds all the resources tracked by `rc2` *in place*.
func (rc resourceCounter) Add(rc2 resourceCounter) {
	// There is no easy and cheap way yet to have the upper bound available for resource counters,
	// so out of necessity we keep going with the assumption that if we got this far, we succesfully
	// passed the filtering stage - which HAD and must have the upper bound for resource counters
	// available. So just add, accumulating all the resources, and don't worry.
	for resName, resQty := range rc2 {
		qty := rc[resName]
		qty.Add(resQty)
		rc[resName] = qty
	}
}

// SubWithChecks subtracts all the resources tracked by `rc2`, reporting errors, *in place*.
func (rc resourceCounter) SubWithChecks(rc2 resourceCounter) []error {
	// Similarly to Add(), if we got this far we already passed sanity filtering checks,
	// so the guards here are an additional, last line of defence.
	// It's very easy to mess up here. Negative counters are hard to debug and will
	// lead to unpredictable outcome, so we are very careful and very verbose in reporting
	// errors.
	var errs []error
	for resName, resQty := range rc2 {
		curQty, ok := rc[resName]
		if !ok {
			errs = append(errs, fmt.Errorf("cannot subtract %q=%v: not in base set", resName, resQty))
			continue
		}

		if curQty.Cmp(resQty) < 0 {
			errs = append(errs, fmt.Errorf("cannot subtract %q=%v: not enough in base set (%v)", resName, resQty, curQty))
			continue
		}
		curQty.Sub(resQty)
		rc[resName] = curQty
	}
	return errs
}

func adjustWithResourceCounters(nrt *topologyv1alpha1.NodeResourceTopology, rc resourceCounter) []error {
	// We cannot predict on which Zone the workload will be placed.
	// And we should totally not guess. So the only safe (and conservative)
	// choice is to decrement the available resources from *all* the zones.
	// This can cause false negatives, but will never cause false positives,
	// which are much worse.
	var errs []error
	for zi := 0; zi < len(nrt.Zones); zi++ {
		zone := &nrt.Zones[zi] // shortcut
		for ri := 0; ri < len(zone.Resources); ri++ {
			zr := &zone.Resources[ri] // shortcut
			qty, ok := rc[v1.ResourceName(zr.Name)]
			if !ok {
				// this is benign; it is totally possible some resources are not
				// available on some zones (think PCI devices), hence we don't
				// even report this error, being an expected condition
				continue
			}
			if zr.Available.Cmp(qty) < 0 {
				// this should happen rarely, and it is likely caused by
				// a bug elsewhere.
				err := fmt.Errorf("cannot decrement resource %q on %q more than %v (asked %v)",
					zr.Name, nrt.Name, zr.Available, qty)
				errs = append(errs, err)
				zr.Available = resource.MustParse("0")
				continue
			}

			zr.Available.Sub(qty)
		}
	}
	return errs
}
