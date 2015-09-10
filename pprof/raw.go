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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type readMode int

const (
	ignore readMode = iota
	samplesHeader
	samples
	locations
	mappings
)

// funcID is the ID of a given Location in the pprof raw output.
type funcID int

type rawParser struct {
	// err is the first error encountered by the parser.
	err error

	mode     readMode
	funcName map[funcID]string
	records  []*stackRecord
}

// ParseRaw parses the raw pprof output and returns call stacks.
func ParseRaw(input []byte) ([]byte, error) {
	parser := newRawParser()
	if err := parser.parse(input); err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()
	go parser.print(pw)
	return CollapseStacks(pr)
}

func newRawParser() *rawParser {
	return &rawParser{
		funcName: make(map[funcID]string),
	}
}

func (p *rawParser) parse(input []byte) error {
	reader := bufio.NewReader(bytes.NewReader(input))

	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		p.processLine(line)
	}

	return p.err
}

func (p *rawParser) processLine(line string) {
	switch p.mode {
	case ignore:
		if strings.HasPrefix(line, "Samples") {
			p.mode = samplesHeader
			return
		}
	case samplesHeader:
		p.mode = samples
	case samples:
		if strings.HasPrefix(line, "Locations") {
			p.mode = locations
			return
		}
		p.addSample(line)
	case locations:
		if strings.HasPrefix(line, "Mappings") {
			p.mode = mappings
			return
		}
		p.addLocation(line)
	case mappings:
		// Nothing to process.
	}
}

// print prints out the stack traces collected from the raw pprof output.
func (p *rawParser) print(w io.WriteCloser) {
	for _, r := range p.records {
		r.Serialize(p.funcName, w)
		fmt.Fprintln(w)
	}
	w.Close()
}

func findStackCollapse() string {
	for _, v := range []string{"stackcollapse.pl", "./stackcollapse.pl", "./FlameGraph/stackcollapse.pl"} {
		if path, err := exec.LookPath(v); err == nil {
			return path
		}
	}
	return ""
}

// CollapseStacks runs the flamegraph's collapse stacks script.
func CollapseStacks(stacks io.Reader) ([]byte, error) {
	stackCollapse := findStackCollapse()
	if stackCollapse == "" {
		return nil, errors.New("stackcollapse.pl not found")
	}

	cmd := exec.Command(stackCollapse)
	cmd.Stdin = stacks
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

// addSample parses a sample that looks like:
//   1   10000000: 1 2 3 4
// and creates a stackRecord for it.
func (p *rawParser) addSample(line string) {
	// Parse a sample which looks like:
	parts := splitBySpace(line)

	samples, err := strconv.Atoi(parts[0])
	if err != nil {
		p.err = err
		return
	}

	duration, err := strconv.Atoi(strings.TrimSuffix(parts[1], ":"))
	if err != nil {
		p.err = err
		return
	}

	var stack []funcID
	for _, fIDStr := range parts[2:] {
		stack = append(stack, p.toFuncID(fIDStr))
	}

	p.records = append(p.records, &stackRecord{samples, time.Duration(duration), stack})
}

// addLocation parses a location that looks like:
//   292: 0x49dee1 github.com/uber/tchannel/golang.(*Frame).ReadIn :0 s=0
// and creates a mapping from funcID to function name.
func (p *rawParser) addLocation(line string) {
	parts := splitBySpace(line)
	funcID := p.toFuncID(strings.TrimSuffix(parts[0], ":"))
	p.funcName[funcID] = parts[2]
}

type stackRecord struct {
	samples  int
	duration time.Duration
	stack    []funcID
}

// Serialize serializes a call stack for a given stackRecord given the funcID mapping.
func (r *stackRecord) Serialize(funcName map[funcID]string, w io.Writer) {
	for _, funcID := range r.stack {
		fmt.Fprintln(w, funcName[funcID])
	}
	fmt.Fprintln(w, r.samples)
}

// toFuncID converts a string like "8" to a funcID.
func (p *rawParser) toFuncID(s string) funcID {
	i, err := strconv.Atoi(s)
	if err != nil {
		p.err = fmt.Errorf("failed to parse funcID: %v", err)
		return 0
	}
	return funcID(i)
}

var spaceSplitter = regexp.MustCompile(`\s+`)

// splitBySpace splits values separated by 1 or more spaces.
func splitBySpace(s string) []string {
	return spaceSplitter.Split(strings.TrimSpace(s), -1)
}
