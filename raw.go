package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
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

type funcID int

type rawParser struct {
	funcName map[funcID]string
	records  []*stackRecord
}

func newRawParser() *rawParser {
	return &rawParser{
		funcName: make(map[funcID]string),
	}
}

func (p *rawParser) Parse(input []byte) ([]byte, error) {
	var mode readMode
	reader := bufio.NewReader(bytes.NewReader(input))

	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch mode {
		case ignore:
			if strings.HasPrefix(line, "Samples") {
				mode = samplesHeader
				continue
			}
		case samplesHeader:
			mode = samples
		case samples:
			if strings.HasPrefix(line, "Locations") {
				mode = locations
				continue
			}
			p.addSample(line)
		case locations:
			if strings.HasPrefix(line, "Mappings") {
				mode = mappings
				continue
			}
			p.addLocation(line)
		case mappings:
			// Nothing to process.
		}
	}

	pr, pw := io.Pipe()
	go p.Print(pw)
	return p.CollapseStacks(pr)
}

func (p *rawParser) Print(w io.WriteCloser) {
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

func (p *rawParser) CollapseStacks(stacks io.Reader) ([]byte, error) {
	stackCollapse := findStackCollapse()
	if stackCollapse == "" {
		return nil, errors.New("stackcollapse.pl not found")
	}

	cmd := exec.Command(stackCollapse)
	cmd.Stdin = stacks
	return cmd.Output()
}

func (p *rawParser) addSample(line string) {
	// Parse a sample which looks like:
	// 1   10000000: 1 2 3 4
	parts := splitIgnoreEmpty(line, " ")

	samples, err := strconv.Atoi(parts[0])
	if err != nil {
		panic(err)
	}

	duration, err := strconv.Atoi(strings.TrimSuffix(parts[1], ":"))
	if err != nil {
		panic(err)
	}

	var stack []funcID
	for _, fIDStr := range parts[2:] {
		stack = append(stack, toFuncID(fIDStr))
	}

	p.records = append(p.records, &stackRecord{samples, time.Duration(duration), stack})
}

func (p *rawParser) addLocation(line string) {
	// 292: 0x49dee1 github.com/uber/tchannel/golang.(*Frame).ReadIn :0 s=0
	parts := splitIgnoreEmpty(line, " ")
	funcID := toFuncID(strings.TrimSuffix(parts[0], ":"))
	p.funcName[funcID] = parts[2]
}

type stackRecord struct {
	samples  int
	duration time.Duration
	stack    []funcID
}

func (r *stackRecord) Serialize(funcName map[funcID]string, w io.Writer) {
	// Go backwards through the stack
	for _, funcID := range r.stack {
		fmt.Fprintln(w, funcName[funcID])
	}
	fmt.Fprintln(w, r.samples)
}

func toFuncID(s string) funcID {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return funcID(i)
}

// splitIgnoreEmpty does a strings.Split and then removes all empty strings.
func splitIgnoreEmpty(s string, splitter string) []string {
	vals := strings.Split(s, splitter)
	var res []string
	for _, v := range vals {
		if len(v) != 0 {
			res = append(res, v)
		}
	}

	return res
}
