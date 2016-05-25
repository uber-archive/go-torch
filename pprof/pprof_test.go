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
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGetArgs(t *testing.T) {
	four := 4
	tests := []struct {
		opts      Options
		remaining []string
		expected  []string
		wantErr   bool
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
				BaseURL:   "http://localhost:1234/",
				URLSuffix: "/path/to/profile",
				TimeAlias: &four,
			},
			expected: []string{"-seconds", "4", "http://localhost:1234/path/to/profile"},
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
				TimeSeconds: 5,
			},
			expected: []string{"/path/to/binaryfile"},
		},
		{
			opts: Options{
				BinaryFile:  "/path/to/binaryfile",
				BinaryName:  "/path/to/binaryname",
				BaseURL:     "http://localhost:1234",
				URLSuffix:   "/profile",
				TimeSeconds: 5,
			},
			expected: []string{"/path/to/binaryname", "/path/to/binaryfile"},
		},
		{
			opts: Options{
				BinaryFile: "/path/to/binaryfile",
				ExtraArgs:  []string{"-arg1", "-arg2"},
			},
			expected: []string{"-arg1", "-arg2", "/path/to/binaryfile"},
		},
		{
			opts: Options{
				BaseURL:     "%-0", // this makes url.Parse fail.
				URLSuffix:   "/profile",
				TimeSeconds: 5,
			},
			wantErr: true,
		},
		{
			remaining: []string{"binary", "input"},
			expected:  []string{"binary", "input"},
		},
		{
			opts: Options{
				TimeSeconds: 5,
			},
			remaining: []string{"binary", "input"},
			expected:  []string{"-seconds", "5", "binary", "input"},
		},
		{
			opts: Options{
				TimeSeconds: 5,
				// All other fields are ignored when remaining is specified.
				BinaryFile: "/path/to/binaryfile",
				BinaryName: "/path/to/binaryname",
				URLSuffix:  "/ignored",
			},
			remaining: []string{"binary", "input"},
			expected:  []string{"-seconds", "5", "binary", "input"},
		},
	}

	for _, tt := range tests {
		got, err := getArgs(tt.opts, tt.remaining)
		if (err != nil) != tt.wantErr {
			t.Errorf("wantErr %v got error: %v", tt.wantErr, err)
			continue
		}
		if err != nil {
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
	server := httptest.NewServer(http.HandlerFunc(http.NotFound))
	defer server.Close()

	if _, err := runPProf(server.URL); err == nil {
		t.Fatalf("expected error for unknown file")
	}
}

func TestGetPProfRawBadURL(t *testing.T) {
	opts := Options{
		BaseURL: "%-0",
	}
	if _, err := GetRaw(opts, nil); err == nil {
		t.Error("expected bad BaseURL to fail")
	}
}

func TestGetPProfRawSuccess(t *testing.T) {
	opts := Options{
		BinaryFile: "testdata/pprof.1.pb.gz",
	}
	raw, err := GetRaw(opts, nil)
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
