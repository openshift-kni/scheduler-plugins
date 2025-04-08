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

package demuxer

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
)

var (
	ErrMissingKeyFinder         = errors.New("missing key finder")
	ErrMissingKeyValueFormatter = errors.New("missing keyvalue formatter")
)

var (
	// NameSeparator separates names for logr.WithName.
	NameSeparator = "."
)

type Demuxer struct {
	lock    sync.Mutex
	opts    Options
	name    string
	values  []any
	logBufs map[string]*logBuffer
	msgDone func(val string)
}

func (dmx *Demuxer) Register(cb func(string)) {
	dmx.lock.Lock()
	defer dmx.lock.Unlock()
	dmx.msgDone = cb
}

func (dmx *Demuxer) GetBuffer(val string) *bytes.Buffer {
	dmx.lock.Lock()
	defer dmx.lock.Unlock()
	logBuf, ok := dmx.logBufs[val]
	if !ok {
		return nil
	}
	return bytes.NewBuffer(logBuf.Bytes())
}

func (dmx *Demuxer) PopBuffer(val string) *bytes.Buffer {
	dmx.lock.Lock()
	defer dmx.lock.Unlock()
	logBuf, ok := dmx.logBufs[val]
	if !ok {
		return nil
	}
	delete(dmx.logBufs, val)
	return &logBuf.Buffer
}

func NewWithOptions(opts Options) (*Demuxer, error) {
	if opts.KeyFinder == nil {
		return nil, ErrMissingKeyFinder
	}
	if opts.KeyValueFormatter == nil {
		return nil, ErrMissingKeyValueFormatter
	}
	return &Demuxer{
		opts:    opts,
		logBufs: make(map[string]*logBuffer),
		msgDone: func(_ string) {},
	}, nil
}

// Init is not implemented and does not use any runtime info.
func (dmx *Demuxer) Init(info logr.RuntimeInfo) {
	// not implemented
}

// Enabled tests whether this Logger is enabled.
func (dmx *Demuxer) Enabled(level int) bool {
	return matchLoggerByName(dmx.name)
}

func (dmx *Demuxer) Info(level int, msg string, kv ...any) {
	dmx.writeLine(levelString(level), msg, kv...)
}

func (dmx *Demuxer) Error(err error, msg string, kv ...any) {
	dmx.writeLine(LevelError+" "+err.Error(), msg, kv...)
}

func (dmx *Demuxer) WithValues(kv ...any) logr.LogSink {
	return &Demuxer{
		name:    dmx.name,
		values:  append(dmx.values, kv...),
		opts:    dmx.opts,
		logBufs: dmx.logBufs,
		msgDone: dmx.msgDone,
	}
}

func (dmx *Demuxer) WithName(name string) logr.LogSink {
	if dmx.name != "" {
		name = dmx.name + NameSeparator + name
	}
	return &Demuxer{
		name:    name,
		values:  dmx.values,
		opts:    dmx.opts,
		logBufs: dmx.logBufs,
		msgDone: dmx.msgDone,
	}
}

func (dmx *Demuxer) writeLine(level, msg string, kv ...any) {
	var val string
	var ok bool
	if len(dmx.values) > 0 && len(dmx.values)%2 == 0 {
		val, ok = dmx.opts.KeyFinder(dmx.values)
	}
	if !ok && len(kv) > 0 && len(kv)%2 == 0 {
		val, ok = dmx.opts.KeyFinder(kv)
	}
	if !ok {
		fmt.Fprintf(os.Stderr, "XXX: dmx.writeLine KeyFinder miss: %q\n", msg)
		return
	}

	dmx.lock.Lock()
	defer dmx.lock.Unlock()

	ts := time.Now()
	logBuf, ok := dmx.logBufs[val]
	if !ok {
		logBuf = &logBuffer{}
		dmx.logBufs[val] = logBuf
	}

	logBuf.ts = ts
	logBuf.WriteLine(dmx.opts, dmx.name, level, msg, dmx.values, kv)

	dmx.msgDone(val)
}

func matchLoggerByName(name string) bool {
	name = strings.ToLower(name)
	return strings.Contains(name, "knidebug") || strings.Contains(name, "nrtcache") || strings.Contains(name, "NodeResourceTopologyMatch")
}

// Assert conformance to the interfaces.
var _ logr.LogSink = &Demuxer{}

// TODO: implement logr.CallDepthLogSink
