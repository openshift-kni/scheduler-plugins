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
	"testing"

	topologyv1alpha1 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFromPodWithEmptyValue(t *testing.T) {
	rc1 := resourceCounterFromPod(nil)
	if len(rc1) > 0 {
		t.Fatalf("non-zero resource counters from nil pod")
	}

	rc2 := resourceCounterFromPod(&v1.Pod{})
	if len(rc2) > 0 {
		t.Fatalf("non-zero resource counters from zero pod")
	}
}

const (
	nicName = "vendor_A.com/nic"
	gpuName = "vendor_B.com/gpu"
)

func TestFromPodWithMultipleContainers(t *testing.T) {
	pod := v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "cnt-0",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:           resource.MustParse("4"),
							v1.ResourceMemory:        resource.MustParse("2Gi"),
							v1.ResourceName(nicName): resource.MustParse("2"),
						},
					},
				},
				{
					Name: "cnt-1",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:           resource.MustParse("1"),
							v1.ResourceMemory:        resource.MustParse("1Gi"),
							v1.ResourceName(gpuName): resource.MustParse("1"),
						},
					},
				},
			},
		},
	}

	rc := resourceCounterFromPod(&pod)
	if len(rc) == 0 {
		t.Fatalf("missing resource counters from non-empty pod")
	}

	if cpus := rc[v1.ResourceCPU]; cpus.Cmp(resource.MustParse("5")) != 0 {
		t.Errorf("unexpected resource %q: desired %v got %v", v1.ResourceCPU, "5", cpus)
	}
	if mems := rc[v1.ResourceMemory]; mems.Cmp(resource.MustParse("3Gi")) != 0 {
		t.Errorf("unexpected resource %q: desired %v got %v", v1.ResourceMemory, "3Gi", mems)
	}
	if gpus := rc[v1.ResourceName(gpuName)]; gpus.Cmp(resource.MustParse("1")) != 0 {
		t.Errorf("unexpected resource %q: desired %v got %v", gpuName, "1", gpus)
	}
	if nics := rc[v1.ResourceName(nicName)]; nics.Cmp(resource.MustParse("2")) != 0 {
		t.Errorf("unexpected resource %q: desired %v got %v", nicName, "2", nics)
	}
}

func TestFromPodWithMultipleContainersWithAdd(t *testing.T) {
	pod := v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "cnt-0",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:           resource.MustParse("4"),
							v1.ResourceMemory:        resource.MustParse("2Gi"),
							v1.ResourceName(nicName): resource.MustParse("2"),
						},
					},
				},
				{
					Name: "cnt-1",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:           resource.MustParse("1"),
							v1.ResourceMemory:        resource.MustParse("1Gi"),
							v1.ResourceName(gpuName): resource.MustParse("1"),
						},
					},
				},
			},
		},
	}

	rc := resourceCounterFromPod(&pod)

	pod2 := v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "cnt-0",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:           resource.MustParse("8"),
							v1.ResourceMemory:        resource.MustParse("16Gi"),
							v1.ResourceName(gpuName): resource.MustParse("2"),
						},
					},
				},
			},
		},
	}

	rc2 := resourceCounterFromPod(&pod2)
	rc.Add(rc2)

	if len(rc) == 0 {
		t.Fatalf("missing resource counters from non-empty pod")
	}

	if cpus := rc[v1.ResourceCPU]; cpus.Cmp(resource.MustParse("13")) != 0 {
		t.Errorf("unexpected resource %q: desired %v got %v", v1.ResourceCPU, "13", cpus)
	}
	if mems := rc[v1.ResourceMemory]; mems.Cmp(resource.MustParse("19Gi")) != 0 {
		t.Errorf("unexpected resource %q: desired %v got %v", v1.ResourceMemory, "19Gi", mems)
	}
	if gpus := rc[v1.ResourceName(gpuName)]; gpus.Cmp(resource.MustParse("3")) != 0 {
		t.Errorf("unexpected resource %q: desired %v got %v", gpuName, "3", gpus)
	}
	if nics := rc[v1.ResourceName(nicName)]; nics.Cmp(resource.MustParse("2")) != 0 {
		t.Errorf("unexpected resource %q: desired %v got %v", nicName, "2", nics)
	}
}

