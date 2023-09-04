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

const (
	logIDKey = "logID"
)

func StartsWithLogID(kvList ...interface{}) (string, bool) {
	if len(kvList) < 2 {
		return "", false
	}
	return isLogIDPair(kvList[0], kvList[1])
}

func FindLogID(values []interface{}) (string, bool) {
	if len(values) < 2 || len(values)%2 != 0 {
		return "", false // should never happen
	}
	for i := 0; i < len(values); i += 2 {
		if vs, ok := isLogIDPair(values[i], values[i+1]); ok {
			return vs, ok
		}
	}
	return "", false
}

func GetLogID(values []interface{}, kvList ...interface{}) (string, bool) {
	// quick check because we look at most at 2 values
	if len(kvList) >= 2 {
		if vs, ok := isLogIDPair(kvList[0], kvList[1]); ok {
			return vs, ok
		}
	}
	if len(values) < 2 || len(values)%2 != 0 {
		return "", false // should never happen
	}
	for i := 0; i < len(values); i += 2 {
		if vs, ok := isLogIDPair(values[i], values[i+1]); ok {
			return vs, ok
		}
	}
	return "", false
}

func HasLogIDKey(vals []interface{}) bool {
	if len(vals) < 2 || len(vals)%2 != 0 {
		return false
	}
	for i := 0; i < len(vals); i += 2 {
		if isLogIDKey(vals[i]) {
			return true
		}
	}
	return false
}

func isLogIDPair(key, val interface{}) (string, bool) {
	if !isLogIDKey(key) {
		return "", false
	}
	vs, ok := val.(string)
	return vs, ok
}

func isLogIDKey(val interface{}) bool {
	v, ok := val.(string)
	if !ok {
		return false
	}
	return v == logIDKey
}
