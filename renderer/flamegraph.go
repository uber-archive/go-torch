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
	"bytes"
	"errors"
	"os"
	"os/exec"
)

var errNoPerlScript = errors.New("Cannot find flamegraph scripts in the PATH or current " +
	"directory. You can download the script at https://github.com/brendangregg/FlameGraph. " +
	"These scripts should be added to your PATH or in the directory where go-torch is executed. " +
	"Alternatively, you can run go-torch with the --raw flag.")

var (
	stackCollapseScripts = []string{"stackcollapse.pl", "./stackcollapse.pl", "./FlameGraph/stackcollapse.pl"}
	flameGraphScripts    = []string{"flamegraph", "flamegraph.pl", "./flamegraph.pl", "./FlameGraph/flamegraph.pl", "flame-graph-gen"}
)

// findInPath returns the first path that is found in PATH.
func findInPath(paths []string) string {
	for _, v := range paths {
		if path, err := exec.LookPath(v); err == nil {
			return path
		}
	}
	return ""
}

// runScript runs scriptName with the given arguments, and stdin set to inData.
// It returns the stdout on success.
func runScript(scriptName string, args []string, inData []byte) ([]byte, error) {
	cmd := exec.Command(scriptName, args...)
	cmd.Stdin = bytes.NewReader(inData)
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

// CollapseStacks runs the flamegraph's collapse stacks script.
func CollapseStacks(stacks []byte, args ...string) ([]byte, error) {
	stackCollapse := findInPath(stackCollapseScripts)
	if stackCollapse == "" {
		return nil, errNoPerlScript
	}

	return runScript(stackCollapse, nil, stacks)
}

// GenerateFlameGraph runs the flamegraph script to generate a flame graph SVG.
func GenerateFlameGraph(graphInput []byte, args ...string) ([]byte, error) {
	flameGraph := findInPath(flameGraphScripts)
	if flameGraph == "" {
		return nil, errNoPerlScript
	}

	return runScript(flameGraph, args, graphInput)
}
