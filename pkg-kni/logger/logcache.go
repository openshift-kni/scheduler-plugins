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

package logger

import (
	"bytes"
	"sync"
	"time"
)

type LogNode struct {
	logID      string
	data       bytes.Buffer
	lastUpdate time.Time
}

func (ln *LogNode) IsExpired(now time.Time, delta time.Duration) bool {
	return now.Sub(ln.lastUpdate) >= delta
}

type TimeFunc func() time.Time

type LogCache struct {
	mutex sync.Mutex
	// map logID -> data
	nodes    map[string]*LogNode
	timeFunc TimeFunc
}

func NewLogCache(timeFunc TimeFunc) *LogCache {
	return &LogCache{
		nodes:    make(map[string]*LogNode),
		timeFunc: timeFunc,
	}
}

func (lc *LogCache) Put(logID string, args ...string) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()
	buf := lc.bufferFor(logID)
	for _, arg := range args {
		_, err := buf.WriteString(arg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (lc *LogCache) PopExpired(now time.Time, delta time.Duration) []*LogNode {
	ret := []*LogNode{}
	lc.mutex.Lock()
	defer lc.mutex.Unlock()
	for logID, LogNode := range lc.nodes {
		if !LogNode.IsExpired(now, delta) {
			continue
		}
		ret = append(ret, LogNode)
		delete(lc.nodes, logID)
	}
	return ret
}

// Get is (mostly) meant for testing purposes
func (lc *LogCache) Get(logID string) (string, bool) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()
	node, ok := lc.nodes[logID]
	if !ok {
		return "", false
	}
	return node.data.String(), true
}

// Len is (mostly) meant for testing purposes
func (lc *LogCache) Len() int {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()
	return len(lc.nodes)
}

func (lc *LogCache) bufferFor(logID string) *bytes.Buffer {
	node, ok := lc.nodes[logID]
	if !ok {
		node = &LogNode{
			logID: logID,
		}
		lc.nodes[logID] = node
	}
	node.lastUpdate = lc.timeFunc()
	return &node.data
}
