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
	"bytes"
	"testing"

	ggv "github.com/awalterschulze/gographviz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPathAsString(t *testing.T) {
	g := testGraphWithTooltipAndWeight()

	eMap := g.Edges.SrcToDsts
	path := []ggv.Edge{*eMap["N1"]["N2"], *eMap["N2"]["N3"], *eMap["N3"]["N4"]}

	pathString := new(defaultPathStringer).pathAsString(path, g.Nodes.Lookup)

	assert.Equal(t, "function1;function2;function3;function4 9\n", pathString)
}

func TestPathAsStringWithEmptyPath(t *testing.T) {
	path := []ggv.Edge{}

	pathString := new(defaultPathStringer).pathAsString(path, map[string]*ggv.Node{})
	assert.Equal(t, "0\n", pathString)
}

func TestPathAsStringWithNoWeightEdges(t *testing.T) {
	g := testGraphWithTooltip()

	eMap := g.Edges.SrcToDsts
	path := []ggv.Edge{*eMap["N1"]["N2"], *eMap["N2"]["N3"], *eMap["N3"]["N4"]}

	pathString := new(defaultPathStringer).pathAsString(path, g.Nodes.Lookup)

	assert.Equal(t, "function1;function2;function3;function4 0\n", pathString)
}

func TestDFS(t *testing.T) {
	g := testSingleRootGraph()
	eMap := g.Edges.SrcToDsts

	nodeToOutEdges := map[string][]*ggv.Edge{
		"N1": {eMap["N1"]["N2"], eMap["N1"]["N3"], eMap["N1"]["N4"]},
		"N2": {eMap["N2"]["N3"]},
		"N4": {eMap["N4"]["N3"]},
	}

	buffer := new(bytes.Buffer)
	mockPathStringer := new(mockPathStringer)
	anythingType := mock.AnythingOfType("map[string]*gographviz.Node")
	pathOne := []ggv.Edge{*eMap["N1"]["N2"], *eMap["N2"]["N3"]}
	pathTwo := []ggv.Edge{*eMap["N1"]["N3"]}
	pathThree := []ggv.Edge{*eMap["N1"]["N4"], *eMap["N4"]["N3"]}

	mockPathStringer.On("pathAsString", pathOne, anythingType).Return("N1;N2;N3 3\n").Once()
	mockPathStringer.On("pathAsString", pathTwo, anythingType).Return("N1;N3 2\n").Once()
	mockPathStringer.On("pathAsString", pathThree, anythingType).Return("N1;N4;N3 8\n").Once()

	searcherWithTestStringer := &defaultSearcher{
		pathStringer: mockPathStringer,
	}
	searcherWithTestStringer.dfs(searchArgs{
		root:           "N1",
		path:           []ggv.Edge{},
		nodeToOutEdges: nodeToOutEdges,
		nameToNodes:    g.Nodes.Lookup,
		buffer:         buffer,
	})

	correctOutput := "N1;N2;N3 3\nN1;N3 2\nN1;N4;N3 8\n"
	actualOutput := buffer.String()

	assert.Equal(t, correctOutput, actualOutput)
	mockPathStringer.AssertExpectations(t)
}

func TestDFSAlmostEmptyGraph(t *testing.T) {
	g := ggv.NewGraph()
	g.SetName("G")
	g.AddNode("G", "N1", nil)
	g.SetDir(true)

	nodeToOutEdges := map[string][]*ggv.Edge{}
	buffer := new(bytes.Buffer)

	mockPathStringer := new(mockPathStringer)
	anythingType := mock.AnythingOfType("map[string]*gographviz.Node")

	mockPathStringer.On("pathAsString", []ggv.Edge{}, anythingType).Return("").Once()

	searcherWithTestStringer := &defaultSearcher{
		pathStringer: mockPathStringer,
	}
	searcherWithTestStringer.dfs(searchArgs{
		root:           "N1",
		path:           []ggv.Edge{},
		nodeToOutEdges: nodeToOutEdges,
		nameToNodes:    g.Nodes.Lookup,
		buffer:         buffer,
	})

	correctOutput := ""
	actualOutput := buffer.String()

	assert.Equal(t, correctOutput, actualOutput)
	mockPathStringer.AssertExpectations(t)
}