func TestSubPodFromEmpty(t *testing.T) {
	pod := v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "cnt-0",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:           resource.MustParse("4"),
							v1.ResourceMemory:        resource.MustParse("2Gi"),
							v1.ResourceName(nicName): resource.MustParse("2"),
						},
					},
				},
				{
					Name: "cnt-1",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:           resource.MustParse("1"),
							v1.ResourceMemory:        resource.MustParse("1Gi"),
							v1.ResourceName(gpuName): resource.MustParse("1"),
						},
					},
				},
			},
		},
	}

	rc := resourceCounterFromPod(&v1.Pod{})
	errs := rc.SubWithChecks(resourceCounterFromPod(&pod))

	if len(rc) != 0 {
		t.Errorf("created resources while subtracting from empty")
	}
	if len(errs) == 0 {
		t.Errorf("missing errors while subtracting from empty")
	}
}

func TestSubPodHandleNegatives(t *testing.T) {
	pod1 := v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "cnt-0",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:           resource.MustParse("4"),
							v1.ResourceMemory:        resource.MustParse("2Gi"),
							v1.ResourceName(nicName): resource.MustParse("2"),
						},
					},
				},
			},
		},
	}

	pod2 := v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "cnt-0",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:           resource.MustParse("1"),
							v1.ResourceMemory:        resource.MustParse("1Gi"),
							v1.ResourceName(gpuName): resource.MustParse("1"),
						},
					},
				},
			},
		},
	}

	rc := resourceCounterFromPod(&pod1)
	errs := rc.SubWithChecks(resourceCounterFromPod(&pod2))

	if len(errs) == 0 {
		t.Errorf("missing errors while subtracting from empty")
	}

	if cpus := rc[v1.ResourceCPU]; cpus.Cmp(resource.MustParse("3")) != 0 {
		t.Errorf("unexpected resource %q: desired %v got %v", v1.ResourceCPU, "3", cpus)
	}
	if mems := rc[v1.ResourceMemory]; mems.Cmp(resource.MustParse("1Gi")) != 0 {
		t.Errorf("unexpected resource %q: desired %v got %v", v1.ResourceMemory, "1Gi", mems)
	}
	if _, ok := rc[v1.ResourceName(gpuName)]; ok {
		t.Errorf("created resource %q by subtraction", gpuName)
	}
	if nics := rc[v1.ResourceName(nicName)]; nics.Cmp(resource.MustParse("2")) != 0 {
		t.Errorf("unexpected resource %q: desired %v got %v", nicName, "2", nics)
	}
}

func TestAddThenSub(t *testing.T) {
	pod := v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "cnt-0",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:           resource.MustParse("4"),
							v1.ResourceMemory:        resource.MustParse("2Gi"),
							v1.ResourceName(nicName): resource.MustParse("2"),
						},
					},
				},
				{
					Name: "cnt-1",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:           resource.MustParse("1"),
							v1.ResourceMemory:        resource.MustParse("1Gi"),
							v1.ResourceName(gpuName): resource.MustParse("1"),
						},
					},
				},
			},
		},
	}

	rc := resourceCounterFromPod(&pod)

	pod2 := v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "cnt-0",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:           resource.MustParse("8"),
							v1.ResourceMemory:        resource.MustParse("16Gi"),
							v1.ResourceName(gpuName): resource.MustParse("2"),
						},
					},
				},
			},
		},
	}

	rc2 := resourceCounterFromPod(&pod2)
	rc.Add(rc2)

	// quick smoke test
	if mems := rc[v1.ResourceMemory]; mems.Cmp(resource.MustParse("19Gi")) != 0 {
		t.Errorf("unexpected resource %q: desired %v got %v", v1.ResourceMemory, "19Gi", mems)
	}

	rc.SubWithChecks(rc2)

	if len(rc) == 0 {
		t.Fatalf("missing resource counters from non-empty pod")
	}

	if cpus := rc[v1.ResourceCPU]; cpus.Cmp(resource.MustParse("5")) != 0 {
		t.Errorf("unexpected resource %q: desired %v got %v", v1.ResourceCPU, "5", cpus)
	}
	if mems := rc[v1.ResourceMemory]; mems.Cmp(resource.MustParse("3Gi")) != 0 {
		t.Errorf("unexpected resource %q: desired %v got %v", v1.ResourceMemory, "3Gi", mems)
	}
	if gpus := rc[v1.ResourceName(gpuName)]; gpus.Cmp(resource.MustParse("1")) != 0 {
		t.Errorf("unexpected resource %q: desired %v got %v", gpuName, "1", gpus)
	}
	if nics := rc[v1.ResourceName(nicName)]; nics.Cmp(resource.MustParse("2")) != 0 {
		t.Errorf("unexpected resource %q: desired %v got %v", nicName, "2", nics)
	}
}

