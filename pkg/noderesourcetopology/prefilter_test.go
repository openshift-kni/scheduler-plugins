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
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fwk "k8s.io/kube-scheduler/framework"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

func makeTestPod(namespace, name string, containers ...string) *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
	for _, c := range containers {
		pod.Spec.Containers = append(pod.Spec.Containers, v1.Container{Name: c})
	}
	return pod
}

func makeTestPodInfo(pod *v1.Pod) fwk.PodInfo {
	pi, _ := framework.NewPodInfo(pod)
	return pi
}

func TestPreFilter(t *testing.T) {
	tm := &TopologyMatch{}
	cycleState := framework.NewCycleState()

	result, status := tm.PreFilter(context.Background(), cycleState, makeTestPod("ns", "pod"), nil)
	if !status.IsSuccess() {
		t.Fatalf("PreFilter returned non-success: %v", status)
	}
	if result != nil {
		t.Fatalf("PreFilter should return nil result, got %v", result)
	}

	ps, err := readPreemptionStack(cycleState)
	if err != nil {
		t.Fatalf("failed to read PreemptionStack after PreFilter: %v", err)
	}
	if ps.PodsToRemove == nil {
		t.Fatal("PodsToRemove map should be initialized")
	}
	if len(ps.PodsToRemove) != 0 {
		t.Fatalf("PodsToRemove should be empty, got %d entries", len(ps.PodsToRemove))
	}
}

func TestPreFilterExtensions(t *testing.T) {
	tm := &TopologyMatch{}
	ext := tm.PreFilterExtensions()
	if ext == nil {
		t.Fatal("PreFilterExtensions should not return nil")
	}
	if ext != tm {
		t.Fatal("PreFilterExtensions should return the TopologyMatch itself")
	}
}

func TestRemovePod_SingleContainer(t *testing.T) {
	tm := &TopologyMatch{}
	cycleState := framework.NewCycleState()
	tm.PreFilter(context.Background(), cycleState, makeTestPod("ns", "scheduler"), nil)

	victim := makeTestPod("default", "victim-pod", "web")

	status := tm.RemovePod(context.Background(), cycleState, nil, makeTestPodInfo(victim), nil)
	if !status.IsSuccess() {
		t.Fatalf("RemovePod returned non-success: %v", status)
	}

	ps, _ := readPreemptionStack(cycleState)
	storedPod, ok := ps.PodsToRemove["victim-pod"]
	if !ok {
		t.Fatal("expected victim-pod in PodsToRemove")
	}
	if storedPod.Namespace != "default" || storedPod.Name != "victim-pod" {
		t.Fatalf("stored pod mismatch: got %s/%s, want default/victim-pod", storedPod.Namespace, storedPod.Name)
	}
	if len(storedPod.Spec.Containers) != 1 || storedPod.Spec.Containers[0].Name != "web" {
		t.Fatalf("stored pod containers mismatch: got %v", storedPod.Spec.Containers)
	}
}

func TestRemovePod_MultipleContainers(t *testing.T) {
	tm := &TopologyMatch{}
	cycleState := framework.NewCycleState()
	tm.PreFilter(context.Background(), cycleState, makeTestPod("ns", "scheduler"), nil)

	victim := makeTestPod("test-ns", "multi-container-pod", "app", "sidecar", "init")

	status := tm.RemovePod(context.Background(), cycleState, nil, makeTestPodInfo(victim), nil)
	if !status.IsSuccess() {
		t.Fatalf("RemovePod returned non-success: %v", status)
	}

	ps, _ := readPreemptionStack(cycleState)
	storedPod, ok := ps.PodsToRemove["multi-container-pod"]
	if !ok {
		t.Fatal("expected multi-container-pod in PodsToRemove")
	}
	if len(storedPod.Spec.Containers) != 3 {
		t.Fatalf("expected 3 containers, got %d", len(storedPod.Spec.Containers))
	}
	expectedNames := []string{"app", "sidecar", "init"}
	for i, c := range storedPod.Spec.Containers {
		if c.Name != expectedNames[i] {
			t.Errorf("container[%d] name mismatch: got %s, want %s", i, c.Name, expectedNames[i])
		}
	}
}

func TestRemovePod_MultipleVictims(t *testing.T) {
	tm := &TopologyMatch{}
	cycleState := framework.NewCycleState()
	tm.PreFilter(context.Background(), cycleState, makeTestPod("ns", "scheduler"), nil)

	victim1 := makeTestPod("ns", "victim-1", "c1")
	victim2 := makeTestPod("ns", "victim-2", "c2")

	tm.RemovePod(context.Background(), cycleState, nil, makeTestPodInfo(victim1), nil)
	tm.RemovePod(context.Background(), cycleState, nil, makeTestPodInfo(victim2), nil)

	ps, _ := readPreemptionStack(cycleState)
	if len(ps.PodsToRemove) != 2 {
		t.Fatalf("expected 2 victim pods, got %d", len(ps.PodsToRemove))
	}
	if _, ok := ps.PodsToRemove["victim-1"]; !ok {
		t.Error("victim-1 missing")
	}
	if _, ok := ps.PodsToRemove["victim-2"]; !ok {
		t.Error("victim-2 missing")
	}
}

