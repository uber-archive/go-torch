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

// Package main is the entry point of go-torch, a stochastic flame graph
// profiler for Go programs.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"

	"github.com/uber/go-torch/graph"
	"github.com/uber/go-torch/visualization"
)

type torcher struct {
	commander commander
}

type commander interface {
	goTorchCommand(*cli.Context)
}

type defaultCommander struct {
	validator  validator
	pprofer    pprofer
	grapher    graph.Grapher
	visualizer visualization.Visualizer
}

type validator interface {
	validateArgument(string, string, string) error
}

type defaultValidator struct{}

type pprofer interface {
	runPprofCommand(int, string) ([]byte, error)
}

type defaultPprofer struct{}

// newTorcher returns a torcher struct with a default commander
func newTorcher() *torcher {
	return &torcher{
		commander: newCommander(),
	}
}

// newCommander returns a default commander struct with default attributes
func newCommander() commander {
	return &defaultCommander{
		validator:  new(defaultValidator),
		pprofer:    new(defaultPprofer),
		grapher:    graph.NewGrapher(),
		visualizer: visualization.NewVisualizer(),
	}
}

// main is the entry point of the application
func main() {
	t := newTorcher()
	t.createAndRunApp()
}

// createAndRunApp configures and runs a cli.App
func (t *torcher) createAndRunApp() {
	app := cli.NewApp()
	app.Name = "go-torch"
	app.Usage = "go-torch collects stack traces of a Go application and synthesizes them into into a flame graph"
	app.Version = "0.5"
	app.Authors = []cli.Author{{Name: "Ben Sandler", Email: "bens@uber.com"}}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "url, u",
			Value: "http://localhost:8080",
			Usage: "base url of your Go program",
		},
		cli.StringFlag{
			Name:  "suffix, s",
			Value: "/debug/pprof/profile",
			Usage: "url path of pprof profile",
		},
		cli.StringFlag{
			Name:  "binaryinput, b",
			Value: "",
			Usage: "file path of raw binary profile; alternative to having go-torch query pprof endpoint " +
				"(binary profile is anything accepted by https://golang.org/cmd/pprof)",
		},
		cli.IntFlag{
			Name:  "time, t",
			Value: 30,
			Usage: "time in seconds to profile for",
		},
		cli.StringFlag{
			Name:  "file, f",
			Value: "torch.svg",
			Usage: "ouput file name (must be .svg)",
		},
		cli.BoolFlag{
			Name:  "print, p",
			Usage: "print the generated svg to stdout instead of writing to file",
		},
		cli.BoolFlag{
			Name: "raw, r",
			Usage: "print the raw call graph output to stdout instead of creating a flame graph; " +
				"use with Brendan Gregg's flame graph perl script (see https://github.com/brendangregg/FlameGraph)",
		},
	}
	app.Action = t.commander.goTorchCommand
	app.Run(os.Args)
}

// goTorchCommand executes the 'go-torch' command.
func (com *defaultCommander) goTorchCommand(c *cli.Context) {
	url := c.String("url") + c.String("suffix")
	outputFile := c.String("file")
	binaryInput := c.String("binaryinput")
	time := c.Int("time")
	stdout := c.Bool("print")
	raw := c.Bool("raw")

	err := com.validator.validateArgument(outputFile, `\w+\.svg`, "Output file name must be .svg")
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Profiling ...")

	profileSource := ""
	if binaryInput != "" {
		profileSource = binaryInput
	} else {
		profileSource = url
	}

	out, err := com.pprofer.runPprofCommand(time, profileSource)
	if err != nil {
		log.Fatal(err)
	}
	flamegraphInput, err := com.grapher.GraphAsText(out)
	if err != nil {
		log.Fatal(err)
	}
	flamegraphInput = strings.TrimSpace(flamegraphInput)
	if raw {
		fmt.Println(flamegraphInput)
		log.Info("raw call graph output been printed to stdout")
		return
	}
	if err := com.visualizer.GenerateFlameGraph(flamegraphInput, outputFile, stdout); err != nil {
		log.Fatal(err)
	}
}

// runPprofCommand runs the `go tool pprof` command to profile an application.
// It returns the output of the underlying command.
func (p *defaultPprofer) runPprofCommand(time int, profileSource string) ([]byte, error) {
	timeArg := fmt.Sprintf("-seconds=%d", time)

	var buf bytes.Buffer
	cmd := exec.Command("go", "tool", "pprof", "-dot", "-lines", timeArg, profileSource)
	cmd.Stderr = &buf
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// @HACK because 'go tool pprof' doesn't exit on errors with nonzero status codes.
	// Ironically, this means that Go's own os/exec package does not detect its errors.
	// See issue here https://github.com/golang/go/issues/11510
	if len(out) == 0 {
		errText := buf.String()
		return nil, errors.New("pprof returned an error. Here is the raw STDERR output:\n" + errText)
	}

	return out, nil
}

// validateArgument validates a given command line argument with regex. If the
// argument does not match the expected format, this function returns an error.
func (v *defaultValidator) validateArgument(argument, regex, errorMessage string) error {
	match, _ := regexp.MatchString(regex, argument)
	if !match {
		return errors.New(errorMessage)
	}
	return nil
}
