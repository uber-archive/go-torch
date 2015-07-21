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

	"github.com/Sirupsen/logrus"
)

var errNoPerlScript = errors.New("Cannot find flamegraph.pl script in the PATH or current " +
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

type executor interface {
	createFile(string, []byte) error
	runPerlScript(string) ([]byte, error)
}

type defaultExecutor struct{}

// NewVisualizer returns a visualizer struct with default fileCreator
func NewVisualizer() Visualizer {
	return &defaultVisualizer{
		executor: new(defaultExecutor),
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
		logrus.Info("flame graph has been printed to stdout")
		return nil
	}
	if err = v.executor.createFile(outputFilePath, out); err != nil {
		return err
	}
	logrus.Info("flame graph has been created as " + outputFilePath)

	return nil
}

// runPerlScript checks whether the flamegraph script exists in the PATH or current directory and
// then executes it with the graphInput.
func (e *defaultExecutor) runPerlScript(graphInput string) ([]byte, error) {
	perlScript, err := exec.LookPath("flamegraph.pl")
	if err != nil {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		perlScript, err = exec.LookPath(cwd + "/flamegraph.pl")
		if err != nil {
			return nil, errNoPerlScript
		}
	}
	cmd := exec.Command(perlScript, os.Stdin.Name())
	cmd.Stdin = strings.NewReader(graphInput)

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
