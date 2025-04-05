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

package logtracr

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"

	"sigs.k8s.io/scheduler-plugins/pkg-kni/logtracr/demuxer"
	"sigs.k8s.io/scheduler-plugins/pkg-kni/logtracr/flusher"
)

type Config struct {
	LogKey      string
	FlushPeriod time.Duration
	Flusher     flusher.Config
}

type AsyncFlushFunc func()

type Control struct {
	AsyncFlush AsyncFlushFunc
}

func NewTracrWithConfig(ctx context.Context, cfg Config) (logr.LogSink, Control, error) {
	dmx, err := demuxer.NewWithOptions(demuxer.Options{
		KeyFinder:         demuxer.GenericKeyFinder(cfg.LogKey),
		KeyValueFormatter: demuxer.DefaultKeyValueFormatter,
	})
	if err != nil {
		return logr.Discard().GetSink(), Control{}, err
	}

	stdLog := log.New(os.Stderr, "", log.Lshortfile)
	fl := flusher.NewWithLogger(stdr.New(stdLog), cfg.Flusher, dmx)

	dmx.Register(fl.MessageDone)

	// TODO close on cancel?
	flushReqCh := make(chan struct{})

	go flushLoop(ctx, flushReqCh, cfg.FlushPeriod, fl)

	return dmx, Control{
		AsyncFlush: func() {
			flushReqCh <- struct{}{}
		},
	}, nil
}

func flushLoop(ctx context.Context, flushReqCh chan struct{}, flushPeriod time.Duration, fl *flusher.Flusher) {
	ticker := time.NewTicker(flushPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fl.FlushAll()
			return
		case <-flushReqCh:
			fl.FlushAll()
		case ts := <-ticker.C:
			fl.Flush(ts)
		}
	}
}
