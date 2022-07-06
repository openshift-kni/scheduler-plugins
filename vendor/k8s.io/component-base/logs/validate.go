/*
Copyright 2021 The Kubernetes Authors.

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

package logs

import (
<<<<<<< HEAD
	"flag"
	"fmt"
	"strings"

	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/component-base/config"
=======
	"fmt"
	"math"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/config"
	"k8s.io/component-base/logs/registry"
>>>>>>> upstream/master
)

func ValidateLoggingConfiguration(c *config.LoggingConfiguration, fldPath *field.Path) field.ErrorList {
	errs := field.ErrorList{}
	if c.Format != DefaultLogFormat {
<<<<<<< HEAD
		allFlags := UnsupportedLoggingFlags(hyphensToUnderscores)
		for _, fname := range allFlags {
			if flagIsSet(fname, hyphensToUnderscores) {
				errs = append(errs, field.Invalid(fldPath.Child("format"), c.Format, fmt.Sprintf("Non-default format doesn't honor flag: %s", fname)))
			}
		}
	}
	if _, err := LogRegistry.Get(c.Format); err != nil {
		errs = append(errs, field.Invalid(fldPath.Child("format"), c.Format, "Unsupported log format"))
	}
	return errs
}

// hyphensToUnderscores replaces hyphens with underscores
// we should always use underscores instead of hyphens when validate flags
func hyphensToUnderscores(s string) string {
	return strings.Replace(s, "-", "_", -1)
}

func flagIsSet(name string, normalizeFunc func(name string) string) bool {
	f := flag.Lookup(name)
	if f != nil {
		return f.DefValue != f.Value.String()
	}
	if normalizeFunc != nil {
		f = flag.Lookup(normalizeFunc(name))
		if f != nil {
			return f.DefValue != f.Value.String()
		}
	}
	pf := pflag.Lookup(name)
	if pf != nil {
		return pf.DefValue != pf.Value.String()
	}
	panic("failed to lookup unsupported log flag")
=======
		// WordSepNormalizeFunc is just a guess. Commands should use it,
		// but we cannot know for sure.
		allFlags := UnsupportedLoggingFlags(cliflag.WordSepNormalizeFunc)
		for _, f := range allFlags {
			if f.DefValue != f.Value.String() {
				errs = append(errs, field.Invalid(fldPath.Child("format"), c.Format, fmt.Sprintf("Non-default format doesn't honor flag: %s", f.Name)))
			}
		}
	}
	_, err := registry.LogRegistry.Get(c.Format)
	if err != nil {
		errs = append(errs, field.Invalid(fldPath.Child("format"), c.Format, "Unsupported log format"))
	}

	// The type in our struct is uint32, but klog only accepts positive int32.
	if c.Verbosity > math.MaxInt32 {
		errs = append(errs, field.Invalid(fldPath.Child("verbosity"), c.Verbosity, fmt.Sprintf("Must be <= %d", math.MaxInt32)))
	}
	vmoduleFldPath := fldPath.Child("vmodule")
	if len(c.VModule) > 0 && c.Format != "" && c.Format != "text" {
		errs = append(errs, field.Forbidden(vmoduleFldPath, "Only supported for text log format"))
	}
	for i, item := range c.VModule {
		if item.FilePattern == "" {
			errs = append(errs, field.Required(vmoduleFldPath.Index(i), "File pattern must not be empty"))
		}
		if strings.ContainsAny(item.FilePattern, "=,") {
			errs = append(errs, field.Invalid(vmoduleFldPath.Index(i), item.FilePattern, "File pattern must not contain equal sign or comma"))
		}
		if item.Verbosity > math.MaxInt32 {
			errs = append(errs, field.Invalid(vmoduleFldPath.Index(i), item.Verbosity, fmt.Sprintf("Must be <= %d", math.MaxInt32)))
		}
	}

	// Currently nothing to validate for c.Options.

	return errs
>>>>>>> upstream/master
}
