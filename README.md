# Trace package [![GoDoc](https://godoc.org/github.com/SpirentOrion/trace?status.svg)](http://godoc.org/github.com/SpirentOrion/trace)

Trace is a [golang][golang] package that implements a distributed
tracing capability inspired by Google's [Dapper][dapper].  Traces may
be optionally recorded to a database backend or local file.  HTTP
[middleware][middleware] is provided to facilitate easy tracing of
requests and cascading of trace contexts during request fanout
scenarios.

[golang]: http://golang.org/
[dapper]: http://research.google.com/pubs/pub36356.html
[middleware]: http://codegangsta.gitbooks.io/building-web-apps-with-go/content/middleware/README.html

To install the package:

    $ go get github.com/SpirentOrion/trace

## Tracing Data Model

Each trace is logically comprised of one or more spans in a tree-like
structure.  You are free to determine the granularity of traces and
spans.  Typical usages include:

* Each request received at a web API endpoint starts a new trace.
* Each scheduled background task starts a new trace.
* Each span represents some type of start-to-finish activity.  By
  creating new spans you can differentiate between different types or
  stages of activity within a single trace.

Traces are identified by a probabilistically unique 64-bit integer.
All spans within in a trace share the same trace id.  Identifiers are
randomly generated within this number space and thus do not require
use of a centralized allocator.

Spans are also identified by their own unique 64-bit integer values.
Each span records its trace id, the id of its parent span, its
location, its start time, its finish time, and other data that you may
provide.

With this structure it is possible to build a causal record of
activity within trace.  For any trace, activity began with the first
span -- the span with a parent id of 0.  Any spans within the same
trace that have a parent id matching the first span's id were caused
by the first span.  And so on.

If processing of a span requires fanout to other services or processes
the trace context may be propagated using HTTP request headers or
other appropriate mechanisms.  Each span's location indicates where
the activity actually occurred, in terms of both process and hostname.
When activity spans multiple hosts, start and finish times are based
on the host's clock and aren't necessarily synchronized.

## Recording Backends

Currently, two recording backend packages are provided:

| Backend | Import Path |
| :-- | :-- |
| DynamoDB | `github.com/SpirentOrion/trace/dynamorec` |
| YAML | `github.com/SpirentOrion/trace/yamlrec` |

## Example

A simple [example](https://github.com/SpirentOrion/trace/blob/master/example/main.go)
is provided with trace recording via the YAML recorder:

    $ cd $GOPATH/src/github.com/SpirentOrion/trace/example
    $ go run main.go

Separately:

    $ curl -i http://127.0.0.1/foo/bar
    $ cat example.yaml

Note that the YAML recorder only records finished spans.  Each span is
rendered as a separate document in the YAML stream.
