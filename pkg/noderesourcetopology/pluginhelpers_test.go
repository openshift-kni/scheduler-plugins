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
	"fmt"
	"testing"

	topologyv1alpha2 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"
	"github.com/k8stopologyawareschedwg/numaplacement"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	apiconfig "sigs.k8s.io/scheduler-plugins/apis/config"
	nrtcache "sigs.k8s.io/scheduler-plugins/pkg/noderesourcetopology/cache"
)

func TestOnlyNonNUMAResources(t *testing.T) {
	numaNodes := NUMANodeList{
		{
			NUMAID: 0,
			Resources: corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewQuantity(8, resource.DecimalSI),
				corev1.ResourceMemory: resource.MustParse("10Gi"),
				"gpu":                 resource.MustParse("1"),
			},
		},
		{
			NUMAID: 1,
			Resources: corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewQuantity(8, resource.DecimalSI),
				corev1.ResourceMemory: resource.MustParse("10Gi"),
				"nic":                 resource.MustParse("1"),
			},
		},
	}
	testCases := []struct {
		description string
		resources   corev1.ResourceList
		expected    bool
	}{
		{
			description: "all resources missing in NUMANodeList",
			resources: corev1.ResourceList{
				"resource1": resource.MustParse("1"),
				"resource2": resource.MustParse("1"),
			},
			expected: true,
		},
		{
			description: "resource is present in both NUMA nodes",
			resources: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("1"),
			},
			expected: false,
		},
		{
			description: "more than resource is present in both NUMA nodes",
			resources: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("1"),
			},
			expected: false,
		},
		{
			description: "resource is present only in NUMA node 0",
			resources: corev1.ResourceList{
				"gpu": resource.MustParse("1"),
			},
			expected: false,
		},
		{
			description: "resource is present only in NUMA node 1",
			resources: corev1.ResourceList{
				"nic": resource.MustParse("1"),
			},
			expected: false,
		},
		{
			description: "two distinct resources from different NUMA nodes",
			resources: corev1.ResourceList{
				"nic": resource.MustParse("1"),
				"gpu": resource.MustParse("1"),
			},
			expected: false,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			result := onlyNonNUMAResources(numaNodes, testCase.resources)
			if result != testCase.expected {
				t.Fatalf("expected %t to equal %t", result, testCase.expected)
			}
		})
	}
}

func TestGetForeignPodsDetectMode(t *testing.T) {
	detectAll := apiconfig.ForeignPodsDetectAll
	detectNone := apiconfig.ForeignPodsDetectNone
	detectOnlyExclusiveResources := apiconfig.ForeignPodsDetectOnlyExclusiveResources

	testCases := []struct {
		description string
		cfg         *apiconfig.NodeResourceTopologyCache
		expected    apiconfig.ForeignPodsDetectMode
	}{
		{
			description: "nil config",
			expected:    apiconfig.ForeignPodsDetectAll,
		},
		{
			description: "empty config",
			cfg:         &apiconfig.NodeResourceTopologyCache{},
			expected:    apiconfig.ForeignPodsDetectAll,
		},
		{
			description: "explicit all",
			cfg: &apiconfig.NodeResourceTopologyCache{
				ForeignPodsDetect: &detectAll,
			},
			expected: apiconfig.ForeignPodsDetectAll,
		},
		{
			description: "explicit disable",
			cfg: &apiconfig.NodeResourceTopologyCache{
				ForeignPodsDetect: &detectNone,
			},
			expected: apiconfig.ForeignPodsDetectNone,
		},
		{
			description: "explicit OnlyExclusiveResources",
			cfg: &apiconfig.NodeResourceTopologyCache{
				ForeignPodsDetect: &detectOnlyExclusiveResources,
			},
			expected: apiconfig.ForeignPodsDetectOnlyExclusiveResources,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			got := getForeignPodsDetectMode(klog.Background(), testCase.cfg)
			if got != testCase.expected {
				t.Errorf("foreign pods detect mode got %v expected %v", got, testCase.expected)
			}
		})
	}
}

