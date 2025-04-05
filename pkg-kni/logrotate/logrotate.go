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

package logrotate

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

const (
	FileExtension string = ".log"
)

type Directory string

type Result struct {
	TotalKept       int
	TotalRotated    int
	TotalFreedSpace int64
	TotalUsedSpace  int64
	Errors          []error
}

func (dir Directory) ByAge(now time.Time, maxAge time.Duration) Result {
	ret := Result{}
	entries, err := os.ReadDir(string(dir))
	if err != nil {
		ret.Errors = append(ret.Errors, err)
		return ret
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fName := entry.Name()
		if !strings.HasSuffix(fName, FileExtension) {
			continue
		}

		fInfo, err := entry.Info()
		if err != nil {
			ret.Errors = append(ret.Errors, err)
			continue
		}

		if now.Sub(fInfo.ModTime()) < maxAge {
			ret.TotalKept += 1
			ret.TotalUsedSpace += fInfo.Size()
			continue
		}

		fullPath := filepath.Join(string(dir), fName)
		err = os.Remove(fullPath)
		if err != nil {
			ret.Errors = append(ret.Errors, err)
			continue
		}

		ret.TotalRotated += 1
		ret.TotalFreedSpace += fInfo.Size()
	}

	return ret
}

func (dir Directory) LoopByAge(ctx context.Context, lh logr.Logger, maxAge, period time.Duration) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case ts := <-ticker.C:
			res := dir.ByAge(ts, maxAge)
			for _, err := range res.Errors {
				lh.Error(err, "rotating directory", "method", "age", "ts", ts)
			}
			lh.V(4).Info("rotation summary", "kept", res.TotalKept, "rotated", res.TotalRotated, "storageUsed", res.TotalUsedSpace, "storageFreed", res.TotalFreedSpace)
		}
	}
}
