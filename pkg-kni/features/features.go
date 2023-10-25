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

package features

import (
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/component-base/featuregate"
	"k8s.io/klog/v2"
)

func Desired() map[string]bool {
	return map[string]bool{
		"UnauthenticatedHTTP2DOSMitigation": true,
	}
}

func Names() []featuregate.Feature {
	gates := Desired()
	feats := make([]featuregate.Feature, 0, len(gates))
	for key := range gates {
		feats = append(feats, featuregate.Feature(key))
	}
	return feats
}

func LogState(tag string, verbose int, keys []featuregate.Feature) {
	header := tag + " feature gate"
	for _, key := range keys {
		klog.V(klog.Level(verbose)).InfoS(header, "name", key, "enabled", utilfeature.DefaultFeatureGate.Enabled(key))
	}
}