func TestRemovePod_WithoutPreFilter(t *testing.T) {
	tm := &TopologyMatch{}
	cycleState := framework.NewCycleState()

	victim := makeTestPod("ns", "victim", "c1")

	status := tm.RemovePod(context.Background(), cycleState, nil, makeTestPodInfo(victim), nil)
	if status.IsSuccess() {
		t.Fatal("RemovePod should fail when PreFilter was not called (no state in CycleState)")
	}
	if status.Code() != fwk.Error {
		t.Fatalf("expected Error status, got %v", status.Code())
	}
}

func TestRemovePod_PodWithNoContainers(t *testing.T) {
	tm := &TopologyMatch{}
	cycleState := framework.NewCycleState()
	tm.PreFilter(context.Background(), cycleState, makeTestPod("ns", "scheduler"), nil)

	victim := makeTestPod("ns", "empty-pod")

	status := tm.RemovePod(context.Background(), cycleState, nil, makeTestPodInfo(victim), nil)
	if !status.IsSuccess() {
		t.Fatalf("RemovePod should succeed for pod with no containers: %v", status)
	}

	ps, _ := readPreemptionStack(cycleState)
	storedPod, ok := ps.PodsToRemove["empty-pod"]
	if !ok {
		t.Fatal("expected empty-pod in PodsToRemove")
	}
	if len(storedPod.Spec.Containers) != 0 {
		t.Fatalf("expected 0 containers, got %d", len(storedPod.Spec.Containers))
	}
}

func TestClone(t *testing.T) {
	original := &PreemptionStack{
		PodsToRemove: PodsInfo{
			"pod-a": makeTestPod("ns", "pod-a", "c1"),
		},
	}

	cloned := original.Clone().(*PreemptionStack)

	storedPod, ok := cloned.PodsToRemove["pod-a"]
	if !ok {
		t.Fatal("cloned state missing pod-a")
	}
	if storedPod.Name != "pod-a" {
		t.Fatalf("cloned pod name mismatch: got %s, want pod-a", storedPod.Name)
	}

	delete(cloned.PodsToRemove, "pod-a")
	if _, ok := original.PodsToRemove["pod-a"]; !ok {
		t.Fatal("deleting from clone should not affect original")
	}
}

func TestReadPreemptionStack_InvalidState(t *testing.T) {
	cycleState := framework.NewCycleState()
	cycleState.Write(stateVictimPodsKey, &badStateData{})

	_, err := readPreemptionStack(cycleState)
	if err == nil {
		t.Fatal("expected error for invalid state type")
	}
}

func TestAddPod_ReprieveVictim(t *testing.T) {
	tm := &TopologyMatch{}
	cycleState := framework.NewCycleState()
	tm.PreFilter(context.Background(), cycleState, makeTestPod("ns", "scheduler"), nil)

	victim := makeTestPod("ns", "victim-pod", "c1")

	tm.RemovePod(context.Background(), cycleState, nil, makeTestPodInfo(victim), nil)

	ps, _ := readPreemptionStack(cycleState)
	if _, ok := ps.PodsToRemove["victim-pod"]; !ok {
		t.Fatal("victim-pod should be in PodsToRemove before AddPod")
	}

	status := tm.AddPod(context.Background(), cycleState, nil, makeTestPodInfo(victim), nil)
	if !status.IsSuccess() {
		t.Fatalf("AddPod returned non-success: %v", status)
	}

	ps, _ = readPreemptionStack(cycleState)
	if len(ps.PodsToRemove) != 0 {
		t.Fatalf("PodsToRemove should be empty after reprieve, got %d entries", len(ps.PodsToRemove))
	}
}

func TestAddPod_WithoutPreFilter(t *testing.T) {
	tm := &TopologyMatch{}
	cycleState := framework.NewCycleState()

	victim := makeTestPod("ns", "victim", "c1")

	status := tm.AddPod(context.Background(), cycleState, nil, makeTestPodInfo(victim), nil)
	if status.IsSuccess() {
		t.Fatal("AddPod should fail when PreFilter was not called (no state in CycleState)")
	}
	if status.Code() != fwk.Error {
		t.Fatalf("expected Error status, got %v", status.Code())
	}
}

