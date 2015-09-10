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
	"bytes"
	"reflect"
	"testing"
)

func TestGetArgs(t *testing.T) {
	tests := []struct {
		opts     Options
		expected []string
	}{
		{
			opts: Options{
				BaseURL:     "http://localhost:1234",
				URLSuffix:   "/path/to/profile",
				TimeSeconds: 5,
			},
			expected: []string{"-seconds", "5", "http://localhost:1234/path/to/profile"},
		},
		{
			opts: Options{
				BaseURL:     "http://localhost:1234/",
				URLSuffix:   "/path/to/profile",
				TimeSeconds: 5,
			},
			expected: []string{"-seconds", "5", "http://localhost:1234/path/to/profile"},
		},
		{
			opts: Options{
				BaseURL:     "http://localhost:1234/test",
				URLSuffix:   "/path/to/profile",
				TimeSeconds: 5,
			},
			expected: []string{"-seconds", "5", "http://localhost:1234/path/to/profile"},
		},
		{
			opts: Options{
				BinaryFile:  "/path/to/binaryfile",
				BaseURL:     "http://localhost:1234",
				URLSuffix:   "/profile",
				TimeSeconds: 5},
			expected: []string{"/path/to/binaryfile"},
		},
		{
			opts: Options{
				BinaryFile:  "/path/to/binaryfile",
				BinaryName:  "/path/to/binaryname",
				BaseURL:     "http://localhost:1234",
				URLSuffix:   "/profile",
				TimeSeconds: 5},
			expected: []string{"/path/to/binaryname", "/path/to/binaryfile"},
		},
	}

	for _, tt := range tests {
		got, err := getArgs(tt.opts)
		if err != nil {
			t.Errorf("failed to get pprof args: %v", err)
			continue
		}

		if !reflect.DeepEqual(tt.expected, got) {
			t.Errorf("got incorrect args for %v:\n  got %v\n want %v", tt.opts, got, tt.expected)
		}
	}
}

func TestRunPProfUnknownFlag(t *testing.T) {
	if _, err := runPProf("-unknownFlag"); err == nil {
		t.Fatalf("expected error for unknown flag")
	}
}

func TestRunPProfMissingFile(t *testing.T) {
	if _, err := runPProf("unknown-file"); err == nil {
		t.Fatalf("expected error for unknown file")
	}
}

func TestRunPProfInvalidURL(t *testing.T) {
	if _, err := runPProf("http://127.0.0.1:999/profile"); err == nil {
		t.Fatalf("expected error for unknown file")
	}
}

func TestGetPProfRaw(t *testing.T) {
	opts := Options{
		BinaryFile: "testdata/pprof.1.pb.gz",
	}
	raw, err := GetRaw(opts)
	if err != nil {
		t.Fatalf("getPProfRaw failed: %v", err)
	}

	expectedSubstrings := []string{
		"Duration: 3s",
		"Samples",
		"Locations",
		"main.fib",
	}
	for _, substr := range expectedSubstrings {
		if !bytes.Contains(raw, []byte(substr)) {
			t.Errorf("pprof raw output missing expected string: %s\ngot:\n%s", substr, raw)
		}
	}
}
