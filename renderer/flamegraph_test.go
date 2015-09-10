// Copyright (c) 2015 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package renderer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindInPatch(t *testing.T) {
	const realCmd1 = "ls"
	const realCmd2 = "cat"
	const fakeCmd1 = "should-not-find-this"
	const fakeCmd2 = "not-going-to-exist"

	tests := []struct {
		paths    []string
		expected string
	}{
		{
			paths: []string{},
		},
		{
			paths:    []string{realCmd1},
			expected: realCmd1,
		},
		{
			paths:    []string{fakeCmd1, realCmd1},
			expected: realCmd1,
		},
		{
			paths:    []string{fakeCmd1, realCmd1, fakeCmd2, realCmd2},
			expected: realCmd1,
		},
	}

	for _, tt := range tests {
		got := findInPath(tt.paths)
		var gotFile string
		if got != "" {
			gotFile = filepath.Base(got)
		}
		if gotFile != tt.expected {
			t.Errorf("findInPaths(%v) got %v, want %v", tt.paths, gotFile, tt.expected)
		}

		// Verify that the returned path exists.
		if got != "" {
			_, err := os.Stat(got)
			if err != nil {
				t.Errorf("returned path %v failed to stat: %v", got, err)

			}
		}
	}
}