func TestDFSMultipleRootsLeaves(t *testing.T) {
	g := testMultiRootGraph()

	eMap := g.Edges.SrcToDsts

	nodeToOutEdges := map[string][]*ggv.Edge{
		"N1": {eMap["N1"]["N2"], eMap["N1"]["N3"]},
		"N4": {eMap["N4"]["N5"], eMap["N4"]["N6"]},
		"N6": {eMap["N6"]["N5"]},
	}

	buffer := new(bytes.Buffer)
	mockPathStringer := new(mockPathStringer)
	anythingType := mock.AnythingOfType("map[string]*gographviz.Node")
	pathOne := []ggv.Edge{*eMap["N1"]["N2"]}
	pathTwo := []ggv.Edge{*eMap["N1"]["N3"]}
	pathThree := []ggv.Edge{*eMap["N4"]["N5"]}
	pathFour := []ggv.Edge{*eMap["N4"]["N6"], *eMap["N6"]["N5"]}

	mockPathStringer.On("pathAsString", pathOne, anythingType).Return("N1;N2 3\n").Once()
	mockPathStringer.On("pathAsString", pathTwo, anythingType).Return("N1;N3 2\n").Once()
	mockPathStringer.On("pathAsString", pathThree, anythingType).Return("N4;N5 8\n").Once()
	mockPathStringer.On("pathAsString", pathFour, anythingType).Return("N4;N6;N5 7\n").Once()

	searcherWithTestStringer := &defaultSearcher{
		pathStringer: mockPathStringer,
	}

	searcherWithTestStringer.dfs(searchArgs{
		root:           "N1",
		path:           []ggv.Edge{},
		nodeToOutEdges: nodeToOutEdges,
		nameToNodes:    g.Nodes.Lookup,
		buffer:         buffer,
	})
	searcherWithTestStringer.dfs(searchArgs{
		root:           "N4",
		path:           []ggv.Edge{},
		nodeToOutEdges: nodeToOutEdges,
		nameToNodes:    g.Nodes.Lookup,
		buffer:         buffer,
	})

	correctOutput := "N1;N2 3\nN1;N3 2\nN4;N5 8\nN4;N6;N5 7\n"
	actualOutput := buffer.String()

	assert.Equal(t, correctOutput, actualOutput)
	mockPathStringer.AssertExpectations(t)
}

func TestGetInDegreeZeroNodes(t *testing.T) {
	g := testMultiRootGraph()

	correctInDegreeZeroNodes := []string{"N1", "N4"}
	actualInDegreeZeroNodes := new(defaultCollectionGetter).getInDegreeZeroNodes(g)
	assert.Equal(t, correctInDegreeZeroNodes, actualInDegreeZeroNodes)
}

func TestGetInDegreeZeroNodesEmptyGraph(t *testing.T) {
	g := ggv.NewGraph()
	g.SetName("G")
	g.SetDir(true)

	var correctInDegreeZeroNodes []string
	actualInDegreeZeroNodes := new(defaultCollectionGetter).getInDegreeZeroNodes(g)
	assert.Equal(t, correctInDegreeZeroNodes, actualInDegreeZeroNodes)
}

func TestGetInDegreeZeroNodesIgnoreClusterNodes(t *testing.T) {
	g := testGraphWithClusterNodes()

	correctInDegreeZeroNodes := []string{"N1"}
	actualInDegreeZeroNodes := new(defaultCollectionGetter).getInDegreeZeroNodes(g)
	assert.Equal(t, correctInDegreeZeroNodes, actualInDegreeZeroNodes)
}

func TestGenerateNodeToOutEdges(t *testing.T) {
	g := testMultiRootGraph()

	eMap := g.Edges.SrcToDsts

	correctNodeToOutEdges := map[string][]*ggv.Edge{
		"N1": {eMap["N1"]["N2"], eMap["N1"]["N3"]},
		"N4": {eMap["N4"]["N5"], eMap["N4"]["N6"]},
		"N6": {eMap["N6"]["N5"]},
	}
	actualNodeToOutEdges := new(defaultCollectionGetter).generateNodeToOutEdges(g)
	assert.Equal(t, correctNodeToOutEdges, actualNodeToOutEdges)
}