func TestAdjustWithResourceCounters(t *testing.T) {
	nrt := &topologyv1alpha1.NodeResourceTopology{
		ObjectMeta:       metav1.ObjectMeta{Name: "node"},
		TopologyPolicies: []string{string(topologyv1alpha1.SingleNUMANodePodLevel)},
		Zones: topologyv1alpha1.ZoneList{
			{
				Name: "node-0",
				Type: "Node",
				Resources: topologyv1alpha1.ResourceInfoList{
					MakeTopologyResInfo(cpu, "20", "20"),
					MakeTopologyResInfo(memory, "32Gi", "32Gi"),
				},
			},
			{
				Name: "node-1",
				Type: "Node",
				Resources: topologyv1alpha1.ResourceInfoList{
					MakeTopologyResInfo(cpu, "20", "20"),
					MakeTopologyResInfo(memory, "32Gi", "32Gi"),
					MakeTopologyResInfo(nicName, "8", "8"),
				},
			},
		},
	}

	pod := v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "cnt-0",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:           resource.MustParse("16"),
							v1.ResourceMemory:        resource.MustParse("4Gi"),
							v1.ResourceName(nicName): resource.MustParse("2"),
						},
					},
				},
				{
					Name: "cnt-1",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("2"),
							v1.ResourceMemory: resource.MustParse("2Gi"),
						},
					},
				},
			},
		},
	}

	rc := resourceCounterFromPod(&pod)
	errs := adjustWithResourceCounters(nrt, rc)
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}

	cpuInfo0 := findResourceInfo(nrt.Zones[0].Resources, cpu)
	if cpuInfo0.Capacity.Cmp(resource.MustParse("20")) != 0 {
		t.Errorf("bad capacity for resource %q on zone %d: expected %v got %v", cpu, 0, "20", cpuInfo0.Capacity)
	}
	if cpuInfo0.Available.Cmp(resource.MustParse("2")) != 0 {
		t.Errorf("bad availability for resource %q on zone %d: expected %v got %v", cpu, 0, "2", cpuInfo0.Available)
	}
	cpuInfo1 := findResourceInfo(nrt.Zones[1].Resources, cpu)
	if cpuInfo1.Capacity.Cmp(resource.MustParse("20")) != 0 {
		t.Errorf("bad capacity for resource %q on zone %d: expected %v got %v", cpu, 1, "20", cpuInfo1.Capacity)
	}
	if cpuInfo1.Available.Cmp(resource.MustParse("2")) != 0 {
		t.Errorf("bad availability for resource %q on zone %d: expected %v got %v", cpu, 1, "2", cpuInfo1.Available)
	}

	memInfo0 := findResourceInfo(nrt.Zones[0].Resources, memory)
	if memInfo0.Capacity.Cmp(resource.MustParse("32Gi")) != 0 {
		t.Errorf("bad capacity for resource %q on zone %d: expected %v got %v", memory, 0, "32Gi", memInfo0.Capacity)
	}
	if memInfo0.Available.Cmp(resource.MustParse("26Gi")) != 0 {
		t.Errorf("bad availability for resource %q on zone %d: expected %v got %v", memory, 0, "26Gi", memInfo0.Available)
	}
	memInfo1 := findResourceInfo(nrt.Zones[1].Resources, memory)
	if memInfo1.Capacity.Cmp(resource.MustParse("32Gi")) != 0 {
		t.Errorf("bad capacity for resource %q on zone %d: expected %v got %v", memory, 1, "32Gi", memInfo1.Capacity)
	}
	if memInfo1.Available.Cmp(resource.MustParse("26Gi")) != 0 {
		t.Errorf("bad availability for resource %q on zone %d: expected %v got %v", memory, 1, "26Gi", memInfo1.Available)
	}

	devInfo0 := findResourceInfo(nrt.Zones[0].Resources, nicName)
	if devInfo0 != nil {
		t.Errorf("unexpected device %q on zone %d", nicName, 0)
	}

	devInfo1 := findResourceInfo(nrt.Zones[1].Resources, nicName)
	if devInfo1 == nil {
		t.Fatalf("expected device %q on zone %d, but missing", nicName, 1)
	}
	if devInfo1.Capacity.Cmp(resource.MustParse("8")) != 0 {
		t.Errorf("bad capacity for resource %q on zone %d: expected %v got %v", nicName, 1, "8", devInfo1.Capacity)
	}
	if devInfo1.Available.Cmp(resource.MustParse("6")) != 0 {
		t.Errorf("bad availability for resource %q on zone %d: expected %v got %v", nicName, 1, "6", devInfo1.Available)
	}
}

func findResourceInfo(rinfos []topologyv1alpha1.ResourceInfo, name string) *topologyv1alpha1.ResourceInfo {
	for idx := 0; idx < len(rinfos); idx++ {
		if rinfos[idx].Name == name {
			return &rinfos[idx]
		}
	}
	return nil
}
