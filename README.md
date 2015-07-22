# go-torch [![Build Status](https://travis-ci.org/uber/go-torch.svg?branch=master)](https://travis-ci.org/uber/go-torch) [![GoDoc](https://godoc.org/github.com/uber/go-torch?status.svg)](https://godoc.org/github.com/uber/go-torch)

## Synopsis

Tool for stochastically profiling Go programs. Collects stack traces and
synthesizes them into into a flame graph. Uses Go's built in pprof library.

## Example Flame Graph

![Inception](http://uber.github.io/go-torch/meta.svg)

## Basic Usage

```
$ go-torch --help

NAME:
   go-torch - go-torch collects stack traces of a Go application and synthesizes them into into a [flame graph](http://www.brendangregg.com/FlameGraphs/cpuflamegraphs.html)

USAGE:
   go-torch [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --url, -u "http://localhost:8080"   base url of your Go program
   --suffix, -s "/debug/pprof/profile" url path of pprof profile
   --binaryinput, -b          file path of raw binary profile; alternative to having go-torch query pprof endpoint (binary profile is anything accepted by https://golang.org/cmd/pprof)
   --time, -t "30"         time in seconds to profile for
   --file, -f "torch.svg"     ouput file name (must be .svg)
   --print, -p          print the generated svg to stdout instead of writing to file
   --raw, -r            print the raw call graph output to stdout instead of creating a flame graph; use with Brendan Gregg's flame graph perl script (see https://github.com/brendangregg/FlameGraph)
   --help, -h           show help
   --version, -v        print the version

```

### File Example

```
$ go-torch --time=15 --file "torch.svg" --url http://localhost:8080
INFO[0000] Profiling ...
INFO[0015] flame graph has been created as torch.svg
```

### Stdout Example

```
$ go-torch --time=15 --print --url http://localhost:8080
INFO[0000] Profiling ...
<svg>
...
</svg>
INFO[0015] flame graph has been printed to stdout
```

### Raw Example

```
$ go-torch --time=15 --raw --url http://localhost:8080
INFO[0000] Profiling ...
function1;function2 3
...
INFO[0015] raw call graph output been printed to stdout
```

## Installation

```
$ go get github.com/uber/go-torch
```

### Install the Go dependencies:

```
$ go get github.com/tools/godep
$ godep restore
```

### Get the flame graph script:

```
$ git clone https://github.com/brendangregg/FlameGraph.git
```

## Integrating With Your Application

Expose a pprof endpoint. Official Go docs are
[here](https://golang.org/pkg/net/http/pprof/). If your application is already
running a server on the DefaultServeMux, just add this import to your
application.

```go
import _ "net/http/pprof"
```

If your application is not using the DefaultServeMux, you can still easily expose pprof endpoints by manually registering the net/http/pprof handlers or by using a library like [this one](https://github.com/e-dard/netbug).

## Run the Tests

```
$ go test ./...
ok    github.com/uber/go-torch   0.012s
ok    github.com/uber/go-torch/graph   0.017s
ok    github.com/uber/go-torch/visualization 0.052s
```
