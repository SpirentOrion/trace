package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/SpirentOrion/trace"
	"github.com/SpirentOrion/trace/yamlrec"
)

func ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Spans that run as goroutines must be started with trace.Go in order to maintain trace/span context
	ts, _ := trace.Continue("work", "sub1")
	trace.Go(ts, sub2)

	// Spans that run normally are started with trace.Run
	ts, _ = trace.Continue("work", "sub2")
	trace.Run(ts, sub3)
}

func sub2() {
	time.Sleep(5 * time.Second)
}

func sub3() {
	time.Sleep(1 * time.Second)
}

func main() {
	// Record traces to example.yaml
	f, err := os.OpenFile("example.yaml", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	rec, err := yamlrec.New(f)
	if err != nil {
		panic(err)
	}

	err = trace.Record(rec, 100, log.New(os.Stderr, "[trace] ", log.LstdFlags))
	if err != nil {
		panic(err)
	}

	// Run a simple HTTP server to generate traces
	h := trace.NewHandler()
	fmt.Println("Listening on :8000")
	http.ListenAndServe(":8000", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		h.ServeHTTP(rw, req, ServeHTTP)
	}))
}
