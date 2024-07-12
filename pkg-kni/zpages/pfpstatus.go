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

package zpages

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"

	"github.com/k8stopologyawareschedwg/podfingerprint"
)

type StatusInfo struct {
	podfingerprint.Status
	LastWrite time.Time `json:"lastWrite"`
	SeqNo     int64     `json:"seqNo"`
}

type PFPStatus struct {
	logh logr.Logger
	lock sync.RWMutex
	// nodeName -> statusInfo
	data map[string]StatusInfo
}

func NewPFPStatus(logh logr.Logger) *PFPStatus {
	st := &PFPStatus{
		data: make(map[string]StatusInfo),
		logh: logh,
	}
	ch := make(chan podfingerprint.Status)
	podfingerprint.SetCompletionSink(ch)
	go st.CollectForever(context.Background(), ch)
	return st
}

func (pfps *PFPStatus) CollectForever(ctx context.Context, updates <-chan podfingerprint.Status) {
	var seqNo int64 // 63 bits ought to be enough for anybody
	pfps.logh.V(4).Info("status collection loop started")
	defer pfps.logh.V(4).Info("status collection loop finished")
	for {
		select {
		case <-ctx.Done():
			return
		// always keep dequeueing messages to not block the sender
		case st := <-updates:
			seqNo += 1
			// intentionally ignore errors, must keep going
			sti := StatusInfo{
				Status:    st,
				LastWrite: time.Now(),
				SeqNo:     seqNo,
			}
			pfps.lock.Lock()
			pfps.data[sti.NodeName] = sti
			pfps.lock.Unlock()
		}
	}
}

func (pfps *PFPStatus) List() []StatusInfo {
	nodes := make([]StatusInfo, 0, 2) // always return non-nil
	pfps.lock.RLock()
	for _, info := range pfps.data {
		nodes = append(nodes, info)
	}
	pfps.lock.RUnlock()
	return nodes
}

func (pfps *PFPStatus) Get(key string) (StatusInfo, bool) {
	pfps.lock.RLock()
	defer pfps.lock.RUnlock()
	val, ok := pfps.data[key]
	return val, ok
}