func makeNRT(zones ...topologyv1alpha2.Zone) *topologyv1alpha2.NodeResourceTopology {
	return &topologyv1alpha2.NodeResourceTopology{
		Zones: zones,
	}
}

func makeZone(name string, resources ...topologyv1alpha2.ResourceInfo) topologyv1alpha2.Zone {
	return topologyv1alpha2.Zone{
		Name:      name,
		Type:      "Node",
		Resources: resources,
	}
}

func makeResourceInfo(name string, capacity, allocatable, available string) topologyv1alpha2.ResourceInfo {
	return topologyv1alpha2.ResourceInfo{
		Name:        name,
		Capacity:    resource.MustParse(capacity),
		Allocatable: resource.MustParse(allocatable),
		Available:   resource.MustParse(available),
	}
}

func TestAddResourcesToNodeResourcesTopology(t *testing.T) {
	testCases := []struct {
		description     string
		nrt             *topologyv1alpha2.NodeResourceTopology
		numaToResources map[int]corev1.ResourceList
		expectedAvail   map[int]map[string]string // zoneIdx -> resourceName -> expected available
	}{
		{
			description: "empty numaToResources changes nothing",
			nrt: makeNRT(
				makeZone("node-0", makeResourceInfo("cpu", "8", "8", "4")),
				makeZone("node-1", makeResourceInfo("cpu", "8", "8", "6")),
			),
			numaToResources: map[int]corev1.ResourceList{},
			expectedAvail: map[int]map[string]string{
				0: {"cpu": "4"},
				1: {"cpu": "6"},
			},
		},
		{
			description: "add cpu to single zone",
			nrt: makeNRT(
				makeZone("node-0", makeResourceInfo("cpu", "8", "8", "4")),
				makeZone("node-1", makeResourceInfo("cpu", "8", "8", "6")),
			),
			numaToResources: map[int]corev1.ResourceList{
				0: {corev1.ResourceCPU: resource.MustParse("3")},
			},
			expectedAvail: map[int]map[string]string{
				0: {"cpu": "7"},
				1: {"cpu": "6"},
			},
		},
		{
			description: "add to both zones",
			nrt: makeNRT(
				makeZone("node-0",
					makeResourceInfo("cpu", "8", "8", "2"),
					makeResourceInfo("memory", "16Gi", "16Gi", "8Gi"),
				),
				makeZone("node-1",
					makeResourceInfo("cpu", "8", "8", "4"),
					makeResourceInfo("memory", "16Gi", "16Gi", "10Gi"),
				),
			),
			numaToResources: map[int]corev1.ResourceList{
				0: {corev1.ResourceCPU: resource.MustParse("2")},
				1: {corev1.ResourceCPU: resource.MustParse("1"), corev1.ResourceMemory: resource.MustParse("2Gi")},
			},
			expectedAvail: map[int]map[string]string{
				0: {"cpu": "4", "memory": "8Gi"},
				1: {"cpu": "5", "memory": "12Gi"},
			},
		},
		{
			description: "add restores to full allocatable",
			nrt: makeNRT(
				makeZone("node-0", makeResourceInfo("cpu", "4", "4", "0")),
			),
			numaToResources: map[int]corev1.ResourceList{
				0: {corev1.ResourceCPU: resource.MustParse("4")},
			},
			expectedAvail: map[int]map[string]string{
				0: {"cpu": "4"},
			},
		},
		{
			description: "add exceeding allocatable is skipped",
			nrt: makeNRT(
				makeZone("node-0", makeResourceInfo("cpu", "8", "8", "6")),
			),
			numaToResources: map[int]corev1.ResourceList{
				0: {corev1.ResourceCPU: resource.MustParse("5")},
			},
			expectedAvail: map[int]map[string]string{
				0: {"cpu": "6"},
			},
		},
		{
			description: "resource name not in zone is ignored",
			nrt: makeNRT(
				makeZone("node-0", makeResourceInfo("cpu", "8", "8", "4")),
			),
			numaToResources: map[int]corev1.ResourceList{
				0: {"gpu": resource.MustParse("1")},
			},
			expectedAvail: map[int]map[string]string{
				0: {"cpu": "4"},
			},
		},
		{
			description: "NUMA ID not in numaToResources is untouched",
			nrt: makeNRT(
				makeZone("node-0", makeResourceInfo("cpu", "8", "8", "4")),
				makeZone("node-1", makeResourceInfo("cpu", "8", "8", "6")),
			),
			numaToResources: map[int]corev1.ResourceList{
				5: {corev1.ResourceCPU: resource.MustParse("2")},
			},
			expectedAvail: map[int]map[string]string{
				0: {"cpu": "4"},
				1: {"cpu": "6"},
			},
		},
		{
			description: "multiple resources added to same zone",
			nrt: makeNRT(
				makeZone("node-0",
					makeResourceInfo("cpu", "8", "8", "2"),
					makeResourceInfo("memory", "16Gi", "16Gi", "8Gi"),
					makeResourceInfo("hugepages-2Mi", "512Mi", "512Mi", "256Mi"),
				),
			),
			numaToResources: map[int]corev1.ResourceList{
				0: {
					corev1.ResourceCPU:    resource.MustParse("2"),
					corev1.ResourceMemory: resource.MustParse("4Gi"),
				},
			},
			expectedAvail: map[int]map[string]string{
				0: {"cpu": "4", "memory": "12Gi", "hugepages-2Mi": "256Mi"},
			},
		},
		{
			description: "zones out of order - NUMA ID 1 before NUMA ID 0 in slice",
			nrt: makeNRT(
				makeZone("node-1", makeResourceInfo("cpu", "8", "8", "6")),
				makeZone("node-0", makeResourceInfo("cpu", "8", "8", "4")),
			),
			numaToResources: map[int]corev1.ResourceList{
				0: {corev1.ResourceCPU: resource.MustParse("3")},
			},
			expectedAvail: map[int]map[string]string{
				0: {"cpu": "6"}, // slice idx 0 is node-1, untouched
				1: {"cpu": "7"}, // slice idx 1 is node-0, added
			},
		},
		{
			description: "non-NUMA zone is skipped",
			nrt: makeNRT(
				makeZone("socket-0", makeResourceInfo("cpu", "16", "16", "16")),
				makeZone("node-0", makeResourceInfo("cpu", "8", "8", "4")),
			),
			numaToResources: map[int]corev1.ResourceList{
				0: {corev1.ResourceCPU: resource.MustParse("2")},
			},
			expectedAvail: map[int]map[string]string{
				0: {"cpu": "16"}, // socket-0, not a NUMA zone, untouched
				1: {"cpu": "6"},  // node-0, added
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			originalNRT := tc.nrt.DeepCopy()
			result := addResourcesToNodeResourcesTopology(klog.Background(), tc.nrt, tc.numaToResources)

			// verify original is not mutated
			for zIdx, zone := range tc.nrt.Zones {
				for rIdx, res := range zone.Resources {
					orig := originalNRT.Zones[zIdx].Resources[rIdx]
					if res.Available.Cmp(orig.Available) != 0 {
						t.Errorf("original NRT mutated: zone %d resource %s: got %s, want %s",
							zIdx, res.Name, res.Available.String(), orig.Available.String())
					}
				}
			}

			// verify result
			for zIdx, expectedResources := range tc.expectedAvail {
				if zIdx >= len(result.Zones) {
					t.Fatalf("zone index %d out of range (have %d zones)", zIdx, len(result.Zones))
				}
				for _, res := range result.Zones[zIdx].Resources {
					expectedStr, ok := expectedResources[res.Name]
					if !ok {
						continue
					}
					expected := resource.MustParse(expectedStr)
					if res.Available.Cmp(expected) != 0 {
						t.Errorf("zone %d resource %s: available got %s, want %s",
							zIdx, res.Name, res.Available.String(), expected.String())
					}
				}
			}
		})
	}
}

