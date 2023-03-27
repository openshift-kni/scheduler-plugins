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

package nrtstatus

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/k8stopologyawareschedwg/podfingerprint"
)

func TestDumpNodeStatus(t *testing.T) {
	testDir, err := os.MkdirTemp("", "dumpnodest")
	if err != nil {
		t.Errorf("mkdirtemp failed: %v", err)
	}
	defer os.RemoveAll(testDir)

	fmt.Fprintf(os.Stderr, "testDir=%q\n", testDir)

	// copypasted from example: https://pkg.go.dev/time#Parse - no spacial meaning
	timeStamp, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	if err != nil {
		t.Errorf("time parse failed: %v", err)
	}

	st := StatusInfo{
		NodeName: "testNode-1",
		Data: podfingerprint.Status{
			FingerprintExpected: "PFPExpectedOrig",
			FingerprintComputed: "PFPComputedOrig",
			Pods: []podfingerprint.NamespacedName{
				{
					Namespace: "NSOrig1",
					Name:      "Name1",
				},
				{
					Namespace: "NSOrig2",
					Name:      "Name2",
				},
			},
		},
		LastModified: timeStamp,
	}

	err = DumpNodeStatus(testDir, &st)
	if err != nil {
		t.Errorf("dumpnodestatus failed: %v", err)
	}

	st2, err := LoadNodeStatus(testDir, st.NodeName)
	if err != nil {
		t.Errorf("loadnodestatus failed: %v", err)
	}

	if diff := cmp.Diff(st, *st2); diff != "" {
		t.Errorf("RTT check failed:\n%v", diff)
	}
}
