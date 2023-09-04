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
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"k8s.io/component-base/featuregate"
	logsapi "k8s.io/component-base/logs/api/v1"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

const (
	KNILogFormat = "kni"

	KNILoggerVerboseEnvVar  = "KNI_LOGGER_VERBOSE"
	KNILoggerIntervalEnvVar = "KNI_LOGGER_DUMP_INTERVAL"
	KNILoggerDirEnvVar      = "KNI_LOGGER_DUMP_DIR"
)

func Setup(ctx context.Context) logr.Logger {
	backend := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	logh := stdr.New(backend)

	dumpDir, ok := os.LookupEnv(KNILoggerDirEnvVar)
	if !ok || dumpDir == "" {
		logh.Info("disabled", "reason", "directory", "variableFound", ok, "valueGiven", dumpDir != "")
		return logh
	}

	val, ok := os.LookupEnv(KNILoggerIntervalEnvVar)
	if !ok {
		logh.Info("disabled", "reason", "interval", "variableFound", ok)
		return logh
	}
	interval, err := time.ParseDuration(val)
	if err != nil {
		logh.Info("cannot parse", "interval", val, "error", err)
		return logh
	}
	if interval == 0 {
		return logh
	}

	verb, ok := os.LookupEnv(KNILoggerVerboseEnvVar)
	if !ok {
		logh.Info("disabled", "reason", "verbose", "variableFound", ok)
		return logh
	}
	verbose, err := strconv.Atoi(verb)
	if err != nil {
		logh.Info("cannot parse", "verbose", verb, "error", err)
		return logh
	}

	logh.Info("setting up done", "verbose", verbose, "interval", interval, "directory", dumpDir)

	cache := NewLogCache(time.Now)

	factory := Factory{
		backend: backend,
		cache:   cache,
		verbose: verbose,
	}
	err = logsapi.RegisterLogFormat(KNILogFormat, factory, logsapi.LoggingBetaOptions)
	if err != nil {
		logh.Error(err, "error registering logformat %q", KNILogFormat)
		return logh
	}

	go RunForever(ctx, logh, interval, dumpDir, cache)

	logh.Info("KNI logger ready and running")

	return logh
}

// Factory produces JSON logger instances.
type Factory struct {
	backend *log.Logger
	cache   *LogCache
	verbose int
}

var _ logsapi.LogFormatFactory = Factory{}

func (f Factory) Feature() featuregate.Feature {
	return logsapi.LoggingBetaOptions
}

func (f Factory) Create(c logsapi.LoggingConfiguration) (logr.Logger, func()) {
	return New(f.backend, f.cache, f.verbose, stdr.Options{}), func() {}
}
