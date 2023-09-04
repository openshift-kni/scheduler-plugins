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
	"testing"
)

func TestHasLogIDKey(t *testing.T) {
	tests := []struct {
		name     string
		values   []interface{}
		expected bool
	}{
		{
			name:     "nil",
			values:   nil,
			expected: false,
		},
		{
			name:     "empty",
			values:   []interface{}{},
			expected: false,
		},
		{
			name:     "missing val",
			values:   []interface{}{"logID"},
			expected: false,
		},
		{
			name:     "missing key",
			values:   []interface{}{"foobar"},
			expected: false,
		},
		{
			name:     "minimal",
			values:   []interface{}{"logID", "FOO"},
			expected: true,
		},
		{
			name:     "uneven",
			values:   []interface{}{"logID", "FOO", "BAR"},
			expected: false,
		},
		{
			name:     "multikey",
			values:   []interface{}{"logID", "AAA", "fizz", "buzz"},
			expected: true,
		},
		{
			name:     "multikey-mispell",
			values:   []interface{}{"logid", "AAA", "fizz", "buzz"},
			expected: false,
		},
		{
			name:     "multikey-mistype",
			values:   []interface{}{"logID", 12345, "fizz", "buzz"},
			expected: true,
		},
		{
			name:     "ok-not-first",
			values:   []interface{}{"foo", "bar", "logID", "AAA", "fizz", "buzz"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasLogIDKey(tt.values)
			if got != tt.expected {
				t.Errorf("values=%v got=[%v] expected=[%v]", tt.values, got, tt.expected)
			}
		})
	}
}

func TestGetLogID(t *testing.T) {
	tests := []struct {
		name       string
		values     []interface{} // always even (ensured by other API contracts)
		kvList     []interface{}
		expectedID string
		expectedOK bool
	}{
		{
			name:       "nil",
			values:     nil,
			kvList:     []interface{}{},
			expectedID: "",
			expectedOK: false,
		},
		{
			name:       "empty",
			values:     []interface{}{},
			kvList:     []interface{}{},
			expectedID: "",
			expectedOK: false,
		},
		{
			name:       "missing val",
			values:     []interface{}{},
			kvList:     []interface{}{"logID"},
			expectedID: "",
			expectedOK: false,
		},
		{
			name:       "missing key",
			values:     []interface{}{},
			kvList:     []interface{}{"foobar"},
			expectedID: "",
			expectedOK: false,
		},
		{
			name:       "minimal",
			values:     []interface{}{},
			kvList:     []interface{}{"logID", "FOO"},
			expectedID: "FOO",
			expectedOK: true,
		},
		{
			name:   "uneven",
			values: []interface{}{},
			kvList: []interface{}{"logID", "FOO", "BAR"},
			// from the limited perspective of getting logID, this is OK
			expectedID: "FOO",
			expectedOK: true,
		},
		{
			name:       "multikey",
			values:     []interface{}{},
			kvList:     []interface{}{"logID", "AAA", "fizz", "buzz"},
			expectedID: "AAA",
			expectedOK: true,
		},
		{
			name:   "ok-not-first",
			values: []interface{}{},
			// not the first to save search time, can be changed in the future
			kvList:     []interface{}{"foo", "bar", "logID", "BBB", "fizz", "buzz"},
			expectedID: "",
			expectedOK: false,
		},
		{
			name:   "missing-both",
			values: []interface{}{"alpha", "1", "beta", "2"},
			// not the first to save search time, can be changed in the future
			kvList:     []interface{}{"foo", "bar", "fizz", "buzz"},
			expectedID: "",
			expectedOK: false,
		},
		{
			name:       "values-ok-not-first",
			values:     []interface{}{"alpha", "1", "logID", "BBB", "beta", "2"},
			kvList:     []interface{}{"foo", "bar", "fizz", "buzz"},
			expectedID: "BBB",
			expectedOK: true,
		},
		{
			name:       "kvList-prevails",
			values:     []interface{}{"logID", "values", "nodeName", "localhost"},
			kvList:     []interface{}{"logID", "kvList", "foo", "bar", "fizz", "buzz"},
			expectedID: "kvList",
			expectedOK: true,
		},
		{
			name:       "kvList-mislpaleced",
			values:     []interface{}{"first", "taken", "logID", "values", "nodeName", "localhost"},
			kvList:     []interface{}{"useless", "value", "logID", "kvList", "foo", "bar", "fizz", "buzz"},
			expectedID: "values",
			expectedOK: true,
		},
		{
			name:       "kvList-mistype",
			values:     []interface{}{"alpha", "1", "beta", "2"},
			kvList:     []interface{}{"logID", 12345, "fizz", "buzz"},
			expectedID: "",
			expectedOK: false,
		},
		{
			name:       "kvList-mistype-full",
			values:     []interface{}{"alpha", "1", "beta", "2"},
			kvList:     []interface{}{123, 45, "fizz", "buzz"},
			expectedID: "",
			expectedOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := GetLogID(tt.values, tt.kvList...)
			if ok != tt.expectedOK {
				t.Fatalf("values=%v kvList=%v OK got=[%v] expected=[%v]", tt.values, tt.kvList, ok, tt.expectedOK)
			}
			if got != tt.expectedID {
				t.Errorf("values=%v kvList=%v ID got=[%v] expected=[%v]", tt.values, tt.kvList, got, tt.expectedID)
			}
		})
	}
}
