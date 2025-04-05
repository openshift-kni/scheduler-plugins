/*
 * Copyright 2025 Red Hat, Inc.
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

package fanout

import (
	"sync"

	"github.com/go-logr/logr"
)

type Fanout struct {
	// protects the `leaves` slice, not the logger instances
	rwlock sync.RWMutex
	leaves []logr.LogSink
}

func NewWithLeaves(leaves ...logr.LogSink) logr.LogSink {
	return &Fanout{
		leaves: leaves,
	}
}

func (fo *Fanout) Init(info logr.RuntimeInfo) {
	for _, leaf := range fo.leaves {
		leaf.Init(info)
	}
}

func (fo *Fanout) Enabled(level int) bool {
	// need to postpone the decision,
	// this causes overhead but it is unavoidable.
	return true
}

// Info dispatches the call to all the leaves
func (fo *Fanout) Info(level int, msg string, kv ...any) {
	fo.rwlock.RLock() // we don't mutate the leaves slice
	defer fo.rwlock.RUnlock()
	for idx := range fo.leaves {
		if !fo.leaves[idx].Enabled(level) {
			continue
		}
		fo.leaves[idx].Info(level, msg, kv...)
	}
}

// Error dispatches the call to all the leaves
func (fo *Fanout) Error(err error, msg string, kv ...any) {
	fo.rwlock.RLock() // we don't mutate the leaves slice
	defer fo.rwlock.RUnlock()
	for idx := range fo.leaves {
		// errors must be always propagated
		fo.leaves[idx].Error(err, msg, kv...)
	}
}

// WithValues dispatches the call to all the leaves, mutating them
func (fo *Fanout) WithValues(kv ...any) logr.LogSink {
	fo.rwlock.Lock()
	defer fo.rwlock.Unlock()
	leaves := make([]logr.LogSink, 0, len(fo.leaves))
	for _, leaf := range fo.leaves {
		leaves = append(leaves, leaf.WithValues(kv...))
	}
	return &Fanout{
		leaves: leaves,
	}
}

// WithName dispatches the call to all the leaves, mutating them
func (fo *Fanout) WithName(name string) logr.LogSink {
	fo.rwlock.Lock()
	defer fo.rwlock.Unlock()
	leaves := make([]logr.LogSink, 0, len(fo.leaves))
	for _, leaf := range fo.leaves {
		leaves = append(leaves, leaf.WithName(name))
	}
	return &Fanout{
		leaves: leaves,
	}
}

// Assert conformance to the interfaces.
var _ logr.LogSink = &Fanout{}

// TODO: implement logr.CallDepthLogSink
