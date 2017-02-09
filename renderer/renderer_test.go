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
	"reflect"
	"testing"

	"github.com/uber/go-torch/stack"
)

func TestToFlameInput(t *testing.T) {
	profile := &stack.Profile{
		SampleNames: []string{"samples/count"},
		Samples: []*stack.Sample{
			{Funcs: []string{"func1", "func2"}, Counts: []int64{10}},
			{Funcs: []string{"func3"}, Counts: []int64{8}},
			{Funcs: []string{"func4", "func5", "func6"}, Counts: []int64{3}},
		},
	}

	expected := "func1;func2 10\nfunc3 8\nfunc4;func5;func6 3\n"

	out, err := ToFlameInput(profile, 0)
	if err != nil {
		t.Fatalf("ToFlameInput failed: %v", err)
	}

	if !reflect.DeepEqual(expected, string(out)) {
		t.Errorf("ToFlameInput failed:\n  got %s\n want %s", out, expected)
	}
}
