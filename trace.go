package trace

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"
)

const defaultBuffer = 256

var (
	errMissingContext  = errors.New("context is required to start trace recording")
	errMissingRecorder = errors.New("recorder is required to start trace recording")
)

// Logger is an interface compatible with log.Logger.
type Logger interface {
	Println(v ...interface{})
}

// Record starts trace recording in a context. If the context is canceled then
// trace recording stops.
func Record(ctx context.Context, rec Recorder) (context.Context, error) {
	if ctx == nil {
		return nil, errMissingContext
	}
	if rec == nil {
		return nil, errMissingRecorder
	}

	buffer := contextBuffer(ctx)
	if buffer <= 0 {
		buffer = defaultBuffer
	}

	spans := make(chan *Span, buffer)
	ctx = withSpans(ctx, spans)

	go record(ctx, rec)
	return ctx, nil
}

func record(ctx context.Context, rec Recorder) {
	spans, logger := contextSpans(ctx), contextLogger(ctx)
	for {
		select {
		case span := <-spans:
			span.StartStr = span.Start.Format(time.RFC3339Nano)
			span.FinishStr = span.Finish.Format(time.RFC3339Nano)
			if err := rec.Record(span); err != nil {
				if logger != nil {
					msg := fmt.Sprintf("trace: failed to record trace %x span %x: %s", span.TraceID, span.SpanID, err)
					logger.Println(msg)
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

// Do starts a new trace span if recording is active in the context. If a trace
// is already active in ctx then the new trace span continues under the existing
// trace id, otherwise a new trace id is generated.
//
// The activity function act is always invoked, either with a new context
// representing a new trace span, or with the caller-provided ctx if recording
// is not active or an error occurs.
func Do(ctx context.Context, kind string, name string, act func(context.Context)) {
	// If recording is not active, we are done right away
	var spans chan *Span
	if ctx != nil {
		spans = contextSpans(ctx)
	}
	if spans == nil {
		act(ctx)
		return
	}

	// Generate a new span id
	spanID, err := GenerateID(ctx)
	if err != nil {
		act(ctx)
		return
	}

	// If a a trace is already active then the new span continues under the existing
	var traceID, parentID int64
	if span := contextSpan(ctx); span != nil {
		traceID = span.TraceID
		parentID = span.SpanID
	} else {
		if traceID = contextTraceID(ctx); traceID == 0 {
			if traceID, err = GenerateID(ctx); err != nil {
				act(ctx)
				return
			}
		}
		parentID = contextParentID(ctx)
	}

	// Allocate a new span and ensure that it is recorded even if a panic occurs
	span := &Span{
		SpanID:   spanID,
		TraceID:  traceID,
		ParentID: parentID,
		Process:  contextProcess(ctx),
		Kind:     kind,
		Name:     name,
		Start:    time.Now(),
	}
	defer func() {
		span.Finish = time.Now()
		spans <- span
	}()

	// Perform the activity with a new context
	act(withSpan(ctx, span))
}

// CurrentSpanID returns the current span id if a trace is active in the
// context. Otherwise it returns 0.
func CurrentSpanID(ctx context.Context) (spanID int64) {
	if ctx != nil {
		if span := contextSpan(ctx); span != nil {
			spanID = span.SpanID
		}
	}
	return
}

// CurrentTraceID returns the current trace id if a trace is active in the
// context. Otherwise it returns 0.
func CurrentTraceID(ctx context.Context) (traceID int64) {
	if ctx != nil {
		if span := contextSpan(ctx); span != nil {
			traceID = span.TraceID
		}
	}
	return
}

// Annotate returns a map that can be used to store trace span-specific data if
// a trace is active in ctx. Otherwise it returns nil.
func Annotate(ctx context.Context) map[string]interface{} {
	if span := contextSpan(ctx); span != nil {
		if span.Data == nil {
			span.Data = make(map[string]interface{})
		}
		return span.Data
	}
	return nil
}

// GenerateId returns a probablistically unique 64-bit id if a trace is active
// in ctx. All id values returned by this function will be positive integers.
// This may be useful for callers that want to generate their own id values.
func GenerateID(ctx context.Context) (int64, error) {
	for {
		// Return a random int64, constrained to positive values
		var x uint64
		if err := binary.Read(rand.Reader, binary.LittleEndian, &x); err != nil {
			if logger := contextLogger(ctx); logger != nil {
				msg := fmt.Sprintf("trace: error reading from rand.Reader: %s", err)
				logger.Println(msg)
			}
			return 0, err
		}

		if id := int64(x & math.MaxInt64); id > 0 {
			return id, nil
		}
	}
}