func TestAddResourcesToNodeResourcesTopology_NilNRT(t *testing.T) {
	result := addResourcesToNodeResourcesTopology(klog.Background(), nil, map[int]corev1.ResourceList{
		0: {corev1.ResourceCPU: resource.MustParse("1")},
	})
	if result != nil {
		t.Fatal("expected nil result for nil NRT input")
	}
}

func TestAddResourcesToNodeResourcesTopology_EmptyResources(t *testing.T) {
	nrt := makeNRT(
		makeZone("node-0", makeResourceInfo("cpu", "8", "8", "8")),
	)
	result := addResourcesToNodeResourcesTopology(klog.Background(), nrt, map[int]corev1.ResourceList{})
	if result != nrt {
		t.Fatal("expected same pointer back when numaToResources is empty (no DeepCopy)")
	}
}

type fakeNUMAInfo struct {
	affinities map[string]int // "ns/pod/container" -> numaID
}

func (f *fakeNUMAInfo) Containers() int {
	return len(f.affinities)
}

func (f *fakeNUMAInfo) NUMAAffinity(id numaplacement.ContainerID) (int, error) {
	return f.NUMAAffinityContainer(id.Namespace, id.PodName, id.ContainerName)
}

func (f *fakeNUMAInfo) NUMAAffinityContainer(namespace, podName, containerName string) (int, error) {
	key := fmt.Sprintf("%s/%s/%s", namespace, podName, containerName)
	numaID, ok := f.affinities[key]
	if !ok {
		return -1, fmt.Errorf("unknown container %s", key)
	}
	return numaID, nil
}

