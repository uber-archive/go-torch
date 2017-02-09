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
	"strconv"
	"strings"

	"github.com/uber/go-torch/pprof"
	"github.com/uber/go-torch/renderer"
	"github.com/uber/go-torch/torchlog"

	gflags "github.com/jessevdk/go-flags"
)

// options are the parameters for go-torch.
type options struct {
	PProfOptions pprof.Options `group:"pprof Options"`
	OutputOpts   outputOptions `group:"Output Options"`
}

type outputOptions struct {
	File              string `short:"f" long:"file" default:"torch.svg" description:"Output file name (must be .svg)"`
	Print             bool   `short:"p" long:"print" description:"Print the generated svg to stdout instead of writing to file"`
	Raw               bool   `short:"r" long:"raw" description:"Print the raw call graph output to stdout instead of creating a flame graph; use with Brendan Gregg's flame graph perl script (see https://github.com/brendangregg/FlameGraph)"`
	Title             string `long:"title" default:"Flame Graph" description:"Graph title to display in the output file"`
	Width             int64  `long:"width" default:"1200" description:"Generated graph width"`
	Hash              bool   `long:"hash" description:"Colors are keyed by function name hash"`
	Colors            string `long:"colors" default:"" description:"set color palette. choices are: hot (default), mem, io, wakeup, chain, java, js, perl, red, green, blue, aqua, yellow, purple, orange"`
	ConsistentPalette bool   `long:"cp" description:"Use consistent palette (palette.map)"`
	Reverse           bool   `long:"reverse" description:"Generate stack-reversed flame graph"`
	Inverted          bool   `long:"inverted" description:"icicle graph"`
}

// main is the entry point of the application
func main() {
	if err := runWithArgs(os.Args[1:]...); err != nil {
		torchlog.Fatalf("Failed: %v", err)
	}
}

func runWithArgs(args ...string) error {
	opts := &options{}

	parser := gflags.NewParser(opts, gflags.Default|gflags.IgnoreUnknown)
	parser.Usage = "[options] [binary] <profile source>"

	remaining, err := parser.ParseArgs(args)
	if err != nil {
		if flagErr, ok := err.(*gflags.Error); ok && flagErr.Type == gflags.ErrHelp {
			os.Exit(0)
		}
		return fmt.Errorf("could not parse options: %v", err)
	}
	if err := validateOptions(opts); err != nil {
		return fmt.Errorf("invalid options: %v", err)
	}

	return runWithOptions(opts, remaining)
}

func runWithOptions(allOpts *options, remaining []string) error {
	pprofRawOutput, err := pprof.GetRaw(allOpts.PProfOptions, remaining)
	if err != nil {
		return fmt.Errorf("could not get raw output from pprof: %v", err)
	}

	profile, err := pprof.ParseRaw(pprofRawOutput)
	if err != nil {
		return fmt.Errorf("could not parse raw pprof output: %v", err)
	}

	sampleIndex := pprof.SelectSample(remaining, profile.SampleNames)
	flameInput, err := renderer.ToFlameInput(profile, sampleIndex)
	if err != nil {
		return fmt.Errorf("could not convert stacks to flamegraph input: %v", err)
	}

	opts := allOpts.OutputOpts
	if opts.Raw {
		torchlog.Print("Printing raw flamegraph input to stdout")
		fmt.Printf("%s\n", flameInput)
		return nil
	}

	var flameGraphArgs = buildFlameGraphArgs(opts)
	flameGraph, err := renderer.GenerateFlameGraph(flameInput, flameGraphArgs...)
	if err != nil {
		return fmt.Errorf("could not generate flame graph: %v", err)
	}

	if opts.Print {
		torchlog.Print("Printing svg to stdout")
		fmt.Printf("%s\n", flameGraph)
		return nil
	}

	torchlog.Printf("Writing svg to %v", opts.File)
	if err := ioutil.WriteFile(opts.File, flameGraph, 0666); err != nil {
		return fmt.Errorf("could not write output file: %v", err)
	}

	return nil
}

func validateOptions(opts *options) error {
	file := opts.OutputOpts.File
	if file != "" && !strings.HasSuffix(file, ".svg") {
		return fmt.Errorf("output file must end in .svg")
	}
	if opts.PProfOptions.TimeSeconds < 1 {
		return fmt.Errorf("seconds must be an integer greater than 0")
	}

	// extra FlameGraph options
	if opts.OutputOpts.Title == "" {
		return fmt.Errorf("flamegraph title should not be empty")
	}
	if opts.OutputOpts.Width <= 0 {
		return fmt.Errorf("flamegraph default width is 1200 pixels")
	}
	if opts.OutputOpts.Colors != "" {
		switch opts.OutputOpts.Colors {
		case "hot", "mem", "io", "wakeup", "chain", "java", "js", "perl", "red", "green", "blue", "aqua", "yellow", "purple", "orange":
			// valid
		default:
			return fmt.Errorf("unknown flamegraph colors %q", opts.OutputOpts.Colors)
		}
	}

	return nil
}

func buildFlameGraphArgs(opts outputOptions) []string {
	var args []string

	if opts.Title != "" {
		args = append(args, "--title", opts.Title)
	}

	if opts.Width > 0 {
		args = append(args, "--width", strconv.FormatInt(opts.Width, 10))
	}

	if opts.Colors != "" {
		args = append(args, "--colors", opts.Colors)
	}

	if opts.Hash {
		args = append(args, "--hash")
	}

	if opts.ConsistentPalette {
		args = append(args, "--cp")
	}

	if opts.Reverse {
		args = append(args, "--reverse")
	}

	if opts.Inverted {
		args = append(args, "--inverted")
	}

	return args
}
