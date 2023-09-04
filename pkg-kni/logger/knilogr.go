/*
 * Copyright 2019 The logr Authors.
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
 *
 * Derived from https://github.com/go-logr/stdr/blob/v1.2.2/stdr.go
 */

package logger

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"

	"github.com/ffromani/formatr"
)

type knilogger struct {
	formatr.Formatter
	std      stdr.StdLogger
	cache    *LogCache
	logID    string
	hasLogID bool
	verbose  int
}

func New(std stdr.StdLogger, lc *LogCache, verbose int, opts stdr.Options) logr.Logger {
	sl := &knilogger{
		Formatter: formatr.NewFormatterKlog(formatr.Options{
			LogCaller: formatr.MessageClass(opts.LogCaller),
		}),
		std:     std,
		cache:   lc,
		verbose: verbose,
	}

	// For skipping our own logger.Info/Error.
	sl.Formatter.AddCallDepth(1 + opts.Depth)

	return logr.New(sl)
}

func (l knilogger) Enabled(level int) bool {
	return true
}

func (l knilogger) Info(level int, msg string, kvList ...interface{}) {
	prefix, args := l.FormatInfo(level, msg, kvList)
	if prefix != "" {
		args = prefix + ": " + args
	}
	l.storeLog(args, kvList)
	// because we can be here because either we have enough verbosiness
	// OR because stored logID. So we must redo this check.
	if l.verbose < level {
		return
	}
	_ = l.std.Output(l.Formatter.GetDepth()+1, args)
}

func (l knilogger) Error(err error, msg string, kvList ...interface{}) {
	prefix, args := l.FormatError(err, msg, kvList)
	if prefix != "" {
		args = prefix + ": " + args
	}
	l.storeLog(args, kvList)
	_ = l.std.Output(l.Formatter.GetDepth()+1, args)
}

func (l knilogger) storeLog(args string, kvList []interface{}) {
	if l.cache == nil {
		return
	}
	logID, ok := l.logID, l.hasLogID
	if !ok {
		logID, ok = StartsWithLogID(kvList...)
	}
	if !ok {
		return
	}
	// less precise, but should still be good enough
	ts := time.Now().Format(time.StampMicro)
	// ignore error
	l.cache.Put(logID, ts, " ", args, "\n")
}

func (l knilogger) WithName(name string) logr.LogSink {
	l.Formatter.AddName(name)
	return &l
}

func (l knilogger) WithValues(kvList ...interface{}) logr.LogSink {
	l.logID, l.hasLogID = FindLogID(kvList)
	l.Formatter.AddValues(kvList)
	return &l
}

func (l knilogger) WithCallDepth(depth int) logr.LogSink {
	l.Formatter.AddCallDepth(depth)
	return &l
}

var _ logr.LogSink = &knilogger{}
var _ logr.CallDepthLogSink = &knilogger{}
