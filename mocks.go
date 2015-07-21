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
	"github.com/codegangsta/cli"
	"github.com/stretchr/testify/mock"
)

type mockVisualizer struct {
	mock.Mock
}

func (m *mockVisualizer) GenerateFlameGraph(_a0 string, _a1 string, _a2 bool) error {
	ret := m.Called(_a0, _a1, _a2)

	r0 := ret.Error(0)

	return r0
}

type mockGrapher struct {
	mock.Mock
}

func (m *mockGrapher) GraphAsText(_a0 []byte) (string, error) {
	ret := m.Called(_a0)

	r0 := ret.Get(0).(string)
	r1 := ret.Error(1)

	return r0, r1
}

type mockCommander struct {
	mock.Mock
}

func (m *mockCommander) goTorchCommand(_a0 *cli.Context) {
	m.Called(_a0)
}

type mockValidator struct {
	mock.Mock
}

func (m *mockValidator) validateArgument(_a0 string, _a1 string, _a2 string) error {
	ret := m.Called(_a0, _a1, _a2)

	r0 := ret.Error(0)

	return r0
}

type mockPprofer struct {
	mock.Mock
}

func (m *mockPprofer) runPprofCommand(_a0 int, _a1 string) ([]byte, error) {
	ret := m.Called(_a0, _a1)

	var r0 []byte
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]byte)
	}
	r1 := ret.Error(1)

	return r0, r1
}
