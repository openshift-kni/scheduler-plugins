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

package pfpstatus

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/k8stopologyawareschedwg/podfingerprint"
)

const (
	BaseDirectory = "/run/pfpstatus"
)

type StatusInfo struct {
	podfingerprint.Status
	LastWrite time.Time `json:"lastWrite"`
	SeqNo     int64     `json:"seqNo"`
}

func RunForever(ctx context.Context, logger logr.Logger, baseDirectory string, updates <-chan podfingerprint.Status) {
	// let's try to keep the amount of code we do in init() at minimum.
	// This may happen if the container didn't have the directory mounted
	discard := !existsBaseDirectory(baseDirectory)
	if discard {
		logger.Info("base directory not found, will discard everything", "baseDirectory", baseDirectory)
	}

	var seqNo int64 // 63 bits ought to be enough for anybody
	logger.V(4).Info("status update loop started")
	defer logger.V(4).Info("status update loop finished")
	for {
		select {
		case <-ctx.Done():
			return
		// always keep dequeueing messages to not block the sender
		case st := <-updates:
			if discard {
				return
			}
			seqNo += 1
			// intentionally ignore errors, must keep going
			sti := StatusInfo{
				Status:    st,
				LastWrite: time.Now(),
				SeqNo:     seqNo,
			}
			DumpNodeStatus(BaseDirectory, sti)
		}
	}
}

func nodeNameToFileName(name string) string {
	return strings.ReplaceAll(name, ".", "_")
}

func DumpNodeStatus(statusDir string, st StatusInfo) error {
	nodeName := nodeNameToFileName(st.NodeName)

	dst, err := os.CreateTemp(statusDir, nodeName)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(dst).Encode(st); err != nil {
		dst.Close() // swallow error because we want to bubble up the encoding error
		return err
	}
	if err := dst.Close(); err != nil {
		return err
	}
	return os.Rename(dst.Name(), filepath.Join(statusDir, nodeName+".json"))
}

func LoadNodeStatus(statusDir, nodeName string) (StatusInfo, error) {
	nodeName = nodeNameToFileName(nodeName)

	src, err := os.Open(filepath.Join(statusDir, nodeName+".json"))
	defer src.Close()
	if err != nil {
		return StatusInfo{}, err
	}
	var st StatusInfo
	err = json.NewDecoder(src).Decode(&st)
	return st, err
}

func existsBaseDirectory(baseDir string) bool {
	info, err := os.Stat(baseDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}