func TestAddPod_AfterMultipleRemoves(t *testing.T) {
	tm := &TopologyMatch{}
	cycleState := framework.NewCycleState()
	tm.PreFilter(context.Background(), cycleState, makeTestPod("ns", "scheduler"), nil)

	victim1 := makeTestPod("ns", "victim-1", "c1")
	victim2 := makeTestPod("ns", "victim-2", "c2")

	tm.RemovePod(context.Background(), cycleState, nil, makeTestPodInfo(victim1), nil)
	tm.RemovePod(context.Background(), cycleState, nil, makeTestPodInfo(victim2), nil)

	status := tm.AddPod(context.Background(), cycleState, nil, makeTestPodInfo(victim1), nil)
	if !status.IsSuccess() {
		t.Fatalf("AddPod returned non-success: %v", status)
	}

	ps, _ := readPreemptionStack(cycleState)
	if _, ok := ps.PodsToRemove["victim-1"]; ok {
		t.Error("victim-1 should be deleted after reprieve")
	}
	storedPod, ok := ps.PodsToRemove["victim-2"]
	if !ok {
		t.Fatal("victim-2 should still be in PodsToRemove")
	}
	if storedPod.Name != "victim-2" {
		t.Errorf("victim-2 pod name mismatch: got %s", storedPod.Name)
	}
}

func TestAddPod_NonVictimPod(t *testing.T) {
	tm := &TopologyMatch{}
	cycleState := framework.NewCycleState()
	tm.PreFilter(context.Background(), cycleState, makeTestPod("ns", "scheduler"), nil)

	victim := makeTestPod("ns", "victim", "c1")
	nonVictim := makeTestPod("ns", "bystander", "c2")

	tm.RemovePod(context.Background(), cycleState, nil, makeTestPodInfo(victim), nil)

	status := tm.AddPod(context.Background(), cycleState, nil, makeTestPodInfo(nonVictim), nil)
	if !status.IsSuccess() {
		t.Fatalf("AddPod returned non-success: %v", status)
	}

	ps, _ := readPreemptionStack(cycleState)
	if _, ok := ps.PodsToRemove["victim"]; !ok {
		t.Error("victim should still be in PodsToRemove")
	}
	if len(ps.PodsToRemove) != 1 {
		t.Errorf("expected 1 victim, got %d", len(ps.PodsToRemove))
	}
}

func TestAddPod_EmptyStack(t *testing.T) {
	tm := &TopologyMatch{}
	cycleState := framework.NewCycleState()
	tm.PreFilter(context.Background(), cycleState, makeTestPod("ns", "scheduler"), nil)

	pod := makeTestPod("ns", "some-pod", "c1")

	status := tm.AddPod(context.Background(), cycleState, nil, makeTestPodInfo(pod), nil)
	if !status.IsSuccess() {
		t.Fatalf("AddPod returned non-success: %v", status)
	}

	ps, _ := readPreemptionStack(cycleState)
	if len(ps.PodsToRemove) != 0 {
		t.Errorf("expected empty PodsToRemove, got %d entries", len(ps.PodsToRemove))
	}
}

func TestGetPods_Empty(t *testing.T) {
	pi := PodsInfo{}
	pods := pi.GetPods()
	if len(pods) != 0 {
		t.Fatalf("expected 0 pods, got %d", len(pods))
	}
}

func TestGetPods_SinglePod(t *testing.T) {
	pi := PodsInfo{
		"pod-a": makeTestPod("ns", "pod-a", "c1"),
	}
	pods := pi.GetPods()
	if len(pods) != 1 {
		t.Fatalf("expected 1 pod, got %d", len(pods))
	}
	if pods[0].Name != "pod-a" {
		t.Errorf("expected pod-a, got %s", pods[0].Name)
	}
}

func TestGetPods_MultiplePods(t *testing.T) {
	pi := PodsInfo{
		"pod-a": makeTestPod("ns", "pod-a", "c1"),
		"pod-b": makeTestPod("ns", "pod-b", "c2", "c3"),
		"pod-c": makeTestPod("other", "pod-c"),
	}
	pods := pi.GetPods()
	if len(pods) != 3 {
		t.Fatalf("expected 3 pods, got %d", len(pods))
	}
	names := map[string]bool{}
	for _, p := range pods {
		names[p.Name] = true
	}
	for _, expected := range []string{"pod-a", "pod-b", "pod-c"} {
		if !names[expected] {
			t.Errorf("missing pod %s in result", expected)
		}
	}
}

func TestGetPods_ReturnsValueCopies(t *testing.T) {
	original := makeTestPod("ns", "pod-a", "c1")
	pi := PodsInfo{
		"pod-a": original,
	}
	pods := pi.GetPods()
	pods[0].Name = "mutated"
	if original.Name != "pod-a" {
		t.Fatal("GetPods should return value copies; original was mutated")
	}
}

type badStateData struct{}

func (b *badStateData) Clone() fwk.StateData { return b }
