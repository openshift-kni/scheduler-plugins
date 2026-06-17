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

package noderesourcetopology

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/wait"
	k8scache "k8s.io/client-go/tools/cache"
	v1qos "k8s.io/kubernetes/pkg/apis/core/v1/helper/qos"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	"github.com/go-logr/logr"
	topologyv1alpha2 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"
	"github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2/helper"
	"github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2/helper/numanode"

	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	apiconfig "sigs.k8s.io/scheduler-plugins/apis/config"
	nrtcache "sigs.k8s.io/scheduler-plugins/pkg/noderesourcetopology/cache"
	"sigs.k8s.io/scheduler-plugins/pkg/noderesourcetopology/logging"
	"sigs.k8s.io/scheduler-plugins/pkg/noderesourcetopology/podprovider"
	"sigs.k8s.io/scheduler-plugins/pkg/noderesourcetopology/resourcerequests"
	"sigs.k8s.io/scheduler-plugins/pkg/noderesourcetopology/stringify"
)

const (
	maxNUMAId = 64
)

func initNodeTopologyInformer(ctx context.Context, lh logr.Logger,
	tcfg *apiconfig.NodeResourceTopologyMatchArgs, handle framework.Handle) (nrtcache.Interface, error) {
	client, err := ctrlclient.NewWithWatch(handle.KubeConfig(), ctrlclient.Options{Scheme: scheme})
	if err != nil {
		lh.Error(err, "cannot create client for NodeTopologyResource", "kubeConfig", handle.KubeConfig())
		return nil, err
	}

	if tcfg.DiscardReservedNodes {
		return nrtcache.NewDiscardReserved(lh.WithName(logging.SubsystemNRTCache), client), nil
	}

	if tcfg.CacheResyncPeriodSeconds <= 0 {
		return nrtcache.NewPassthrough(lh.WithName(logging.SubsystemNRTCache), client), nil
	}

	podSharedInformer, podLister, isPodRelevant := podprovider.NewFromHandle(lh, handle, tcfg.Cache)

	nrtCache, err := nrtcache.NewOverReserve(ctx, lh.WithName(logging.SubsystemNRTCache), tcfg.Cache, client, podLister, isPodRelevant)
	if err != nil {
		return nil, err
	}

	initNodeTopologyForeignPodsDetection(lh, tcfg.Cache, handle, podSharedInformer, nrtCache)

	resyncPeriod := time.Duration(tcfg.CacheResyncPeriodSeconds) * time.Second
	go wait.Forever(nrtCache.Resync, resyncPeriod)

	lh.V(3).Info("enable NodeTopology cache (needs the Reserve plugin)", "resyncPeriod", resyncPeriod)

	return nrtCache, nil
}

func initNodeTopologyForeignPodsDetection(lh logr.Logger, cfg *apiconfig.NodeResourceTopologyCache, handle framework.Handle, podSharedInformer k8scache.SharedInformer, nrtCache *nrtcache.OverReserve) {
	foreignPodsDetect := getForeignPodsDetectMode(lh, cfg)

	if foreignPodsDetect == apiconfig.ForeignPodsDetectNone {
		lh.Info("foreign pods detection disabled by configuration")
		return
	}
	fwk, ok := handle.(framework.Framework)
	if !ok {
		lh.Info("cannot determine the scheduler profile names - no foreign pod detection enabled")
		return
	}

	profileName := fwk.ProfileName()
	lh.Info("setting up foreign pods detection", "name", profileName, "mode", foreignPodsDetect)

	if foreignPodsDetect == apiconfig.ForeignPodsDetectOnlyExclusiveResources {
		nrtcache.TrackOnlyForeignPodsWithExclusiveResources()
	} else {
		nrtcache.TrackAllForeignPods()
	}
	nrtcache.RegisterSchedulerProfileName(lh.WithName(logging.SubsystemForeignPods), profileName)
	nrtcache.SetupForeignPodsDetector(lh.WithName(logging.SubsystemForeignPods), profileName, podSharedInformer, nrtCache)
}

func createNUMANodeList(lh logr.Logger, zones topologyv1alpha2.ZoneList) NUMANodeList {
	numaIDToZoneIDx := make([]int, maxNUMAId)
	nodes := NUMANodeList{}
	// filter non Node zones and create idToIdx lookup array
	for i, zone := range zones {
		if zone.Type != helper.ZoneTypeNUMANode {
			continue
		}

		numaID, err := numanode.NameToID(zone.Name)
		if err != nil || numaID > maxNUMAId {
			lh.Error(err, "error getting the numaID", "zone", zone.Name, "numaID", numaID)
			continue
		}

		numaIDToZoneIDx[numaID] = i

		resources := extractResources(zone)
		numaItems := []interface{}{"numaCell", numaID}
		lh.V(6).Info("extracted NUMA resources", stringify.ResourceListToLoggableWithValues(numaItems, resources)...)
		nodes = append(nodes, NUMANode{NUMAID: numaID, Resources: resources})
	}

	// iterate over nodes and fill them with Costs
	for i, node := range nodes {
		nodes[i] = *node.WithCosts(extractCosts(zones[numaIDToZoneIDx[node.NUMAID]].Costs))
	}

	return nodes
}

