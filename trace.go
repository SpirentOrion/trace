package trace

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/jtolds/gls"
)

const (
	// gls.ContextManager keys
	spanIdKey  = "trace:spanid"
	traceIdKey = "trace:traceid"
)

var (
	// Process is process name used when New or Continue create new Spans.
	Process string

	// Internal state
	cm    *gls.ContextManager
	spans chan *Span

	// Errors
	errBufferRequired = errors.New("buffer must be greater than zero")
	errRecNotActive   = errors.New("trace recording isn't active")
)

func init() {
	// "argv[0]:pid@hostname"
	host, _ := os.Hostname()
	Process = fmt.Sprintf("%s:%d@%s", os.Args[0], os.Getpid(), host)
}

// Span tracks a processing activity within a trace.
type Span struct {
	SpanId    int64                  `yaml:"span_id"`
	TraceId   int64                  `yaml:"trace_id"`
	ParentId  int64                  `yaml:"parent_id"`
	Process   string                 `yaml:",omitempty"`
	Kind      string                 `yaml:",omitempty"`
	Name      string                 `yaml:",omitempty"`
	Start     time.Time              `yaml:"-"`
	StartStr  string                 `yaml:"start,omitempty"`
	Finish    time.Time              `yaml:"-"`
	FinishStr string                 `yaml:"finish,omitempty"`
	DataMap   map[string]interface{} `yaml:",omitempty,inline"`
}

func (s *Span) Data() map[string]interface{} {
	if s.DataMap == nil {
		s.DataMap = make(map[string]interface{})
	}
	return s.DataMap
}

// Recorder instances persist Spans to an external datastore.
type Recorder interface {
	Record(s *Span) error
}

// CurrentSpanId returns the caller's current span id.
func CurrentSpanId() int64 {
	if cm == nil {
		return 0
	}

	v, ok := cm.GetValue(spanIdKey)
	if !ok {
		return 0
	}

	spanId, ok := v.(int64)
	if !ok {
		return 0
	}

	return spanId
}

// CurrentTraceId returns the caller's current trace id.
func CurrentTraceId() int64 {
	if cm == nil {
		return 0
	}

	v, ok := cm.GetValue(traceIdKey)
	if !ok {
		return 0
	}

	traceId, ok := v.(int64)
	if !ok {
		return 0
	}

	return traceId
}

// Record starts recording in a goroutine.  Because Run must not be
// allowed to block, buffer must be greater than zero.  If a Logger is
// provided, then errors that occur during recording will be logged.
func Record(rec Recorder, buffer int, logger *log.Logger) error {
	if buffer < 1 {
		return errBufferRequired
	}

	cm = gls.NewContextManager()
	spans = make(chan *Span, buffer)
	go record(rec, logger)
	return nil
}

func record(rec Recorder, logger *log.Logger) {
	for s := range spans {
		if err := rec.Record(s); err != nil {
			if logger != nil {
				log.Printf("failed to record trace %x span %x: %s", s.TraceId, s.SpanId, err)
			}
		}
	}
}

// New starts a new trace.  If recording is active, a new Span is
// allocated and returned, otherwise no allocation occurs and nil is
// returned (along with an error).
//
// As a caller convenience, if traceId is non-zero, then that value is
// used instead of generating a probablistically unique id.  This may
// be useful for callers that want to generate their own id values.
func New(traceId int64, kind string, name string) (*Span, error) {
	if spans == nil {
		return nil, errRecNotActive
	}

	spanId, err := GenerateId()
	if err != nil {
		return nil, err
	}

	if traceId == 0 {
		traceId, err = GenerateId()
		if err != nil {
			return nil, err
		}
	}

	return &Span{
		SpanId:  spanId,
		TraceId: traceId,
		Process: Process,
		Kind:    kind,
		Name:    name,
	}, nil
}

// Continue continues an existing trace.  If recording is active, a
// new Span instance is allocated and returned, otherwise no
// allocation occurs and nil is returned (along with an error).
func Continue(kind string, name string) (*Span, error) {
	if spans == nil {
		return nil, errRecNotActive
	}

	parentId := CurrentSpanId()
	traceId := CurrentTraceId()
	if parentId == 0 || traceId == 0 {
		s, err := New(0, kind, name)
		return s, err
	}

	spanId, err := GenerateId()
	if err != nil {
		return nil, err
	}

	return &Span{
		SpanId:   spanId,
		TraceId:  traceId,
		ParentId: parentId,
		Process:  Process,
		Kind:     kind,
		Name:     name,
	}, nil
}

