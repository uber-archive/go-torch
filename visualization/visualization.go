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

// Package visualization handles the generation of the
// flame graph visualization.
package visualization

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
)

var errNoPerlScript = errors.New("Cannot find flamegraph script in the PATH or current " +
	"directory. You can download the script at https://github.com/brendangregg/FlameGraph. " +
	"Alternatively, you can run go-torch with the --raw flag.")

// Visualizer takes a graph in the format specified at
// https://github.com/brendangregg/FlameGraph and creates a svg flame graph
// using Brendan Gregg's flame graph perl script
type Visualizer interface {
	GenerateFlameGraph(string, string, bool) error
}

type defaultVisualizer struct {
	executor executor
}

type osWrapper interface {
	execLookPath(string) (string, error)
	cmdOutput(*exec.Cmd) ([]byte, error)
}

type defaultOSWrapper struct{}

type executor interface {
	createFile(string, []byte) error
	runPerlScript(string) ([]byte, error)
}

type defaultExecutor struct {
	osWrapper osWrapper
}

func newExecutor() executor {
	return &defaultExecutor{
		osWrapper: new(defaultOSWrapper),
	}
}

// NewVisualizer returns a visualizer struct with default fileCreator
func NewVisualizer() Visualizer {
	return &defaultVisualizer{
		executor: newExecutor(),
	}
}

// GenerateFlameGraph is the standard implementation of Visualizer
func (v *defaultVisualizer) GenerateFlameGraph(graphInput, outputFilePath string, stdout bool) error {
	out, err := v.executor.runPerlScript(graphInput)
	if err != nil {
		return err
	}
	if stdout {
		fmt.Println(string(out))
		log.Info("flame graph has been printed to stdout")
		return nil
	}
	if err = v.executor.createFile(outputFilePath, out); err != nil {
		return err
	}
	log.Info("flame graph has been created as " + outputFilePath)

	return nil
}

// runPerlScript checks whether the flamegraph script exists in the PATH or current directory and
// then executes it with the graphInput.
func (e *defaultExecutor) runPerlScript(graphInput string) ([]byte, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	possibilities := []string{"flamegraph.pl", cwd + "/flamegraph.pl", "flame-graph-gen"}
	perlScript := ""
	for _, path := range possibilities {
		perlScript, err = e.osWrapper.execLookPath(path)
		// found a valid script
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, errNoPerlScript
	}
	cmd := exec.Command(perlScript, os.Stdin.Name())
	cmd.Stdin = strings.NewReader(graphInput)
	out, err := e.osWrapper.cmdOutput(cmd)
	return out, err
}

// execLookPath is a tiny wrapper around exec.LookPath to enable test mocking
func (w *defaultOSWrapper) execLookPath(path string) (fullPath string, err error) {
	return exec.LookPath(path)
}

// cmdOutput is a tiny wrapper around cmd.Output to enable test mocking
func (w *defaultOSWrapper) cmdOutput(cmd *exec.Cmd) ([]byte, error) {
	return cmd.Output()
}

// createFile creates a file at a given path with given contents. If a file
// already exists at the path, it will be overwritten and replaced.
func (e *defaultExecutor) createFile(filePath string, fileContents []byte) error {
	os.Remove(filePath)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(fileContents)
	if err != nil {
		return err
	}
	return nil
}
