// Copyright 2022 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pkg

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_isUnderWD(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		expected error
	}{
		{
			name:     "some valid path",
			path:     "./some/valid/path",
			expected: nil,
		},
		{
			name:     "some valid path",
			path:     "../pkg/some/valid/path",
			expected: nil,
		},
		{
			name:     "parent invalid path",
			path:     "../invalid/path",
			expected: errorInvalidDirectory,
		},
		{
			name:     "some invalid fullpath",
			path:     "/some/invalid/fullpath",
			expected: errorInvalidDirectory,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := isUnderWD(tt.path)
			if !errCmp(err, tt.expected) {
				t.Errorf(cmp.Diff(err, tt.expected))
			}
		})
	}
}