// Run records a Span (to provide visibility that the span has
// started), invokes the function f, and then records the Span a
// second time (to update the finish time).
func Run(s *Span, f func()) {
	// If New or Continue returned nil, then ts is probably also
	// nil. We quietly tolerate so that callers don't need to know
	// or care whether recording is active.
	if s != nil && spans != nil {
		// Setup to record the span finish
		defer func() {
			s.Finish = time.Now()
			s.FinishStr = s.Finish.Format(time.RFC3339Nano)
			spans <- s
		}()

		// Save the span start time
		s.Start = time.Now()
		s.StartStr = s.Start.Format(time.RFC3339Nano)

		// Stash the span id and trace id on the stack and invoke f
		values := gls.Values{
			spanIdKey:  s.SpanId,
			traceIdKey: s.TraceId,
		}

		cm.SetValues(values, f)
	} else {
		f()
	}
}

// Go functions similarly to Run, except that f is run in a new goroutine.
func Go(s *Span, f func()) {
	if s != nil && spans != nil {
		gls.Go(func() {
			Run(s, f)
		})
	} else {
		go f()
	}
}

// GenerateId returns a probablistically unique 64-bit id.  All id
// values returned by this function will be positive integers.  This
// may be useful for callers that want to generate their own id
// values.
func GenerateId() (int64, error) {
	// Return a random int64, constrained to positive values
	for retry := 0; retry < 3; retry++ {
		var x uint64
		if err := binary.Read(rand.Reader, binary.LittleEndian, &x); err != nil {
			return 0, err
		}

		id := int64(x & math.MaxInt64)
		if id > 0 {
			return id, nil
		}
	}

	// Failsafe
	return 0, errors.New("rand.Reader failed to produce a useable value")
}

type Handler struct {
	// Kind is the kind value used when starting new traces.
	Kind string
	// HeaderKey is the key used when ServeHTTP inserts an id header in requests or responses.
	HeaderKey string
	// HonorReqHeader determines whether or not ServeHTTP honors id headers in requests.
	HonorReqHeader bool
	// AddRespHeader determines whether or not ServeHTTP adds id headers to responses.
	AddRespHeader bool
}

// NewHandler creates a middleware handler that facilitates HTTP
// request tracing.
//
// If the request contains an id header and HonorReqHeader is true,
// then the id values are used (allowing trace contexts to span
// services).  Otherwise a new trace id is generated. An id header is
// optionally added to the response.
func NewHandler() *Handler {
	return &Handler{
		Kind:           "request",
		HeaderKey:      "X-Request-Id",
		HonorReqHeader: false,
		AddRespHeader:  true,
	}
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	var (
		traceId, parentId int64
		s                 *Span
	)

	// Optionally honor incoming id headers. If present, header must be in the form "traceId:parentId".
	if h.HonorReqHeader {
		if hdr := req.Header.Get(h.HeaderKey); hdr != "" {
			var traceId, parentId int64
			n, _ := fmt.Sscanf(hdr, "%d:%d", &traceId, &parentId)
			if n < 2 || traceId < 1 || parentId < 1 {
				traceId = 0
				parentId = 0
			}
		}
	}

	// Start a new trace, either using an existing id (from the request header) or a new one
	s, err := New(traceId, h.Kind, fmt.Sprintf("%s %s", req.Method, req.URL.Path))
	if err == nil {
		s.ParentId = parentId

		// Add headers
		req.Header.Set(h.HeaderKey, fmt.Sprintf("%d:%d", s.TraceId, s.SpanId))
		if h.AddRespHeader {
			rw.Header().Set(h.HeaderKey, fmt.Sprintf("%d", s.TraceId))
		}

		// Invoke the next handler
		Run(s, func() {
			next(rw, req)
		})
	} else {
		// Invoke the next handler
		next(rw, req)
	}
}
