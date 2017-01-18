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

package pprof

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectSample(t *testing.T) {
	names := []string{
		"samples/count",
		"cpu/nanoseconds",
		"alloc_objects/count",
		"alloc_space/bytes",
		"inuse_objects/count",
		"inuse_space/bytes",
	}

	tests := []struct {
		args []string
		want int
	}{
		{
			args: nil,
			want: 0,
		},
		{
			args: []string{"-sample_index", "5"},
			want: 5,
		},
		{
			// missing argument for sample_index
			args: []string{"-sample_index"},
			want: 0,
		},
		{
			// negative sample index is out of range.
			args: []string{"-sample_index", "-1"},
			want: 0,
		},
		{
			// sample index is not a number.
			args: []string{"-sample_index", "nan"},
			want: 0,
		},
		{
			// index out of range.
			args: []string{"-sample_index", "10"},
			want: 0,
		},
		{
			args: []string{"-unknown", "options"},
			want: 0,
		},
		{
			args: []string{"-alloc_objects"},
			want: 2,
		},
		{
			args: []string{"-alloc_space"},
			want: 3,
		},
		{
			args: []string{"-inuse_objects"},
			want: 4,
		},
		{
			args: []string{"-inuse_space"},
			want: 5,
		},
	}

	for _, tt := range tests {
		got := SelectSample(tt.args, names)
		assert.Equal(t, tt.want, got, "Args: %v", tt.args)
	}

}
