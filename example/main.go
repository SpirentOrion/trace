package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/SpirentOrion/trace"
)

func ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Spans may be recorded directly in a single goroutine
	trace.Do(req.Context(), "work", "sub2", sub2)

	// Or, spans may run as new goroutines
	trace.Do(req.Context(), "work", "sub3", func(ctx context.Context) {
		go sub3(ctx)
	})
}

func sub2(ctx context.Context) {
	time.Sleep(5 * time.Second)
}

func sub3(ctx context.Context) {
	time.Sleep(1 * time.Second)
}

func main() {
	// Record traces to example.json
	f, err := os.OpenFile("example.json", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	var rec trace.Recorder
	if rec, err = trace.NewJSONRecorder(f); err != nil {
		panic(err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer func() {
		cancelFunc()
	}()

	if ctx, err = trace.Record(ctx, rec); err != nil {
		panic(err)
	}

	// Run a simple HTTP server to generate traces
	h := trace.NewHandler(ctx)
	fmt.Println("Listening on :8000")
	http.ListenAndServe(":8000", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		h.ServeHTTP(rw, req, ServeHTTP)
	}))
}
