// Copyright (c) 2017 Uber Technologies, Inc.
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

import "strconv"

// SelectSample returns the index of the sample to use given the
// sample names.
func SelectSample(args, names []string) int {
	selected := 0

	findName := func(needle string) {
		for i, name := range names {
			if name == needle {
				selected = i
			}
		}
	}

	for i, arg := range args {
		switch arg {
		case "-inuse_space":
			findName("inuse_space/bytes")
		case "-inuse_objects":
			findName("inuse_objects/count")
		case "-alloc_space":
			findName("alloc_space/bytes")
		case "-alloc_objects":
			findName("alloc_objects/count")
		case "-sample_index":
			// Check if there's another argument after this
			if i+1 >= len(args) {
				continue
			}

			if parsed, ok := parseSampleIndex(args[i+1], names); ok {
				selected = parsed
			}
		}
	}

	return selected
}

func parseSampleIndex(s string, names []string) (int, bool) {
	parsed, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}

	if parsed >= len(names) || parsed < 0 {
		return 0, false
	}

	return parsed, true
}
