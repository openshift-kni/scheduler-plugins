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
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

func TestFrameworkResourceToLoggable(t *testing.T) {
	tests := []struct {
		name      string
		resources *framework.Resource
		expected  string
	}{
		{
			name:      "empty",
			resources: &framework.Resource{},
			expected:  ` cpu="0 (0)" memory="0 (0 B)"`,
		},
		{
			name: "only cpus",
			resources: &framework.Resource{
				MilliCPU: 16000,
			},
			expected: ` cpu="16000 (16)" memory="0 (0 B)"`,
		},
		{
			name: "only Memory",
			resources: &framework.Resource{
				Memory: 16 * 1024 * 1024 * 1024,
			},
			expected: ` cpu="0 (0)" memory="17179869184 (16 GiB)"`,
		},
		{
			name: "cpus and Memory",
			resources: &framework.Resource{
				MilliCPU: 24000,
				Memory:   16 * 1024 * 1024 * 1024,
			},
			expected: ` cpu="24000 (24)" memory="17179869184 (16 GiB)"`,
		},
		{
			name: "cpus, Memory, hugepages-2Mi",
			resources: &framework.Resource{
				MilliCPU: 24000,
				Memory:   16 * 1024 * 1024 * 1024,
				ScalarResources: map[corev1.ResourceName]int64{
					corev1.ResourceName("hugepages-2Mi"): 1 * 1024 * 1024 * 1024,
				},
			},
			expected: ` cpu="24000 (24)" memory="17179869184 (16 GiB)" hugepages-2Mi="1073741824 (1.0 GiB)"`,
		},
		{
			name: "cpus, Memory, hugepages-2Mi, ephemeral-storage",
			resources: &framework.Resource{
				MilliCPU:         24000,
				Memory:           16 * 1024 * 1024 * 1024,
				EphemeralStorage: 128 * 1024 * 1204 * 1024,
				ScalarResources: map[corev1.ResourceName]int64{
					corev1.ResourceName("hugepages-2Mi"): 1 * 1024 * 1024 * 1024,
				},
			},
			expected: ` cpu="24000 (24)" memory="17179869184 (16 GiB)" ephemeral-storage="161598144512 (150 GiB)" hugepages-2Mi="1073741824 (1.0 GiB)"`,
		},
		{
			name: "cpus, Memory, hugepages-2Mi, hugepages-1Gi",
			resources: &framework.Resource{
				MilliCPU: 24000,
				Memory:   16 * 1024 * 1024 * 1024,
				ScalarResources: map[corev1.ResourceName]int64{
					corev1.ResourceName("hugepages-2Mi"): 1 * 1024 * 1024 * 1024,
					corev1.ResourceName("hugepages-1Gi"): 2 * 1024 * 1024 * 1024,
				},
			},
			expected: ` cpu="24000 (24)" memory="17179869184 (16 GiB)" hugepages-1Gi="2147483648 (2.0 GiB)" hugepages-2Mi="1073741824 (1.0 GiB)"`,
		},
		{
			name: "cpus, Memory, hugepages-2Mi, devices",
			resources: &framework.Resource{
				MilliCPU: 24000,
				Memory:   16 * 1024 * 1024 * 1024,
				ScalarResources: map[corev1.ResourceName]int64{
					corev1.ResourceName("hugepages-2Mi"):         1 * 1024 * 1024 * 1024,
					corev1.ResourceName("example.com/netdevice"): 16,
					corev1.ResourceName("awesome.net/gpu"):       4,
				},
			},
			expected: ` cpu="24000 (24)" memory="17179869184 (16 GiB)" awesome.net/gpu="4" example.com/netdevice="16" hugepages-2Mi="1073741824 (1.0 GiB)"`,
		},
		{
			name: "cpus, Memory, hugepages-2Mi, devices, ephemeral-storage",
			resources: &framework.Resource{
				MilliCPU:         24000,
				Memory:           16 * 1024 * 1024 * 1024,
				EphemeralStorage: 128 * 1024 * 1024 * 1024,
				ScalarResources: map[corev1.ResourceName]int64{
					corev1.ResourceName("hugepages-2Mi"):         1 * 1024 * 1024 * 1024,
					corev1.ResourceName("example.com/netdevice"): 16,
					corev1.ResourceName("awesome.net/gpu"):       4,
				},
			},
			expected: ` cpu="24000 (24)" memory="17179869184 (16 GiB)" ephemeral-storage="137438953472 (128 GiB)" awesome.net/gpu="4" example.com/netdevice="16" hugepages-2Mi="1073741824 (1.0 GiB)"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			keysAndValues := frameworkResourceToLoggable(tt.resources)
			kvListFormat(&buf, keysAndValues...)
			got := buf.String()
			if got != tt.expected {
				t.Errorf("got=[%s] expected=[%s]", got, tt.expected)
			}
		})
	}
}

func TestCheckRequest(t *testing.T) {
	tests := []struct {
		name     string
		pod      *corev1.Pod
		nodeInfo *framework.NodeInfo
		expected []string
	}{
		{
			name: "empty",
			pod:  makePod("uid0", "namespace0", "name0"),
			nodeInfo: newNodeInfo(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-0",
				},
			}),
			expected: []string{
				`target resource requests none`,
			},
		},
		{
			name: "cpu",
			pod: withRequest(makePod("uid1", "namespace1", "name1"), corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("2"),
			}),
			nodeInfo: newNodeInfo(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-0",
				},
			}),
			expected: []string{
				`insufficient node resources" pod="namespace1/name1" podUID="uid1" node="node-0" resource="cpu"`,
			},
		},
		{
			name: "memory",
			pod: withRequest(makePod("uid2", "namespace2", "name2"), corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("8Gi"),
			}),
			nodeInfo: newNodeInfo(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-0",
				},
			}),
			expected: []string{
				`insufficient node resources" pod="namespace2/name2" podUID="uid2" node="node-0" resource="memory"`,
			},
		},
		{
			name: "devices",
			pod: withRequest(makePod("uid3", "namespace3", "name3"), corev1.ResourceList{
				corev1.ResourceName("example.com/device"): resource.MustParse("2"),
			}),
			nodeInfo: newNodeInfo(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-0",
				},
			}),
			expected: []string{
				`insufficient node resources" pod="namespace3/name3" podUID="uid3" node="node-0" resource="example.com/device"`,
			},
		},
		{
			name: "cpu and memory",
			pod: withRequest(makePod("uid4", "namespace4", "name4"), corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("16Gi"),
			}),
			nodeInfo: newNodeInfo(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-0",
				},
			}),
			expected: []string{
				`insufficient node resources" pod="namespace4/name4" podUID="uid4" node="node-0" resource="memory"`,
				`insufficient node resources" pod="namespace4/name4" podUID="uid4" node="node-0" resource="cpu"`,
			},
		},
		{
			name: "memory and devices",
			pod: withRequest(makePod("uid5", "namespace5", "name5"), corev1.ResourceList{
				corev1.ResourceMemory:                     resource.MustParse("16Gi"),
				corev1.ResourceName("example.com/device"): resource.MustParse("2"),
			}),
			nodeInfo: newNodeInfo(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-0",
				},
			}),
			expected: []string{
				`insufficient node resources" pod="namespace5/name5" podUID="uid5" node="node-0" resource="example.com/device"`,
				`insufficient node resources" pod="namespace5/name5" podUID="uid5" node="node-0" resource="memory"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runWithStderr(os.Stderr, func() error {
				checkRequest(klog.FromContext(context.TODO()), tt.pod, tt.nodeInfo)
				return nil
			})
			if err != nil {
				t.Errorf("%v", err)
			}
			for _, exp := range tt.expected {
				if !strings.Contains(got, exp) {
					t.Errorf("got=[%s] not found in expected=[%s]", got, exp)
				}
			}
		})
	}

}
func newNodeInfo(node *corev1.Node) *framework.NodeInfo {
	ninfo := framework.NewNodeInfo()
	ninfo.SetNode(node)
	return ninfo
}

func runWithStderr(oldStderr *os.File, fn func() error) (string, error) {
	defer func() {
		// don't hurt to do it twice/anyway
		os.Stderr = oldStderr
	}()
	f, err := os.CreateTemp("", "stderrtest")
	if err != nil {
		return "", err
	}
	tmpName := f.Name()
	defer os.Remove(tmpName)

	os.Stderr = f
	err = fn()
	os.Stderr = oldStderr // restore ASAP

	if err != nil {
		return "", err
	}
	err = f.Close()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(tmpName)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// taken from klog

const missingValue = "(MISSING)"

func kvListFormat(b *bytes.Buffer, keysAndValues ...interface{}) {
	for i := 0; i < len(keysAndValues); i += 2 {
		var v interface{}
		k := keysAndValues[i]
		if i+1 < len(keysAndValues) {
			v = keysAndValues[i+1]
		} else {
			v = missingValue
		}
		b.WriteByte(' ')

		switch v.(type) {
		case string, error:
			b.WriteString(fmt.Sprintf("%s=%q", k, v))
		case []byte:
			b.WriteString(fmt.Sprintf("%s=%+q", k, v))
		default:
			if _, ok := v.(fmt.Stringer); ok {
				b.WriteString(fmt.Sprintf("%s=%q", k, v))
			} else {
				b.WriteString(fmt.Sprintf("%s=%+v", k, v))
			}
		}
	}
}

func makePod(uid, namespace, name string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			UID:       types.UID(uid),
		},
	}
}

func withRequest(pod *corev1.Pod, rl corev1.ResourceList) *corev1.Pod {
	pod.Spec.Containers = []corev1.Container{
		{
			Name: "cnt",
			Resources: corev1.ResourceRequirements{
				Requests: rl,
			},
		},
	}
	return pod
}
