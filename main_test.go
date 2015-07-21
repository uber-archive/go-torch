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
	"testing"

	"github.com/codegangsta/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateAndRunApp(t *testing.T) {
	mockCommander := new(mockCommander)
	torcher := &torcher{
		commander: mockCommander,
	}

	var validateContext = func(args mock.Arguments) {
		context := args.Get(0).(*cli.Context)
		assert.NotNil(t, context)
		assert.Equal(t, "go-torch", context.App.Name)
	}
	mockCommander.On("goTorchCommand", mock.AnythingOfType("*cli.Context")).Return().Run(validateContext).Once()

	torcher.createAndRunApp()
}

func TestCreateAndRunAppDefaultValues(t *testing.T) {
	mockCommander := new(mockCommander)
	torcher := &torcher{
		commander: mockCommander,
	}

	validateDefaults := func(args mock.Arguments) {
		context := args.Get(0).(*cli.Context)
		assert.Equal(t, 30, context.Int("time"))
		assert.Equal(t, "http://localhost:8080", context.String("url"))
		assert.Equal(t, "/debug/pprof/profile", context.String("suffix"))
		assert.Equal(t, "torch.svg", context.String("file"))
		assert.Equal(t, "", context.String("binaryinput"))
		assert.Equal(t, false, context.Bool("print"))
		assert.Equal(t, false, context.Bool("raw"))
		assert.Equal(t, 9, len(context.App.Flags))
	}
	mockCommander.On("goTorchCommand", mock.AnythingOfType(
		"*cli.Context")).Return().Run(validateDefaults)

	torcher.createAndRunApp()
}

func TestGoTorchCommand(t *testing.T) {
	mockValidator := new(mockValidator)
	mockPprofer := new(mockPprofer)
	mockGrapher := new(mockGrapher)
	mockVisualizer := new(mockVisualizer)
	commander := &defaultCommander{
		validator:  mockValidator,
		pprofer:    mockPprofer,
		grapher:    mockGrapher,
		visualizer: mockVisualizer,
	}
	samplePprofOutput := []byte("out")

	mockValidator.On("validateArgument", "torch.svg", `\w+\.svg`,
		"Output file name must be .svg").Return(nil).Once()
	mockPprofer.On("runPprofCommand", 30, "http://localhost/hi").Return(samplePprofOutput, nil).Once()
	mockGrapher.On("GraphAsText", samplePprofOutput).Return("1;2;3 3", nil).Once()
	mockVisualizer.On("GenerateFlameGraph", "1;2;3 3", "torch.svg", false).Return(nil).Once()

	createSampleContext(commander)

	mockValidator.AssertExpectations(t)
	mockPprofer.AssertExpectations(t)
	mockGrapher.AssertExpectations(t)
	mockVisualizer.AssertExpectations(t)
}

func TestGoTorchCommandRawOutput(t *testing.T) {
	mockValidator := new(mockValidator)
	mockPprofer := new(mockPprofer)
	mockGrapher := new(mockGrapher)
	mockVisualizer := new(mockVisualizer)
	commander := &defaultCommander{
		validator:  mockValidator,
		pprofer:    mockPprofer,
		grapher:    mockGrapher,
		visualizer: mockVisualizer,
	}
	samplePprofOutput := []byte("out")
	mockValidator.On("validateArgument", "torch.svg", `\w+\.svg`,
		"Output file name must be .svg").Return(nil).Once()
	mockPprofer.On("runPprofCommand", 30, "http://localhost/hi").Return(samplePprofOutput, nil).Once()
	mockGrapher.On("GraphAsText", samplePprofOutput).Return("1;2;3 3", nil).Once()

	createSampleContextForRaw(commander)

	mockValidator.AssertExpectations(t)
	mockPprofer.AssertExpectations(t)
	mockGrapher.AssertExpectations(t)
	mockVisualizer.AssertExpectations(t) // ensure that mockVisualizer was never called
}

func TestGoTorchCommandBinaryInput(t *testing.T) {
	mockValidator := new(mockValidator)
	mockPprofer := new(mockPprofer)
	mockGrapher := new(mockGrapher)
	mockVisualizer := new(mockVisualizer)
	commander := &defaultCommander{
		validator:  mockValidator,
		pprofer:    mockPprofer,
		grapher:    mockGrapher,
		visualizer: mockVisualizer,
	}
	samplePprofOutput := []byte("out")
	mockValidator.On("validateArgument", "torch.svg", `\w+\.svg`,
		"Output file name must be .svg").Return(nil).Once()
	mockPprofer.On("runPprofCommand", 30, "/path/to/binary/input").Return(samplePprofOutput, nil).Once()
	mockGrapher.On("GraphAsText", samplePprofOutput).Return("1;2;3 3", nil).Once()
	mockVisualizer.On("GenerateFlameGraph", "1;2;3 3", "torch.svg", false).Return(nil).Once()

	createSampleContextForBinaryInput(commander)

	mockValidator.AssertExpectations(t)
	mockPprofer.AssertExpectations(t)
	mockGrapher.AssertExpectations(t)
	mockVisualizer.AssertExpectations(t)
}

func TestValidateArgumentFail(t *testing.T) {
	validator := new(defaultValidator)
	assert.Error(t, validator.validateArgument("bad bad", `\w+\.svg`, "Message"))
}

func TestValidateArgumentPass(t *testing.T) {
	assert.NotPanics(t, func() {
		new(defaultValidator).validateArgument("good.svg", `\w+\.svg`, "Message")
	})
}

func createSampleContext(commander *defaultCommander) {
	app := cli.NewApp()
	app.Name = "go-torch"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "url, u",
			Value: "http://localhost",
		},
		cli.StringFlag{
			Name:  "suffix, s",
			Value: "/hi",
		},
		cli.IntFlag{
			Name:  "time, t",
			Value: 30,
		},
		cli.StringFlag{
			Name:  "file, f",
			Value: "torch.svg",
		},
	}
	app.Action = commander.goTorchCommand
	app.Run([]string{"go-torch"})
}

func createSampleContextForRaw(commander *defaultCommander) {
	app := cli.NewApp()
	app.Name = "go-torch"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "url, u",
			Value: "http://localhost",
		},
		cli.StringFlag{
			Name:  "suffix, s",
			Value: "/hi",
		},
		cli.IntFlag{
			Name:  "time, t",
			Value: 30,
		},
		cli.StringFlag{
			Name:  "file, f",
			Value: "torch.svg",
		},
		cli.BoolTFlag{
			Name: "raw, r",
		},
	}
	app.Action = commander.goTorchCommand
	app.Run([]string{"go-torch"})
}

func createSampleContextForBinaryInput(commander *defaultCommander) {
	app := cli.NewApp()
	app.Name = "go-torch"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "url, u",
			Value: "http://localhost",
		},
		cli.StringFlag{
			Name:  "suffix, s",
			Value: "/hi",
		},
		cli.StringFlag{
			Name:  "binaryinput, b",
			Value: "/path/to/binary/input",
		},
		cli.IntFlag{
			Name:  "time, t",
			Value: 30,
		},
		cli.StringFlag{
			Name:  "file, f",
			Value: "torch.svg",
		},
	}
	app.Action = commander.goTorchCommand
	app.Run([]string{"go-torch"})
}
