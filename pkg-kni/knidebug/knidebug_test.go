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
	"fmt"
	"os"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

func TestFrameworkResourceToLoggable(t *testing.T) {
	tests := []struct {
		name      string
		logKey    string
		resources *framework.Resource
		expected  string
	}{
		{
			name:      "empty",
			logKey:    "",
			resources: &framework.Resource{},
			expected:  ` logKey="" cpu="0 (0)" memory="0 (0 B)"`,
		},
		{
			name:      "only logKey",
			logKey:    "TEST1",
			resources: &framework.Resource{},
			expected:  ` logKey="TEST1" cpu="0 (0)" memory="0 (0 B)"`,
		},
		{
			name:   "only CPUs",
			logKey: "TEST1",
			resources: &framework.Resource{
				MilliCPU: 16000,
			},
			expected: ` logKey="TEST1" cpu="16000 (16)" memory="0 (0 B)"`,
		},
		{
			name:   "only Memory",
			logKey: "TEST2",
			resources: &framework.Resource{
				Memory: 16 * 1024 * 1024 * 1024,
			},
			expected: ` logKey="TEST2" cpu="0 (0)" memory="17179869184 (16 GiB)"`,
		},
		{
			name:   "CPUs and Memory, no logKey",
			logKey: "",
			resources: &framework.Resource{
				MilliCPU: 24000,
				Memory:   16 * 1024 * 1024 * 1024,
			},
			expected: ` logKey="" cpu="24000 (24)" memory="17179869184 (16 GiB)"`,
		},
		{
			name:   "CPUs and Memory",
			logKey: "TEST3",
			resources: &framework.Resource{
				MilliCPU: 24000,
				Memory:   16 * 1024 * 1024 * 1024,
			},
			expected: ` logKey="TEST3" cpu="24000 (24)" memory="17179869184 (16 GiB)"`,
		},
		{
			name:   "CPUs, Memory, hugepages-2Mi",
			logKey: "TEST4",
			resources: &framework.Resource{
				MilliCPU: 24000,
				Memory:   16 * 1024 * 1024 * 1024,
				ScalarResources: map[corev1.ResourceName]int64{
					corev1.ResourceName("hugepages-2Mi"): 1 * 1024 * 1024 * 1024,
				},
			},
			expected: ` logKey="TEST4" cpu="24000 (24)" memory="17179869184 (16 GiB)" hugepages-2Mi="1073741824 (1.0 GiB)"`,
		},
		{
			name:   "CPUs, Memory, hugepages-2Mi, hugepages-1Gi",
			logKey: "TEST4",
			resources: &framework.Resource{
				MilliCPU: 24000,
				Memory:   16 * 1024 * 1024 * 1024,
				ScalarResources: map[corev1.ResourceName]int64{
					corev1.ResourceName("hugepages-2Mi"): 1 * 1024 * 1024 * 1024,
					corev1.ResourceName("hugepages-1Gi"): 2 * 1024 * 1024 * 1024,
				},
			},
			expected: ` logKey="TEST4" cpu="24000 (24)" memory="17179869184 (16 GiB)" hugepages-1Gi="2147483648 (2.0 GiB)" hugepages-2Mi="1073741824 (1.0 GiB)"`,
		},
		{
			name:   "CPUs, Memory, hugepages-2Mi, devices",
			logKey: "TEST4",
			resources: &framework.Resource{
				MilliCPU: 24000,
				Memory:   16 * 1024 * 1024 * 1024,
				ScalarResources: map[corev1.ResourceName]int64{
					corev1.ResourceName("hugepages-2Mi"):         1 * 1024 * 1024 * 1024,
					corev1.ResourceName("example.com/netdevice"): 16,
					corev1.ResourceName("awesome.net/gpu"):       4,
				},
			},
			expected: ` logKey="TEST4" cpu="24000 (24)" memory="17179869184 (16 GiB)" awesome.net/gpu="4" example.com/netdevice="16" hugepages-2Mi="1073741824 (1.0 GiB)"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			keysAndValues := frameworkResourceToLoggable(tt.logKey, tt.resources)
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
		name       string
		logKey     string
		podRequest *framework.Resource
		nodeInfo   *framework.NodeInfo
		expected   []string
	}{
		{
			name:       "empty",
			logKey:     "",
			podRequest: &framework.Resource{},
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
			name:   "cpu",
			logKey: "",
			podRequest: &framework.Resource{
				MilliCPU: 2000,
			},
			nodeInfo: newNodeInfo(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-0",
				},
			}),
			expected: []string{
				`insufficient node resources" logKey="" node="node-0" resource="CPU"`,
			},
		},
		{
			name:   "memory",
			logKey: "",
			podRequest: &framework.Resource{
				Memory: 8 * 1024 * 1024 * 1024,
			},
			nodeInfo: newNodeInfo(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-0",
				},
			}),
			expected: []string{
				`insufficient node resources" logKey="" node="node-0" resource="memory"`,
			},
		},
		{
			name:   "devices",
			logKey: "",
			podRequest: &framework.Resource{
				ScalarResources: map[corev1.ResourceName]int64{
					corev1.ResourceName("example.com/device"): 2,
				},
			},
			nodeInfo: newNodeInfo(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-0",
				},
			}),
			expected: []string{
				`insufficient node resources" logKey="" node="node-0" resource="example.com/device"`,
			},
		},
		{
			name:   "cpu and memory",
			logKey: "",
			podRequest: &framework.Resource{
				MilliCPU: 4000,
				Memory:   16 * 1024 * 1024 * 1024,
			},
			nodeInfo: newNodeInfo(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-0",
				},
			}),
			expected: []string{
				`insufficient node resources" logKey="" node="node-0" resource="memory"`,
				`insufficient node resources" logKey="" node="node-0" resource="CPU"`,
			},
		},
		{
			name:   "memory and devices",
			logKey: "",
			podRequest: &framework.Resource{
				Memory: 16 * 1024 * 1024 * 1024,
				ScalarResources: map[corev1.ResourceName]int64{
					corev1.ResourceName("example.com/device"): 2,
				},
			},
			nodeInfo: newNodeInfo(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-0",
				},
			}),
			expected: []string{
				`insufficient node resources" logKey="" node="node-0" resource="example.com/device"`,
				`insufficient node resources" logKey="" node="node-0" resource="memory"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runWithStderr(os.Stderr, func() error {
				checkRequest(0, tt.logKey, tt.podRequest, tt.nodeInfo)
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
