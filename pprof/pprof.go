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
	"bytes"
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	"github.com/uber/go-torch/torchlog"
)

// Options are parameters for pprof.
type Options struct {
	BaseURL     string   `short:"u" long:"url" default:"http://localhost:8080" description:"Base URL of your Go program"`
	URLSuffix   string   `long:"suffix" default:"/debug/pprof/profile" description:"URL path of pprof profile"`
	BinaryFile  string   `short:"b" long:"binaryinput" description:"File path of previously saved binary profile. (binary profile is anything accepted by https://golang.org/cmd/pprof)"`
	BinaryName  string   `long:"binaryname" description:"File path of the binary that the binaryinput is for, used for pprof inputs"`
	TimeSeconds int      `short:"t" long:"seconds" default:"30" description:"Number of seconds to profile for"`
	ExtraArgs   []string `long:"pprofArgs"  description:"Extra arguments for pprof"`
	TimeAlias   *int     `hidden:"true" long:"time" description:"Alias for backwards compatibility"`
}

// GetRaw returns the raw output from pprof for the given options.
func GetRaw(opts Options, remaining []string) ([]byte, error) {
	args, err := getArgs(opts, remaining)
	if err != nil {
		return nil, err
	}

	return runPProf(args...)
}

// getArgs gets the arguments to run pprof with for a given set of Options.
func getArgs(opts Options, remaining []string) ([]string, error) {
	if opts.TimeAlias != nil {
		opts.TimeSeconds = *opts.TimeAlias
	}
	if len(remaining) > 0 {
		var pprofArgs []string
		if opts.TimeSeconds > 0 {
			pprofArgs = append(pprofArgs, "-seconds", fmt.Sprint(opts.TimeSeconds))
		}
		pprofArgs = append(pprofArgs, remaining...)
		return pprofArgs, nil
	}

	pprofArgs := opts.ExtraArgs
	if opts.BinaryFile != "" {
		if opts.BinaryName != "" {
			pprofArgs = append(pprofArgs, opts.BinaryName)
		}
		pprofArgs = append(pprofArgs, opts.BinaryFile)
	} else {
		u, err := url.Parse(opts.BaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %v", err)
		}

		u.Path = opts.URLSuffix
		pprofArgs = append(pprofArgs, "-seconds", fmt.Sprint(opts.TimeSeconds), u.String())
	}

	return pprofArgs, nil
}

func runPProf(args ...string) ([]byte, error) {
	allArgs := []string{"tool", "pprof", "-raw"}
	allArgs = append(allArgs, args...)

	var buf bytes.Buffer
	torchlog.Printf("Run pprof command: go %v", strings.Join(allArgs, " "))
	cmd := exec.Command("go", allArgs...)
	cmd.Stderr = &buf
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("pprof error: %v\nSTDERR:\n%s", err, buf.Bytes())
	}

	// @HACK because 'go tool pprof' doesn't exit on errors with nonzero status codes.
	// Ironically, this means that Go's own os/exec package does not detect its errors.
	// See issue here https://github.com/golang/go/issues/11510
	if len(out) == 0 {
		return nil, fmt.Errorf("pprof error:\n%s", buf.Bytes())
	}

	return out, nil
}