func TestGenerateNodeToOutEdgesEmptyGraph(t *testing.T) {
	g := ggv.NewGraph()
	g.SetName("G")
	g.SetDir(true)

	correctNodeToOutEdges := make(map[string][]*ggv.Edge)
	actualNodeToOutEdges := new(defaultCollectionGetter).generateNodeToOutEdges(g)
	assert.Equal(t, correctNodeToOutEdges, actualNodeToOutEdges)
}

func TestGraphAsText(t *testing.T) {
	mockSearcher := new(mockSearcher)
	mockCollectionGetter := new(mockCollectionGetter)
	grapher := &defaultGrapher{
		searcher:         mockSearcher,
		collectionGetter: mockCollectionGetter,
	}

	graphAsTextInput := []byte(`digraph "unnamed" {
		node [style=filled fillcolor="#f8f8f8"]
		N1 [tooltip="N1"]
		N2 [tooltip="N2"]
		N3 [tooltip="N3"]
		N4 [tooltip="N4"]
		N5 [tooltip="N5"]
		N6 [tooltip="N6"]
		N1 -> N2 [weight=1]
		N1 -> N3 [weight=2]
		N4 -> N5 [weight=1]
		N4 -> N6 [weight=4]
		N6 -> N5 [weight=4]
		}`)

	fakeWriteToBuffer := func(args mock.Arguments) {
		searchArgs := args.Get(0).(searchArgs)
		if searchArgs.root == "N1" {
			searchArgs.buffer.WriteString("N1;N2 1\nN1;N3 2\n")
		} else {
			searchArgs.buffer.WriteString("N4;N5 1\nN4;N6;N5 8\n")
		}
	}

	mockSearcher.On("dfs", mock.AnythingOfType("searchArgs")).Return().Run(fakeWriteToBuffer).Twice()
	mockCollectionGetter.On("generateNodeToOutEdges",
		mock.AnythingOfType("*gographviz.Graph")).Return(nil).Once() // We can return nil since the mock dfs will ignore this
	mockCollectionGetter.On("getInDegreeZeroNodes",
		mock.AnythingOfType("*gographviz.Graph")).Return([]string{"N1", "N4"}).Once()

	correctGraphAsText := "N1;N2 1\nN1;N3 2\nN4;N5 1\nN4;N6;N5 8\n"

	actualGraphAsText, err := grapher.GraphAsText(graphAsTextInput)
	assert.NoError(t, err)
	assert.Equal(t, correctGraphAsText, actualGraphAsText)
	mockSearcher.AssertExpectations(t)
}

// The returned graph, represented in ascii:
//	+----+     +----+
//	| N2 | <-- | N1 |
//	+----+     +----+
//	             |
//	             |
//	             v
//	           +----+
//	           | N3 |
//	           +----+
//	           +----+
//	           | N4 | -+
//	           +----+  |
//	             |     |
//	             |     |
//	             v     |
//	           +----+  |
//	           | N6 |  |
//	           +----+  |
//	             |     |
//	             |     |
//	             v     |
//	           +----+  |
//	           | N5 | <+
//	           +----+
func testMultiRootGraph() *ggv.Graph {
	g := ggv.NewGraph()
	g.SetName("G")
	g.SetDir(true)
	g.AddNode("G", "N1", nil)
	g.AddNode("G", "N2", nil)
	g.AddNode("G", "N3", nil)
	g.AddNode("G", "N4", nil)
	g.AddNode("G", "N5", nil)
	g.AddNode("G", "N6", nil)
	g.AddEdge("N1", "N2", true, nil)
	g.AddEdge("N1", "N3", true, nil)
	g.AddEdge("N4", "N5", true, nil)
	g.AddEdge("N4", "N6", true, nil)
	g.AddEdge("N6", "N5", true, nil)
	return g
}

