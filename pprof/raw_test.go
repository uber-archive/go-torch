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
	"io/ioutil"
	"reflect"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	rawBytes, err := ioutil.ReadFile("testdata/pprof.raw.txt")
	if err != nil {
		t.Fatalf("Failed to read testdata/pprof.raw.txt: %v", err)
	}

	parser := newRawParser()
	if err := parser.parse(rawBytes); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// line 7 - 249 are stack records in the test file.
	const expectedNumRecords = 242
	if len(parser.records) != expectedNumRecords {
		t.Errorf("Failed to parse all records, got %v records, expected %v",
			len(parser.records), expectedNumRecords)
	}
	expectedRecords := map[int]*stackRecord{
		0:  &stackRecord{1, time.Duration(10000000), []funcID{1, 2, 2, 2, 3, 3, 2, 2, 3, 3, 2, 2, 2, 3, 3, 3, 2, 3, 2, 3, 2, 2, 3, 2, 2, 3, 4, 5, 6}},
		18: &stackRecord{1, time.Duration(10000000), []funcID{14, 2, 2, 3, 2, 2, 3, 2, 2, 3, 3, 3, 2, 2, 2, 3, 3, 2, 3, 3, 3, 3, 3, 2, 4, 5, 6}},
		45: &stackRecord{12, time.Duration(120000000), []funcID{23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34}},
	}
	for recordNum, expected := range expectedRecords {
		if got := parser.records[recordNum]; !reflect.DeepEqual(got, expected) {
			t.Errorf("Unexpected record for %v:\n  got %#v\n want %#v", recordNum, got, expected)
		}
	}

	// line 250 - 290 are locations (or funcID mappings)
	const expectedFuncIDs = 41
	if len(parser.funcName) != expectedFuncIDs {
		t.Errorf("Failed to parse func ID mappings, got %v records, expected %v",
			len(parser.funcName), expectedFuncIDs)
	}
	knownMappings := map[funcID]string{
		1:  "main.fib",
		20: "main.fib",
		34: "runtime.morestack",
	}
	for funcID, expected := range knownMappings {
		if got := parser.funcName[funcID]; got != expected {
			t.Errorf("Unexpected mapping for %v: got %v, want %v", funcID, got, expected)
		}
	}
}

func TestSplitBySpace(t *testing.T) {
	tests := []struct {
		s        string
		expected []string
	}{
		{"", []string{""}},
		{"test", []string{"test"}},
		{"1 2", []string{"1", "2"}},
		{"1  2      3   4 ", []string{"1", "2", "3", "4"}},
	}

	for _, tt := range tests {
		if got := splitBySpace(tt.s); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("splitBySpace(%v) failed:\n  got %#v\n want %#v", tt.s, got, tt.expected)
		}
	}
}
