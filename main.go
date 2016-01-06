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
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/uber/go-torch/pprof"
	"github.com/uber/go-torch/renderer"
	"github.com/uber/go-torch/torchlog"

	gflags "github.com/jessevdk/go-flags"
)

// options are the parameters for go-torch.
type options struct {
	PProfOptions pprof.Options
	File         string `short:"f" long:"file" default:"torch.svg" description:"Output file name (must be .svg)"`
	Print        bool   `short:"p" long:"print" description:"Print the generated svg to stdout instead of writing to file"`
	Raw          bool   `short:"r" long:"raw" description:"Print the raw call graph output to stdout instead of creating a flame graph; use with Brendan Gregg's flame graph perl script (see https://github.com/brendangregg/FlameGraph)"`
	Title        string `long:"title" default:"Flame Graph" description:"Graph title to display in the output file"`
}

// main is the entry point of the application
func main() {
	if err := runWithArgs(os.Args...); err != nil {
		torchlog.Fatalf("Failed: %v", err)
	}
}

func runWithArgs(args ...string) error {
	opts := &options{}
	if _, err := gflags.ParseArgs(opts, args); err != nil {
		if flagErr, ok := err.(*gflags.Error); ok && flagErr.Type == gflags.ErrHelp {
			os.Exit(0)
		}
		return fmt.Errorf("could not parse options: %v", err)
	}
	if err := validateOptions(opts); err != nil {
		return fmt.Errorf("invalid options: %v", err)
	}

	return runWithOptions(opts)
}

func runWithOptions(opts *options) error {
	pprofRawOutput, err := pprof.GetRaw(opts.PProfOptions)
	if err != nil {
		return fmt.Errorf("could not get raw output from pprof: %v", err)
	}

	callStacks, err := pprof.ParseRaw(pprofRawOutput)
	if err != nil {
		return fmt.Errorf("could not parse raw pprof output: %v", err)
	}

	flameInput, err := renderer.ToFlameInput(callStacks)
	if err != nil {
		return fmt.Errorf("could not convert stacks to flamegraph input: %v", err)
	}

	if opts.Raw {
		torchlog.Print("Printing raw flamegraph input to stdout")
		fmt.Printf("%s", flameInput)
		return nil
	}

	flameGraph, err := renderer.GenerateFlameGraph(flameInput, "--title", opts.Title)
	if err != nil {
		return fmt.Errorf("could not generate flame graph: %v", err)
	}

	if opts.Print {
		torchlog.Print("Printing svg to stdout")
		fmt.Printf("%s", flameGraph)
		return nil
	}

	torchlog.Printf("Writing svg to %v", opts.File)
	if err := ioutil.WriteFile(opts.File, flameGraph, 0666); err != nil {
		return fmt.Errorf("could not write output file: %v", err)
	}

	return nil
}

func validateOptions(opts *options) error {
	if opts.File != "" && !strings.HasSuffix(opts.File, ".svg") {
		return fmt.Errorf("output file must end in .svg")
	}
	if opts.PProfOptions.TimeSeconds < 1 {
		return fmt.Errorf("seconds must be an integer greater than 0")
	}
	return nil
}