func guaranteedPod(namespace, name string, containers ...corev1.Container) corev1.Pod {
	return corev1.Pod{
		Spec:       corev1.PodSpec{Containers: containers},
		ObjectMeta: makeTestPod(namespace, name).ObjectMeta,
	}
}

func guaranteedContainer(name string, cpu, memory string) corev1.Container {
	rl := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(cpu),
		corev1.ResourceMemory: resource.MustParse(memory),
	}
	return corev1.Container{
		Name: name,
		Resources: corev1.ResourceRequirements{
			Requests: rl,
			Limits:   rl,
		},
	}
}

func containerWithResources(name string, requests corev1.ResourceList, limits corev1.ResourceList) corev1.Container {
	return corev1.Container{
		Name: name,
		Resources: corev1.ResourceRequirements{
			Requests: requests,
			Limits:   limits,
		},
	}
}

func TestAccumulateResourcesToDeduct(t *testing.T) {
	const nodeName = "worker-1"

	testCases := []struct {
		description    string
		victims        []corev1.Pod
		affinities     map[string]int
		expectedResult map[int]corev1.ResourceList
	}{
		{
			description:    "no victims produces empty map",
			victims:        []corev1.Pod{},
			affinities:     map[string]int{},
			expectedResult: map[int]corev1.ResourceList{},
		},
		{
			description: "Guaranteed pod with integral CPU and memory on NUMA 0",
			victims: []corev1.Pod{
				guaranteedPod("ns", "victim", guaranteedContainer("app", "4", "8Gi")),
			},
			affinities: map[string]int{
				"ns/victim/app": 0,
			},
			expectedResult: map[int]corev1.ResourceList{
				0: {
					corev1.ResourceCPU:    resource.MustParse("4"),
					corev1.ResourceMemory: resource.MustParse("8Gi"),
				},
			},
		},
		{
			description: "Guaranteed pod with two containers on different NUMAs",
			victims: []corev1.Pod{
				guaranteedPod("ns", "victim",
					guaranteedContainer("web", "2", "4Gi"),
					guaranteedContainer("sidecar", "1", "2Gi"),
				),
			},
			affinities: map[string]int{
				"ns/victim/web":     0,
				"ns/victim/sidecar": 1,
			},
			expectedResult: map[int]corev1.ResourceList{
				0: {
					corev1.ResourceCPU:    resource.MustParse("2"),
					corev1.ResourceMemory: resource.MustParse("4Gi"),
				},
				1: {
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
			},
		},
		{
			description: "BestEffort pod with no resources is skipped entirely",
			victims: []corev1.Pod{
				{
					ObjectMeta: makeTestPod("ns", "besteffort").ObjectMeta,
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "app"}},
					},
				},
			},
			affinities:     map[string]int{"ns/besteffort/app": 0},
			expectedResult: map[int]corev1.ResourceList{},
		},
		{
			description: "Burstable pod with only native resources is skipped",
			victims: []corev1.Pod{
				{
					ObjectMeta: makeTestPod("ns", "burstable").ObjectMeta,
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							containerWithResources("app",
								corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2")},
								nil,
							),
						},
					},
				},
			},
			affinities:     map[string]int{"ns/burstable/app": 0},
			expectedResult: map[int]corev1.ResourceList{},
		},
		{
			description: "Burstable pod with device resource is included",
			victims: []corev1.Pod{
				{
					ObjectMeta: makeTestPod("ns", "bu-device").ObjectMeta,
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							containerWithResources("gpu-worker",
								corev1.ResourceList{
									corev1.ResourceCPU:                   resource.MustParse("500m"),
									corev1.ResourceName("vendor.io/gpu"): resource.MustParse("2"),
								},
								nil,
							),
						},
					},
				},
			},
			affinities: map[string]int{"ns/bu-device/gpu-worker": 0},
			expectedResult: map[int]corev1.ResourceList{
				0: {
					corev1.ResourceName("vendor.io/gpu"): resource.MustParse("2"),
				},
			},
		},
		{
			description: "Guaranteed pod with device and native resources",
			victims: []corev1.Pod{
				guaranteedPod("ns", "gu-mixed",
					containerWithResources("main",
						corev1.ResourceList{
							corev1.ResourceCPU:                   resource.MustParse("4"),
							corev1.ResourceMemory:                resource.MustParse("8Gi"),
							corev1.ResourceName("vendor.io/gpu"): resource.MustParse("1"),
						},
						corev1.ResourceList{
							corev1.ResourceCPU:                   resource.MustParse("4"),
							corev1.ResourceMemory:                resource.MustParse("8Gi"),
							corev1.ResourceName("vendor.io/gpu"): resource.MustParse("1"),
						},
					),
				),
			},
			affinities: map[string]int{"ns/gu-mixed/main": 1},
			expectedResult: map[int]corev1.ResourceList{
				1: {
					corev1.ResourceCPU:                   resource.MustParse("4"),
					corev1.ResourceMemory:                resource.MustParse("8Gi"),
					corev1.ResourceName("vendor.io/gpu"): resource.MustParse("1"),
				},
			},
		},
		{
			description: "Guaranteed pod with fractional CPU is not exclusive for CPU",
			victims: []corev1.Pod{
				guaranteedPod("ns", "frac-cpu",
					containerWithResources("app",
						corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
						corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					),
				),
			},
			affinities: map[string]int{"ns/frac-cpu/app": 0},
			expectedResult: map[int]corev1.ResourceList{
				0: {
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			},
		},
		{
			description: "multiple victims accumulate on same NUMA",
			victims: []corev1.Pod{
				guaranteedPod("ns", "v1", guaranteedContainer("app", "2", "4Gi")),
				guaranteedPod("ns", "v2", guaranteedContainer("app", "3", "2Gi")),
			},
			affinities: map[string]int{
				"ns/v1/app": 0,
				"ns/v2/app": 0,
			},
			expectedResult: map[int]corev1.ResourceList{
				0: {
					corev1.ResourceCPU:    resource.MustParse("5"),
					corev1.ResourceMemory: resource.MustParse("6Gi"),
				},
			},
		},
		{
			description: "container not in NUMA affinity query is skipped",
			victims: []corev1.Pod{
				guaranteedPod("ns", "unknown", guaranteedContainer("app", "2", "4Gi")),
			},
			affinities:     map[string]int{},
			expectedResult: map[int]corev1.ResourceList{},
		},
		{
			description: "BestEffort pod with device resource is included",
			victims: []corev1.Pod{
				{
					ObjectMeta: makeTestPod("ns", "be-device").ObjectMeta,
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							containerWithResources("fpga",
								corev1.ResourceList{
									corev1.ResourceName("acme.io/fpga"): resource.MustParse("1"),
								},
								nil,
							),
						},
					},
				},
			},
			affinities: map[string]int{"ns/be-device/fpga": 1},
			expectedResult: map[int]corev1.ResourceList{
				1: {
					corev1.ResourceName("acme.io/fpga"): resource.MustParse("1"),
				},
			},
		},
		{
			description: "Guaranteed pod with hugepages",
			victims: []corev1.Pod{
				guaranteedPod("ns", "hp-pod",
					containerWithResources("app",
						corev1.ResourceList{
							corev1.ResourceCPU:                   resource.MustParse("2"),
							corev1.ResourceMemory:                resource.MustParse("4Gi"),
							corev1.ResourceName("hugepages-2Mi"): resource.MustParse("512Mi"),
						},
						corev1.ResourceList{
							corev1.ResourceCPU:                   resource.MustParse("2"),
							corev1.ResourceMemory:                resource.MustParse("4Gi"),
							corev1.ResourceName("hugepages-2Mi"): resource.MustParse("512Mi"),
						},
					),
				),
			},
			affinities: map[string]int{"ns/hp-pod/app": 0},
			expectedResult: map[int]corev1.ResourceList{
				0: {
					corev1.ResourceCPU:                   resource.MustParse("2"),
					corev1.ResourceMemory:                resource.MustParse("4Gi"),
					corev1.ResourceName("hugepages-2Mi"): resource.MustParse("512Mi"),
				},
			},
		},
		{
			description: "no NUMA affinity query for node produces empty map",
			victims: []corev1.Pod{
				guaranteedPod("ns", "victim", guaranteedContainer("app", "4", "8Gi")),
			},
			affinities:     nil,
			expectedResult: map[int]corev1.ResourceList{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			nrt := makeNRT(
				makeZone("node-0", makeResourceInfo("cpu", "32", "32", "32"), makeResourceInfo("memory", "64Gi", "64Gi", "64Gi")),
				makeZone("node-1", makeResourceInfo("cpu", "32", "32", "32"), makeResourceInfo("memory", "64Gi", "64Gi", "64Gi")),
			)
			nrt.Name = nodeName

			info := nrtcache.CachedNRTInfo{}
			if tc.affinities != nil {
				info.NUMAAffinityQuery = map[string]numaplacement.Info{
					nodeName: &fakeNUMAInfo{affinities: tc.affinities},
				}
			}

			result := accumulateResourcesToDeduct(klog.Background(), nrt, info, tc.victims)

			if len(result) != len(tc.expectedResult) {
				t.Fatalf("expected %d NUMA entries, got %d: %v", len(tc.expectedResult), len(result), result)
			}
			for numaID, expectedResources := range tc.expectedResult {
				gotResources, ok := result[numaID]
				if !ok {
					t.Fatalf("missing NUMA %d in result", numaID)
				}
				for resName, expectedQty := range expectedResources {
					gotQty, ok := gotResources[resName]
					if !ok {
						t.Errorf("NUMA %d: missing resource %s", numaID, resName)
						continue
					}
					if gotQty.Cmp(expectedQty) != 0 {
						t.Errorf("NUMA %d resource %s: got %s, want %s", numaID, resName, gotQty.String(), expectedQty.String())
					}
				}
				for resName, qty := range gotResources {
					if _, ok := expectedResources[resName]; !ok {
						t.Errorf("NUMA %d: unexpected resource %s = %s", numaID, resName, qty.String())
					}
				}
			}
		})
	}
}
