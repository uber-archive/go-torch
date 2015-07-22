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

// Package graph transforms a DOT graph text file into the representation
// expected by the visualization package.
//
// The graph is a directed acyclic graph where nodes represent functions and
// directed edges represent how many times a function calls another.
package graph

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	ggv "github.com/awalterschulze/gographviz"
	"github.com/awalterschulze/gographviz/parser"
)

var errNoActivity = errors.New("Your application is not doing anything right now. Please try again.")

// Grapher handles transforming a DOT graph byte array into the
// representation expected by the visualization package.
type Grapher interface {
	GraphAsText([]byte) (string, error)
}

type defaultGrapher struct {
	searcher         searcher
	collectionGetter collectionGetter
}

type searchArgs struct {
	root           string
	path           []ggv.Edge
	nodeToOutEdges map[string][]*ggv.Edge
	nameToNodes    map[string]*ggv.Node
	buffer         *bytes.Buffer
	colorMap       map[string]color
}

type searcher interface {
	dfs(args searchArgs)
}

type defaultSearcher struct {
	pathStringer pathStringer
}

type collectionGetter interface {
	generateNodeToOutEdges(*ggv.Graph) map[string][]*ggv.Edge
	getInDegreeZeroNodes(*ggv.Graph) []string
}

type defaultCollectionGetter struct{}

type pathStringer interface {
	pathAsString([]ggv.Edge, map[string]*ggv.Node) string
}

type defaultPathStringer struct{}

// Marking colors during depth-first search is a standard way of detecting cycles.
// A node is white before it has been discovered, gray when it is on the recursion stack, and black
// when all of its neighbors have been traversed. A edge terminating at a grey edge implies a back
// edge, which also implies a cycle
// (see: https://en.wikipedia.org/wiki/Cycle_(graph_theory)#Cycle_detection).
type color int

const (
	WHITE color = iota
	GRAY
	BLACK
)

// NewGrapher returns a default grapher struct with default attributes
func NewGrapher() Grapher {
	return &defaultGrapher{
		searcher:         newSearcher(),
		collectionGetter: new(defaultCollectionGetter),
	}
}

// newSearcher returns a default searcher struct with a default pathStringer
func newSearcher() *defaultSearcher {
	return &defaultSearcher{
		pathStringer: new(defaultPathStringer),
	}
}

// GraphAsText is the standard implementation of Grapher
func (g *defaultGrapher) GraphAsText(dotText []byte) (string, error) {
	graphAst, err := parser.ParseBytes(dotText)
	if err != nil {
		return "", err
	}
	dag := ggv.NewGraph() // A directed acyclic graph
	ggv.Analyse(graphAst, dag)

	if len(dag.Edges.Edges) == 0 {
		return "", errNoActivity
	}
	nodeToOutEdges := g.collectionGetter.generateNodeToOutEdges(dag)
	inDegreeZeroNodes := g.collectionGetter.getInDegreeZeroNodes(dag)
	nameToNodes := dag.Nodes.Lookup

	buffer := new(bytes.Buffer)
	colorMap := make(map[string]color)

	for _, root := range inDegreeZeroNodes {
		g.searcher.dfs(searchArgs{
			root:           root,
			path:           nil,
			nodeToOutEdges: nodeToOutEdges,
			nameToNodes:    nameToNodes,
			buffer:         buffer,
			colorMap:       colorMap,
		})
	}

	return buffer.String(), nil
}

// generateNodeToOutEdges takes a graph and generates a mapping of nodes to
// edges originating from nodes.
func (c *defaultCollectionGetter) generateNodeToOutEdges(dag *ggv.Graph) map[string][]*ggv.Edge {
	nodeToOutEdges := make(map[string][]*ggv.Edge)
	for _, edge := range dag.Edges.Edges {
		nodeToOutEdges[edge.Src] = append(nodeToOutEdges[edge.Src], edge)
	}
	return nodeToOutEdges
}

// getInDegreeZeroNodes takes a graph and returns a list of nodes with
// in-degree of 0. In other words, no edges terminate at these nodes.
func (c *defaultCollectionGetter) getInDegreeZeroNodes(dag *ggv.Graph) []string {
	var inDegreeZeroNodes []string
	nodeToInDegree := make(map[string]int)
	for _, edge := range dag.Edges.Edges {
		dst := edge.Dst
		nodeToInDegree[dst]++
	}
	for _, node := range dag.Nodes.Nodes {
		// @HACK This is a hack to fix a bug with gographviz where a cluster
		// 'L' is being parsed as a node. This just checks that all node names
		// begin with N.
		correctPrefix := strings.HasPrefix(node.Name, "N")
		if correctPrefix && nodeToInDegree[node.Name] == 0 {
			inDegreeZeroNodes = append(inDegreeZeroNodes, node.Name)
		}
	}
	return inDegreeZeroNodes
}

// dfs performs a depth-first search traversal of the graph starting from a
// given root node. When a node with no outgoing edges is reached, the path
// taken to that node is written to a buffer.
func (s *defaultSearcher) dfs(args searchArgs) {
	outEdges := args.nodeToOutEdges[args.root]
	if args.colorMap[args.root] == GRAY {
		logrus.Warn("The input call graph contains a cycle. This can't be represented in a " +
			"flame graph, so this path will be ignored. For your record, the ignored path " +
			"is:\n" + strings.TrimSpace(s.pathStringer.pathAsString(args.path, args.nameToNodes)))
		return
	}
	if len(outEdges) == 0 {
		args.buffer.WriteString(s.pathStringer.pathAsString(args.path, args.nameToNodes))
		args.colorMap[args.root] = BLACK
		return
	}
	args.colorMap[args.root] = GRAY
	for _, edge := range outEdges {
		s.dfs(searchArgs{
			root:           edge.Dst,
			path:           append(args.path, *edge),
			nodeToOutEdges: args.nodeToOutEdges,
			nameToNodes:    args.nameToNodes,
			buffer:         args.buffer,
			colorMap:       args.colorMap,
		})
	}
	args.colorMap[args.root] = BLACK
}

// pathAsString takes a path and a mapping of node names to node structs and
// generates the string representation of the path expected by the
// visualization package.
func (p *defaultPathStringer) pathAsString(path []ggv.Edge, nameToNodes map[string]*ggv.Node) string {
	var (
		pathBuffer bytes.Buffer
		weightSum  int
	)
	for _, edge := range path {
		// If the function call represented by the edge happened very rarely,
		// the edge's weight will not be recorded. The edge's label will always
		// be recorded.
		if weightStr, ok := edge.Attrs["weight"]; ok {
			weight, err := strconv.Atoi(weightStr)
			if err != nil { // This should never happen
				logrus.Panic(err)
			}
			weightSum += weight
		}
		functionLabel := getFormattedFunctionLabel(nameToNodes[edge.Src])
		pathBuffer.WriteString(functionLabel + ";")
	}
	if len(path) >= 1 {
		lastEdge := path[len(path)-1]
		lastFunctionLabel := getFormattedFunctionLabel(nameToNodes[lastEdge.Dst])
		pathBuffer.WriteString(lastFunctionLabel + " ")
	}
	pathBuffer.WriteString(fmt.Sprint(weightSum))
	pathBuffer.WriteString("\n")

	return pathBuffer.String()
}

// getFormattedFunctionLabel takes a node and returns a formatted function
// label.
func getFormattedFunctionLabel(node *ggv.Node) string {
	label := node.Attrs["tooltip"]
	label = strings.Replace(label, `\n`, " ", -1)
	label = strings.Replace(label, `"`, "", -1)
	return label
}
