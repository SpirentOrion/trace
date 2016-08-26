package trace

import (
	"context"
	"fmt"
	"os"
)

type contextKeyT int

const (
	contextProcessKey  = contextKeyT(0)
	contextBufferKey   = contextKeyT(1)
	contextLoggerKey   = contextKeyT(2)
	contextTraceIDKey  = contextKeyT(3)
	contextParentIDKey = contextKeyT(4)
	contextSpansKey    = contextKeyT(5)
	contextSpanKey     = contextKeyT(6)
)

var defaultProcess string

func init() {
	hostname, _ := os.Hostname()
	defaultProcess = fmt.Sprintf("%s:%d@%s", os.Args[0], os.Getpid(), hostname) // "argv[0]:pid@hostname"
}

// WithProcess returns a new context.Context instance with a process identifier
// included as a value. This process identifier is used in all trace spans
// created from the new context.
func WithProcess(ctx context.Context, process string) context.Context {
	return context.WithValue(ctx, contextProcessKey, process)
}

func contextProcess(ctx context.Context) (process string) {
	if process, _ = ctx.Value(contextProcessKey).(string); process == "" {
		process = defaultProcess
	}
	return
}

// WithBuffer returns a new context.Context instance with a buffer depth
// included as a value. This buffer value is used when allocating internal
// channels for trace span recording.
func WithBuffer(ctx context.Context, buffer int) context.Context {
	return context.WithValue(ctx, contextBufferKey, buffer)
}

func contextBuffer(ctx context.Context) (buffer int) {
	buffer, _ = ctx.Value(contextBufferKey).(int)
	return
}

// WithLogger returns a new context.Context instance with a Logger as a value.
// This logger is used to log errors that occur during tracing.
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, contextLoggerKey, logger)
}

func contextLogger(ctx context.Context) (logger Logger) {
	logger, _ = ctx.Value(contextLoggerKey).(Logger)
	return
}

// WithTraceID returns a new context.Context instance with a trace id included
// as a value. This may be used to override automatic trace id generation when
// new trace spans are created.
func WithTraceID(ctx context.Context, traceID int64) context.Context {
	return context.WithValue(ctx, contextTraceIDKey, traceID)
}

func contextTraceID(ctx context.Context) (traceID int64) {
	traceID, _ = ctx.Value(contextTraceIDKey).(int64)
	return
}

// WithParentID returns a new context.Context instance with a parent id included
// as a value. This may be used to override parent id determination when new
// trace spans are created.
func WithParentID(ctx context.Context, parentID int64) context.Context {
	return context.WithValue(ctx, contextParentIDKey, parentID)
}

func contextParentID(ctx context.Context) (parentID int64) {
	parentID, _ = ctx.Value(contextParentIDKey).(int64)
	return
}

// WithContext returns a new context.Context instance merging trace and parent
// id values from a context.Context where tracing is active.
func WithContext(ctx context.Context, traceCtx context.Context) context.Context {
	traceID, _ := traceCtx.Value(contextTraceIDKey).(int64)
	parentID, _ := traceCtx.Value(contextParentIDKey).(int64)
	if traceID > 0 && parentID > 0 {
		return WithTraceID(WithParentID(ctx, parentID), traceID)
	} else {
		return ctx
	}
}

func withSpans(ctx context.Context, spans chan *Span) context.Context {
	return context.WithValue(ctx, contextSpansKey, spans)
}

func contextSpans(ctx context.Context) (spans chan *Span) {
	spans, _ = ctx.Value(contextSpansKey).(chan *Span)
	return
}

func withSpan(ctx context.Context, span *Span) context.Context {
	return context.WithValue(ctx, contextSpanKey, span)
}

func contextSpan(ctx context.Context) (span *Span) {
	span, _ = ctx.Value(contextSpanKey).(*Span)
	return
}
