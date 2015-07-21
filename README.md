# go-torch

## Synopsis

Tool for stochastically profiling Go programs. Collects stack traces and synthesizes them into into a flame graph. Uses Go's built in pprof library.

## Basic Usage

```
[go-torch]$ go-torch --help

NAME:
   go-torch - go-torch collects stack traces of a Go application and synthesizes them into into a flame graph

USAGE:
   go-torch [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --url, -u "http://localhost:8080"   base url of your go program
   --suffix, -s "/debug/pprof/profile" url path of pprof profile
   --time, -t "30"         time in seconds to profile for
   --file, -f "torch.svg"     ouput file name (must be .svg)
   --print, -p          print the generated svg to stdout instead of writing to file
   --help, -h           show help
   --version, -v        print the version

```

## File Example

```
$ go-torch --time=15 --file "torch.svg" --url http://localhost:8080
INFO[0000] Profiling ...
INFO[0015] flame graph has been created as torch.svg
```

## Stdout Example

```
$ go-torch --time=15 --print --url http://localhost:8080
INFO[0000] Profiling ...
<svg>
...
</svg>
```

## Installation

```
$ go get github.com/uber/go-torch
```

Install the dependencies:

```
$ go get github.com/tools/godep
$ godep restore
```

## Integrating With Your Application

Expose a pprof endpoint. Official Go docs are [here](https://golang.org/pkg/net/http/pprof/). If your application is already running a server on the DefaultServeMux, just add this import to your application.

```
import _ "net/http/pprof"
```

If your application is not using the DefaultServeMux, you can still easily expose pprof endpoints.

## Run the Tests

```
$ go test ./...
ok    github.com/uber/go-torch   0.010s
ok    github.com/uber/go-torch/graph   0.014s
ok    github.com/uber/go-torch/visualization 0.130s
```
