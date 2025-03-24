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
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
)

const (
	LevelError = "ERROR"
	LevelInfo  = "INFO"
	LevelV     = "V[%d]"
)

var (
	// NameSeparator separates names for logr.WithName.
	NameSeparator = "."
)

type logBuffer struct {
	bytes.Buffer
	ts time.Time // last updated
}

type Demuxer struct {
	// protects the `leaves` slice, not the logger instances
	lock    sync.Mutex
	opts    Options
	name    string
	values  []any
	logBufs map[string]*logBuffer
	msgDone func(val string)
}

type Options struct {
	// KeyFinder can assume len(kv) >= 2 && len(kv)%2 == 0
	KeyFinder func(kv []any) (string, bool)
	// KeyValueFormatter can assume len(kv) > 0
	KeyValueFormatter func(kv []any) string
}

func DefaultKeyValueFormatter(kv ...any) string {
	var sb strings.Builder
	if s, ok := toString(kv[0]); ok {
		sb.WriteString(s)
	}
	for _, x := range kv[1:] {
		if s, ok := toString(x); ok {
			sb.WriteString(" ")
			sb.WriteString(s)
		}
	}
	return sb.String()
}

func DefaultKeyFinder(key string) func(kv []any) (string, bool) {
	return func(kv []any) (string, bool) {
		if s, ok := toString(kv[0]); !ok || s != key {
			return "", false
		}
		return toString(kv[1])
	}
}

func GenericKeyFinder(key string) func(kv []any) (string, bool) {
	return func(kv []any) (string, bool) {
		for idx := 0; idx < len(kv); idx += 2 {
			if s, ok := toString(kv[idx]); ok && s == key {
				return toString(kv[idx+1])
			}
		}
		return "", false
	}
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

func NewWithOptions(opts *Options) *Demuxer {
	return &Demuxer{
		opts:    *opts,
		msgDone: func(_ string) {},
	}
}

// Init is not implemented and does not use any runtime info.
func (dmx *Demuxer) Init(info logr.RuntimeInfo) {
	// not implemented
}

// Enabled tests whether this Logger is enabled.
func (dmx *Demuxer) Enabled(level int) bool {
	return true // hardcoded, we filter using a different way
}

func (dmx *Demuxer) Info(level int, msg string, kv ...any) {
	dmx.writeLine(dmx.levelString(level), msg, kv...)
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
	}
}

func (dmx *Demuxer) writeLine(level, msg string, kv ...any) {
	if len(kv) < 2 || len(kv)%2 != 0 {
		return
	}
	if dmx.opts.KeyFinder == nil {
		return
	}

	val, ok := dmx.opts.KeyFinder(kv)
	if !ok {
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
	logBuf.WriteString(level)
	if dmx.name != "" {
		logBuf.WriteString(" ")
		logBuf.WriteString(dmx.name)
	}
	if msg != "" {
		logBuf.WriteString(" ")
		logBuf.WriteString(msg)
	}
	if len(dmx.values) > 0 {
		logBuf.WriteString(" ")
		logBuf.WriteString(dmx.opts.KeyValueFormatter(dmx.values))
	}
	logBuf.WriteString(" ")
	logBuf.WriteString(dmx.opts.KeyValueFormatter(kv))
	logBuf.WriteString("\n")

	dmx.msgDone(val)
}

func (dmx *Demuxer) levelString(level int) string {
	if level > 0 {
		return fmt.Sprintf(LevelV, level)
	}
	return LevelInfo
}

func toString(v any) (string, bool) {
	if s, ok := v.(string); ok {
		return s, true
	}
	if st, ok := v.(fmt.Stringer); ok {
		return st.String(), true
	}
	return "<unrep>", false
}

// Assert conformance to the interfaces.
var _ logr.LogSink = &Demuxer{}

// TODO: implement logr.CallDepthLogSink
