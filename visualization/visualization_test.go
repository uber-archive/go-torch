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

package visualization

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateFile(t *testing.T) {
	new(defaultExecutor).createFile(".text.svg", []byte("the contents"))

	// teardown
	defer os.Remove(".text.svg")

	actualContents, err := ioutil.ReadFile(".text.svg")
	assert.NoError(t, err)
	assert.Equal(t, "the contents", string(actualContents))
}

func TestCreateFileOverwriteExisting(t *testing.T) {
	new(defaultExecutor).createFile(".text.svg", []byte("delete me"))
	new(defaultExecutor).createFile(".text.svg", []byte("correct answer"))

	// teardown
	defer os.Remove(".text.svg")

	actualContents, err := ioutil.ReadFile(".text.svg")
	assert.NoError(t, err)
	assert.Equal(t, "correct answer", string(actualContents))
}

func TestGenerateFlameGraph(t *testing.T) {
	mockExecutor := new(mockExecutor)
	visualizer := defaultVisualizer{
		executor: mockExecutor,
	}

	graphInput := "N4;N5 1\nN4;N6;N5 8\n"

	mockExecutor.On("runPerlScript", graphInput).Return([]byte("<svg></svg>"), nil).Once()
	mockExecutor.On("createFile", ".text.svg", mock.AnythingOfType("[]uint8")).Return(nil).Once()

	visualizer.GenerateFlameGraph(graphInput, ".text.svg", false)

	mockExecutor.AssertExpectations(t)
}

func TestGenerateFlameGraphPrintsToStdout(t *testing.T) {
	mockExecutor := new(mockExecutor)
	visualizer := defaultVisualizer{
		executor: mockExecutor,
	}
	graphInput := "N4;N5 1\nN4;N6;N5 8\n"
	mockExecutor.On("runPerlScript", graphInput).Return([]byte("<svg></svg>"), nil).Once()
	visualizer.GenerateFlameGraph(graphInput, ".text.svg", true)

	mockExecutor.AssertNotCalled(t, "createFile")
	mockExecutor.AssertExpectations(t)
}

// Underlying errors can occur in runPerlScript(). This test ensures that errors
// like a missing flamegraph.pl script or malformed input are propagated.
func TestGenerateFlameGraphExecError(t *testing.T) {
	mockExecutor := new(mockExecutor)
	visualizer := defaultVisualizer{
		executor: mockExecutor,
	}
	mockExecutor.On("runPerlScript", "").Return(nil, errors.New("bad input")).Once()

	err := visualizer.GenerateFlameGraph("", ".text.svg", false)
	assert.Error(t, err)
	mockExecutor.AssertNotCalled(t, "createFile")
	mockExecutor.AssertExpectations(t)
}

// Smoke test the NewVisualizer method
func TestNewVisualizer(t *testing.T) {
	assert.NotNil(t, NewVisualizer())
}
