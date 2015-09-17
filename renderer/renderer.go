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

package renderer

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/uber/go-torch/stack"
)

// ToFlameInput convers the given stack samples to flame graph input.
func ToFlameInput(samples []*stack.Sample) ([]byte, error) {
	buf := &bytes.Buffer{}
	for _, s := range samples {
		if err := renderSample(buf, s); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// renderSample renders a single stack sample as flame graph input.
func renderSample(w io.Writer, s *stack.Sample) error {
	_, err := fmt.Fprintf(w, "%s %v\n", strings.Join(s.Funcs, ";"), s.Count)
	return err
}
