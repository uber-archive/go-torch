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
	"errors"
	"fmt"
)

var (
	errProfileMustHaveSamples  = errors.New("cannot create a profile with no samples")
	errProfileEmptySampleNames = errors.New("cannot have empty sample names in profile")
)

// Profile represents a parsed pprof profile.
type Profile struct {
	SampleNames []string
	Samples     []*Sample
}

// Sample represents the sample count for a specific call stack.
type Sample struct {
	// Funcs is parent first.
	Funcs  []string
	Counts []int64
}

// NewProfile returns a new profile with the specified sample names.
func NewProfile(names []string) (*Profile, error) {
	if len(names) == 0 {
		return nil, errProfileMustHaveSamples
	}
	for _, name := range names {
		if name == "" {
			return nil, errProfileEmptySampleNames
		}
	}
	return &Profile{SampleNames: names}, nil
}

// NewSample returns a new sample with a copy of the counts.
func NewSample(funcs []string, counts []int64) *Sample {
	s := &Sample{
		Funcs:  funcs,
		Counts: make([]int64, len(counts)),
	}

	// We create a copy of counts, as we may modify them in Add.
	s.Add(counts)
	return s
}

// Add combines counts with the existing counts for this sample.
func (s *Sample) Add(counts []int64) error {
	if len(s.Counts) != len(counts) {
		return fmt.Errorf("cannot add %v values to sample with %v values", len(counts), len(s.Counts))
	}

	for i := range s.Counts {
		s.Counts[i] += counts[i]
	}
	return nil
}
