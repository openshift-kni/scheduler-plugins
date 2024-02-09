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

package logtracr

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"k8s.io/klog/v2"
)

const (
	LogTracrDirEnvVar      string = "LOGTRACR_DUMP_DIR"
	LogTracrIntervalEnvVar string = "LOGTRACR_DUMP_INTERVAL"
	LogTracrVerboseEnvVar  string = "LOGTRACR_VERBOSE"
)

type Config struct {
	Verbose       int           `json:"verbose"`
	DumpInterval  time.Duration `json:"dumpInterval"`
	DumpDirectory string        `json:"dumpDirectory"`
}

type Params struct {
	Conf        Config
	Timestamper TimeFunc
}

func Setup(ctx context.Context) (logr.Logger, bool) {
	backend := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	logh := stdr.New(backend)

	dumpDir, ok := os.LookupEnv(LogTracrDirEnvVar)
	if !ok || dumpDir == "" {
		logh.Info("disabled", "reason", "directory", "variableFound", ok, "valueGiven", dumpDir != "")
		return logh, false
	}

	val, ok := os.LookupEnv(LogTracrIntervalEnvVar)
	if !ok {
		logh.Info("disabled", "reason", "interval", "variableFound", ok)
		return logh, false
	}
	interval, err := time.ParseDuration(val)
	if err != nil {
		logh.Info("cannot parse", "interval", val, "error", err)
		return logh, false
	}
	if interval == 0 {
		return logh, false
	}

	verb, ok := os.LookupEnv(LogTracrVerboseEnvVar)
	if !ok {
		logh.Info("disabled", "reason", "verbose", "variableFound", ok)
		return logh, false
	}
	verbose, err := strconv.Atoi(verb)
	if err != nil {
		logh.Info("cannot parse", "verbose", verb, "error", err)
		return logh, false
	}

	return SetupWithParams(ctx, backend, Params{
		Conf: Config{
			Verbose:       verbose,
			DumpInterval:  interval,
			DumpDirectory: dumpDir,
		},
		Timestamper: time.Now,
	}), true
}

func SetupWithParams(ctx context.Context, backend *log.Logger, params Params) logr.Logger {
	sink := stdr.New(backend)
	sink.Info("starting", "configuration", toJSON(params.Conf))

	traces := NewLogCache(params.Timestamper)

	klog.SetLogger(New(backend, traces, params.Conf.Verbose, stdr.Options{}))
	go RunForever(ctx, sink, params.Conf.DumpInterval, params.Conf.DumpDirectory, traces)

	return sink
}

func toJSON(obj interface{}) string {
	data, err := json.Marshal(obj)
	if err != nil {
		return "<ERROR>"
	}
	return string(data)
}