func extractCosts(costs topologyv1alpha2.CostList) map[int]int {
	nodeCosts := make(map[int]int)

	// return early if CostList is missing
	if len(costs) == 0 {
		return nodeCosts
	}

	for _, cost := range costs {
		numaID, err := numanode.NameToID(cost.Name)
		if err != nil || numaID > maxNUMAId {
			continue
		}
		nodeCosts[numaID] = int(cost.Value)
	}

	return nodeCosts
}

func extractResources(zone topologyv1alpha2.Zone) corev1.ResourceList {
	res := make(corev1.ResourceList)
	for _, resInfo := range zone.Resources {
		res[corev1.ResourceName(resInfo.Name)] = resInfo.Available.DeepCopy()
	}
	return res
}

func onlyNonNUMAResources(numaNodes NUMANodeList, resources corev1.ResourceList) bool {
	for resourceName := range resources {
		for _, node := range numaNodes {
			if _, ok := node.Resources[resourceName]; ok {
				return false
			}
		}
	}

	return true
}

func getForeignPodsDetectMode(lh logr.Logger, cfg *apiconfig.NodeResourceTopologyCache) apiconfig.ForeignPodsDetectMode {
	var foreignPodsDetect apiconfig.ForeignPodsDetectMode
	if cfg != nil && cfg.ForeignPodsDetect != nil {
		foreignPodsDetect = *cfg.ForeignPodsDetect
	} else { // explicitly set to nil?
		foreignPodsDetect = apiconfig.ForeignPodsDetectAll
		lh.Info("foreign pods detection value missing", "fallback", foreignPodsDetect)
	}
	return foreignPodsDetect
}

func updateNRTForPreemption(lh logr.Logger, nrt *topologyv1alpha2.NodeResourceTopology, info nrtcache.CachedNRTInfo, victims []corev1.Pod) *topologyv1alpha2.NodeResourceTopology {
	numaToResourcesToDeduct := accumulateResourcesToDeduct(lh, nrt, info, victims)
	return addResourcesToNodeResourcesTopology(lh, nrt, numaToResourcesToDeduct)
}

func accumulateResourcesToDeduct(lh logr.Logger, nrt *topologyv1alpha2.NodeResourceTopology, info nrtcache.CachedNRTInfo, victims []corev1.Pod) map[int]corev1.ResourceList {
	// we don't know how preemption is choosing the victims, but for the NRT scheduler, we only care about what
	// conatiners are found in the query which are eligable to have their resources reflected in the NRT;
	// if the container ID is not found we continue to the next.
	numaToResourcesToDeduct := make(map[int]v1.ResourceList)
	for _, victim := range victims {
		// pod level filtering
		pQos := v1qos.GetPodQOS(&victim)
		if pQos != v1.PodQOSGuaranteed && !resourcerequests.IncludeNonNative(&victim) {
			continue
		}

		for _, container := range victim.Spec.Containers {
			// container level filtering
			if !resourcerequests.AreExclusiveForContainer(pQos, &container) {
				continue
			}

			query, ok := info.NUMAAffinityQuery[nrt.Name]
			if !ok {
				continue
			}

			numaID, err := query.NUMAAffinityContainer(victim.Namespace, victim.Name, container.Name)
			if err != nil {
				lh.V(2).Error(err, "failed to get NUMA affinity for container", "namespace", victim.Namespace, "pod", victim.Name, "container", container.Name)
				continue
			}

			if _, ok := numaToResourcesToDeduct[numaID]; !ok {
				numaToResourcesToDeduct[numaID] = make(v1.ResourceList)
			}

			for cntRes, cntQty := range container.Resources.Requests {
				//resource level filtering
				if !resourcerequests.IsExclusive(pQos, cntRes, cntQty) {
					continue
				}

				currentQty, ok := numaToResourcesToDeduct[numaID][cntRes]
				if !ok {
					currentQty = resource.Quantity{}
				}
				currentQty.Add(cntQty)
				numaToResourcesToDeduct[numaID][cntRes] = currentQty
			}
		}
	}
	return numaToResourcesToDeduct
}
func addResourcesToNodeResourcesTopology(lh logr.Logger, nrt *topologyv1alpha2.NodeResourceTopology, numaToResources map[int]corev1.ResourceList) *topologyv1alpha2.NodeResourceTopology {
	if nrt == nil || len(numaToResources) == 0 {
		return nrt
	}
	updatedNRT := nrt.DeepCopy()
	for zoneIdx, zone := range updatedNRT.Zones {
		numaID, err := numanode.NameToID(zone.Name)
		if err != nil {
			lh.V(6).Info("skipping non-NUMA zone", "zone", zone.Name, "error", err)
			continue
		}
		resListToSub, ok := numaToResources[numaID]
		if !ok {
			continue
		}

		for resName, resQty := range resListToSub {
			for resIdx := range zone.Resources {
				resource := &updatedNRT.Zones[zoneIdx].Resources[resIdx]
				if resource.Name != string(resName) {
					continue
				}

				tmp := resource.Available.DeepCopy()
				tmp.Add(resQty)
				if tmp.Cmp(resource.Allocatable) > 0 {
					lh.V(2).Error(nil, "resource release request exceeds NUMA allocatable",
						"zone", zone.Name, "resource", resName,
						"allocatable", resource.Allocatable.String(), "requestToRelease", resQty.String(), "numaID", numaID)
					break
				}
				resource.Available.Add(resQty)
				break
			}
		}
	}
	return updatedNRT
}
