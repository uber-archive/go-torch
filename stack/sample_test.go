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

package stack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProfile(t *testing.T) {
	tests := []struct {
		name    string
		names   []string
		wantErr error
	}{
		{
			name:  "valid profile",
			names: []string{"samples/count", "cpu/nanoseconds"},
		},
		{
			name:    "no samples",
			names:   nil,
			wantErr: errProfileMustHaveSamples,
		},
		{
			name:    "sample with empty name",
			names:   []string{"samples/count", "", "cpu/nanoseconds"},
			wantErr: errProfileEmptySampleNames,
		},
	}

	for _, tt := range tests {
		profile, err := NewProfile(tt.names)
		if tt.wantErr != nil {
			assert.Equal(t, tt.wantErr, err, tt.name)
			continue
		}
		assert.NoError(t, err, tt.names)
		assert.NotNil(t, profile, "Expected profile for %v", tt.names)
	}
}

func TestSample(t *testing.T) {
	s := NewSample([]string{"a", "b"}, []int64{1, 2})

	err := s.Add([]int64{3, 4})
	assert.NoError(t, err)

	err = s.Add([]int64{5})
	assert.Error(t, err, "should fail when sample counts mismatch")
}