// The returned graph, represented in ascii:
//	+----+
//	| N1 | -+
//	+----+  |
//	  |     |
//	  |     |
//	  v     |
//	+----+  |
//	| N2 |  |
//	+----+  |
//	  |     |
//	  |     |
//	  v     |
//	+----+  |
//	| N3 |  |
//	+----+  |
//	  |     |
//	  |     |
//	  v     |
//	+----+  |
//	| N4 | <+
//	+----+
func testGraphWithTooltipAndWeight() *ggv.Graph {
	g := ggv.NewGraph()
	g.SetName("G")
	g.SetDir(true)
	g.AddNode("G", "N1", map[string]string{"tooltip": "function1"})
	g.AddNode("G", "N2", map[string]string{"tooltip": "function2"})
	g.AddNode("G", "N3", map[string]string{"tooltip": "function3"})
	g.AddNode("G", "N4", map[string]string{"tooltip": "function4"})
	g.AddEdge("N1", "N2", true, map[string]string{"weight": "5"})
	g.AddEdge("N2", "N3", true, map[string]string{"weight": "2"})
	g.AddEdge("N3", "N4", true, map[string]string{"weight": "2"})
	g.AddEdge("N1", "N4", true, map[string]string{"weight": "1"})
	return g
}

// The returned graph, represented in ascii:
//	+----+
//	| N1 | -+
//	+----+  |
//	  |     |
//	  |     |
//	  v     |
//	+----+  |
//	| N2 |  |
//	+----+  |
//	  |     |
//	  |     |
//	  v     |
//	+----+  |
//	| N3 |  |
//	+----+  |
//	  |     |
//	  |     |
//	  v     |
//	+----+  |
//	| N4 | <+
//	+----+
func testGraphWithTooltip() *ggv.Graph {
	g := ggv.NewGraph()
	g.SetName("G")
	g.SetDir(true)
	g.AddNode("G", "N1", map[string]string{"tooltip": "function1"})
	g.AddNode("G", "N2", map[string]string{"tooltip": "function2"})
	g.AddNode("G", "N3", map[string]string{"tooltip": "function3"})
	g.AddNode("G", "N4", map[string]string{"tooltip": "function4"})
	g.AddEdge("N1", "N2", true, nil)
	g.AddEdge("N2", "N3", true, nil)
	g.AddEdge("N3", "N4", true, nil)
	g.AddEdge("N1", "N4", true, nil)
	return g
}

// The returned graph, represented in ascii:
//	+----+     +----+
//	| N4 | <-- | N1 | -+
//	+----+     +----+  |
//	  |          |     |
//	  |          |     |
//	  |          v     |
//	  |        +----+  |
//	  |        | N2 |  |
//	  |        +----+  |
//	  |          |     |
//	  |          |     |
//	  |          v     |
//	  |        +----+  |
//	  +------> | N3 | <+
//	           +----+
func testSingleRootGraph() *ggv.Graph {
	g := ggv.NewGraph()
	g.SetName("G")
	g.SetDir(true)
	g.AddNode("G", "N1", nil)
	g.AddNode("G", "N2", nil)
	g.AddNode("G", "N3", nil)
	g.AddNode("G", "N4", nil)
	g.AddEdge("N1", "N2", true, nil)
	g.AddEdge("N2", "N3", true, nil)
	g.AddEdge("N4", "N3", true, nil)
	g.AddEdge("N1", "N4", true, nil)
	g.AddEdge("N1", "N3", true, nil)
	return g
}

// The returned graph, represented in ascii:
//	           +----------+
//	           | Ignoreme |
//	           +----------+
//	+----+     +----------+
//	| N2 | <-- |    N1    |
//	+----+     +----------+
//	             |
//	             |
//	             v
//	           +----------+
//	           |    N3    |
//  	       +----------+
func testGraphWithClusterNodes() *ggv.Graph {
	g := ggv.NewGraph()
	g.SetName("G")
	g.SetDir(true)
	g.AddNode("G", "N1", nil)
	g.AddNode("G", "N2", nil)
	g.AddNode("G", "N3", nil)
	g.AddNode("G", "Ignore me!", nil)
	g.AddEdge("N1", "N2", true, nil)
	g.AddEdge("N1", "N3", true, nil)
	return g
}
