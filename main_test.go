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

package main

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	gflags "github.com/jessevdk/go-flags"
)

const testPProfInputFile = "./pprof/testdata/pprof.1.pb.gz"

func getDefaultOptions() *options {
	opts := &options{}
	if _, err := gflags.ParseArgs(opts, nil); err != nil {
		panic(err)
	}
	opts.PProfOptions.BinaryFile = testPProfInputFile
	return opts
}

func TestBadArgs(t *testing.T) {
	err := runWithArgs("-t", "asd")
	if err == nil {
		t.Fatalf("expected run with bad arguments to fail")
	}

	expectedSubstr := []string{
		"could not parse options",
		"invalid argument",
	}
	for _, substr := range expectedSubstr {
		if !strings.Contains(err.Error(), substr) {
			t.Errorf("error is missing message: %v", substr)
		}
	}
}

func TestMain(t *testing.T) {
	os.Args = []string{"go-torch", "--raw", "--binaryinput", testPProfInputFile}
	main()
	// Test should not fatal.
}

func TestMainRemaining(t *testing.T) {
	os.Args = []string{"go-torch", "--raw", testPProfInputFile}
	main()
	// Test should not fatal.
}

func TestInvalidOptions(t *testing.T) {
	tests := []struct {
		args         []string
		errorMessage string
	}{
		{
			args:         []string{"--file", "bad.jpg"},
			errorMessage: "must end in .svg",
		},
		{
			args:         []string{"-t", "0"},
			errorMessage: "seconds must be an integer greater than 0",
		},
		{
			args:         []string{"--title", ""},
			errorMessage: "flamegraph title should not be empty",
		},
		{
			args:         []string{"--width", "0"},
			errorMessage: "flamegraph default width is 1200 pixels",
		},
		{
			args:         []string{"--colors", "foo"},
			errorMessage: "unknown flamegraph colors \"foo\"",
		},
	}

	for _, tt := range tests {
		err := runWithArgs(tt.args...)
		if err == nil {
			t.Errorf("Expected error when running with: %v", tt.args)
			continue
		}

		if !strings.Contains(err.Error(), tt.errorMessage) {
			t.Errorf("Error missing message, got %v want message %v", err.Error(), tt.errorMessage)
		}
	}
}

func TestRunRaw(t *testing.T) {
	opts := getDefaultOptions()
	opts.OutputOpts.Raw = true

	if err := runWithOptions(opts, nil); err != nil {
		t.Fatalf("Run with Raw failed: %v", err)
	}
}

func TestFlameGraphArgs(t *testing.T) {
	opts := getDefaultOptions()
	opts.OutputOpts.Raw = true

	opts.OutputOpts.Hash = true
	opts.OutputOpts.Colors = "perl"
	opts.OutputOpts.ConsistentPalette = true
	opts.OutputOpts.Reverse = true
	opts.OutputOpts.Inverted = true

	expectedCommandWithArgs := []string{"--title", "Flame Graph", "--width", "1200", "--colors", "perl",
		"--hash", "--cp", "--reverse", "--inverted"}

	if !reflect.DeepEqual(expectedCommandWithArgs, buildFlameGraphArgs(opts.OutputOpts)) {
		t.Fatalf("Invalid extra FlameGraph arguments!")
	}

	if err := runWithOptions(opts, nil); err != nil {
		t.Fatalf("Run with extra FlameGraph arguments failed: %v", err)
	}
}

func getTempFilename(t *testing.T, suffix string) string {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	defer f.Close()
	return f.Name() + suffix
}

func TestRunFile(t *testing.T) {
	opts := getDefaultOptions()
	opts.OutputOpts.File = getTempFilename(t, ".svg")

	withScriptsInPath(t, func() {
		if err := runWithOptions(opts, nil); err != nil {
			t.Fatalf("Run with Print failed: %v", err)
		}

		f, err := os.Open(opts.OutputOpts.File)
		if err != nil {
			t.Errorf("Failed to open output file: %v", err)
		}
		defer f.Close()

		// Our fake flamegraph scripts just add script names to the output.
		reader := bufio.NewReader(f)
		line1, err := reader.ReadString('\n')
		if err != nil {
			t.Errorf("Failed to read line 1 in output file: %v", err)
		}
		if !strings.Contains(line1, "flamegraph.pl") {
			t.Errorf("Output file has not been processed by flame graph scripts")
		}
	})
}

func TestRunBadFile(t *testing.T) {
	opts := getDefaultOptions()
	opts.OutputOpts.File = "/dev/zero/invalid/file"

	withScriptsInPath(t, func() {
		if err := runWithOptions(opts, nil); err == nil {
			t.Fatalf("Run with bad file expected to fail")
		}
	})
}

func TestRunPrint(t *testing.T) {
	opts := getDefaultOptions()
	opts.OutputOpts.Print = true

	withScriptsInPath(t, func() {
		if err := runWithOptions(opts, nil); err != nil {
			t.Fatalf("Run with Print failed: %v", err)
		}
		// TODO(prashantv): Verify that output is printed to stdout.
	})
}

// scriptsPath is used to cache the fake scripts if we've already created it.
var scriptsPath string

func withScriptsInPath(t *testing.T, f func()) {
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)

	// Create a temporary directory with fake flamegraph scripts if we haven't already.
	if scriptsPath == "" {
		var err error
		scriptsPath, err = ioutil.TempDir("", "go-torch-scripts")
		if err != nil {
			t.Fatalf("Failed to create temporary scripts dir: %v", err)
		}

		// Create scripts in this path.
		const scriptContents = `#!/bin/sh
		echo $0
		cat
		`
		scriptFile := filepath.Join(scriptsPath, "flamegraph.pl")
		if err := ioutil.WriteFile(scriptFile, []byte(scriptContents), 0777); err != nil {
			t.Errorf("Failed to create script %v: %v", scriptFile, err)
		}
	}

	os.Setenv("PATH", scriptsPath+":"+oldPath)
	f()
}
