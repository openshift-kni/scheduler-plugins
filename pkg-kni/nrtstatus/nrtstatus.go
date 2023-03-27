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
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/k8stopologyawareschedwg/podfingerprint"
)

const (
	BaseDirectory = "/run/nrtstatus"
)

type NRTNodeStatus struct {
	updates chan *StatusInfo
}

var nodes *NRTNodeStatus

func init() {
	nodes = &NRTNodeStatus{}
}

type TracingStatus struct {
	nodeName string
	data     podfingerprint.Status
	updates  chan *StatusInfo
}

func MakeTracingStatus(nodeName string) TracingStatus {
	return TracingStatus{
		nodeName: nodeName,
		updates:  nodes.updates,
	}
}

func (st *TracingStatus) Start(numPods int) {
	st.data.Start(numPods)
}

func (st *TracingStatus) Add(namespace, name string) {
	st.data.Add(namespace, name)
}

func (st *TracingStatus) Sign(computed string) {
	st.data.Sign(computed)
}

func (st *TracingStatus) Check(expected string) {
	st.data.Check(expected)

	if st.updates == nil {
		return
	}
	info := StatusInfo{
		NodeName:     st.nodeName,
		Data:         st.data.Clone(),
		LastModified: time.Now(),
	}
	st.updates <- &info
}

func (st TracingStatus) Repr() string {
	return st.data.Repr()
}

type StatusInfo struct {
	NodeName     string                `json:"nodeName"`
	Data         podfingerprint.Status `json:"data"`
	LastModified time.Time             `json:"lastModified"`
}

func DumpNodeStatus(statusDir string, st *StatusInfo) error {
	dst, err := os.CreateTemp(statusDir, st.NodeName)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(dst).Encode(st); err != nil {
		return err
	}
	if err := dst.Close(); err != nil {
		return err
	}
	return os.Rename(dst.Name(), filepath.Join(statusDir, st.NodeName+".json"))
}

func LoadNodeStatus(statusDir, nodeName string) (*StatusInfo, error) {
	src, err := os.Open(filepath.Join(statusDir, nodeName+".json"))
	defer src.Close()
	if err != nil {
		return nil, err
	}
	var st StatusInfo
	err = json.NewDecoder(src).Decode(&st)
	return &st, err
}

func RunForever(ctx context.Context, expectednodes int) {
	nodes.updates = make(chan *StatusInfo, expectednodes)
	for {
		select {
		case <-ctx.Done():
			return
		case st := <-nodes.updates:
			DumpNodeStatus(BaseDirectory, st)
		}
	}
}
