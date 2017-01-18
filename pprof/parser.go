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
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/uber/go-torch/stack"
)

type readState int

const (
	ignore readState = iota
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

	state       readState
	funcNames   map[funcID]string
	sampleNames []string
	records     []*stackRecord
}

// ParseRaw parses the raw pprof output and returns call stacks.
func ParseRaw(input []byte) (*stack.Profile, error) {
	parser := newRawParser()
	if err := parser.parse(input); err != nil {
		return nil, err
	}

	return parser.toProfile()
}

func newRawParser() *rawParser {
	return &rawParser{
		funcNames: make(map[funcID]string),
	}
}

func (p *rawParser) parse(input []byte) error {
	reader := bufio.NewReader(bytes.NewReader(input))

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if p.state < locations {
					p.setError(fmt.Errorf("parser ended before processing locations, state: %v", p.state))
				}
				break
			}
			return err
		}

		p.processLine(strings.TrimSpace(line))
	}

	return p.err
}

func (p *rawParser) setError(err error) {
	if p.err != nil {
		return
	}
	p.err = err
}

func (p *rawParser) processLine(line string) {
	switch p.state {
	case ignore:
		if strings.HasPrefix(line, "Samples") {
			p.state = samplesHeader
			return
		}
	case samplesHeader:
		p.sampleNames = strings.Split(line, " ")
		p.state = samples
	case samples:
		if strings.HasPrefix(line, "Locations") {
			p.state = locations
			return
		}
		p.addSample(line)
	case locations:
		if strings.HasPrefix(line, "Mappings") {
			p.state = mappings
			return
		}
		p.addLocation(line)
	case mappings:
		// Nothing to process.
	}
}

// toProfile aggregates stack sample counts and returns a profile with unique stack samples.
func (p *rawParser) toProfile() (*stack.Profile, error) {
	profile, err := stack.NewProfile(p.sampleNames)
	if err != nil {
		return nil, err
	}

	samples := make(map[string]*stack.Sample)
	for _, r := range p.records {
		funcNames := r.funcNames(p.funcNames)
		funcKey := strings.Join(funcNames, ";")

		if sample, ok := samples[funcKey]; ok {
			if err := sample.Add(r.samples); err != nil {
				return nil, err
			}
			continue
		}

		samples[funcKey] = stack.NewSample(funcNames, r.samples)
	}

	profile.Samples = make([]*stack.Sample, 0, len(samples))
	for _, s := range samples {
		profile.Samples = append(profile.Samples, s)
	}

	return profile, nil
}

// addLocation parses a location that looks like:
//   292: 0x49dee1 github.com/uber/tchannel/golang.(*Frame).ReadIn :0 s=0
// and creates a mapping from funcID to function name.
func (p *rawParser) addLocation(line string) {
	parts := splitBySpace(line)
	if len(parts) < 4 {
		// Do not error when there is only id and address in the line.
		if len(parts) != 2 {
			p.setError(fmt.Errorf("malformed location line: %v", line))
		}
		return
	}
	funcID := p.toFuncID(strings.TrimSuffix(parts[0], ":"))
	if strings.HasPrefix(parts[2], "M=") {
		p.funcNames[funcID] = parts[3]
	} else {
		p.funcNames[funcID] = parts[2]
	}
}

type stackRecord struct {
	samples []int64
	stack   []funcID
}

// addSample parses a sample that looks like:
//   1   10000000: 1 2 3 4
// and creates a stackRecord for it.
func (p *rawParser) addSample(line string) {
	if strings.Contains(line, "bytes:[") {
		// Memory profiles have a line of bytes:[size1] which contains the size
		// of the object. Skip these lines as we do not use it currently.
		return
	}

	// Split by ":" which separates the data from the function IDs.
	lineParts := strings.Split(line, ":")
	if len(lineParts) != 2 {
		p.setError(fmt.Errorf("malformed sample line: %v", line))
		return
	}

	samples := p.parseInts(lineParts[0])
	funcIDs := p.parseFuncIDs(lineParts[1])

	if len(samples) != len(p.sampleNames) {
		p.setError(fmt.Errorf("line has a different sample count (%v) than sample names (%v): %v",
			len(samples), len(p.sampleNames), line))
		return
	}

	p.records = append(p.records, &stackRecord{
		samples: samples,
		stack:   funcIDs,
	})
}

func getFunctionName(funcNames map[funcID]string, funcID funcID) string {
	if funcName, ok := funcNames[funcID]; ok {
		return funcName
	}
	return fmt.Sprintf("missing-function-%v", funcID)
}

// funcNames returns the function names for this stack sample.
// It returns in parent first order.
func (r *stackRecord) funcNames(funcNames map[funcID]string) []string {
	var names []string
	for i := len(r.stack) - 1; i >= 0; i-- {
		funcID := r.stack[i]
		names = append(names, getFunctionName(funcNames, funcID))
	}
	return names
}

func (p *rawParser) parseFuncIDs(s string) []funcID {
	funcInts := p.parseInts(s)
	funcIDs := make([]funcID, len(funcInts))
	for i, fID := range funcInts {
		funcIDs[i] = funcID(fID)
	}
	return funcIDs
}

func (p *rawParser) parseInts(s string) []int64 {
	ss := splitBySpace(s)
	samples := make([]int64, len(ss))
	for i, s := range ss {
		samples[i] = p.parseInt(s)
	}
	return samples
}

// parseInt converts a string to an int64. It stores any errors using setError.
func (p *rawParser) parseInt(s string) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		p.setError(err)
		return 0
	}

	return v
}

// toFuncID converts a string like "8" to a funcID.
func (p *rawParser) toFuncID(s string) funcID {
	return funcID(p.parseInt(s))
}

var spaceSplitter = regexp.MustCompile(`\s+`)

// splitBySpace splits values separated by 1 or more spaces.
func splitBySpace(s string) []string {
	return spaceSplitter.Split(strings.TrimSpace(s), -1)
}
