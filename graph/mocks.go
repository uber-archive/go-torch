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

package graph

import (
	ggv "github.com/awalterschulze/gographviz"
	"github.com/stretchr/testify/mock"
)

type mockSearcher struct {
	mock.Mock
}

func (m *mockSearcher) dfs(args searchArgs) {
	m.Called(args)
}

type mockCollectionGetter struct {
	mock.Mock
}

func (m *mockCollectionGetter) generateNodeToOutEdges(_a0 *ggv.Graph) map[string][]*ggv.Edge {
	ret := m.Called(_a0)

	var r0 map[string][]*ggv.Edge
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(map[string][]*ggv.Edge)
	}

	return r0
}
func (m *mockCollectionGetter) getInDegreeZeroNodes(_a0 *ggv.Graph) []string {
	ret := m.Called(_a0)

	var r0 []string
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]string)
	}

	return r0
}

type mockPathStringer struct {
	mock.Mock
}

func (m *mockPathStringer) pathAsString(_a0 []ggv.Edge, _a1 map[string]*ggv.Node) string {
	ret := m.Called(_a0, _a1)

	r0 := ret.Get(0).(string)

	return r0
}
