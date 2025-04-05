/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package options

import (
	"time"

	"github.com/spf13/pflag"
)

type TracrOptions struct {
	BaseDirectory string        `json:"baseDirectory"`
	FlushPeriod   time.Duration `json:"flushPeriod"`
	FlushMaxAge   time.Duration `json:"flushMaxAge"`
	RotatePeriod  time.Duration `json:"rotatePeriod"`
	RotateMaxAge  time.Duration `json:"rotateMaxAge"`
}

func NewTracrOptions() *TracrOptions {
	return &TracrOptions{
		FlushPeriod:  2 * time.Second,
		FlushMaxAge:  1 * time.Second,
		RotatePeriod: 5 * time.Minute,
		RotateMaxAge: 12 * time.Hour,
	}
}

// AddFlags adds flags for the tracr options.
func (o *TracrOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}

	fs.StringVar(&o.BaseDirectory, "log-tracr-directory", o.BaseDirectory, "log tracr base directory. If empty, disable the log tracr.")
	fs.DurationVar(&o.FlushPeriod, "log-tracr-flush-period", o.FlushPeriod, "log tracr flush period.")
	fs.DurationVar(&o.FlushMaxAge, "log-tracr-flush-max-age", o.FlushMaxAge, "log tracr per-key-block flush max age.")
	fs.DurationVar(&o.RotatePeriod, "log-tracr-rotate-period", o.RotatePeriod, "log tracr rotation period.")
	fs.DurationVar(&o.RotateMaxAge, "log-tracr-rotate-max-age", o.RotateMaxAge, "log tracr per-key-block rotation max age.")
}
