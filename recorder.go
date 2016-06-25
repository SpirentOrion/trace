package trace

import (
	"encoding/json"
	"io"
	"time"
)

var (
	newlineBuffer = []byte{'\n'}
)

// Span tracks an activity within a trace.
// Note: YAML struct tags are included for backwards-compatibility with V1 code.
type Span struct {
	SpanID    int64                  `json:"span_id" yaml:"span_id"`
	TraceID   int64                  `json:"trace_id" yaml:"trace_id"`
	ParentID  int64                  `json:"parent_id" yaml:"parent_id"`
	Process   string                 `json:"process,omitempty" yaml:",omitempty"`
	Kind      string                 `json:"kind,omitempty" yaml:",omitempty"`
	Name      string                 `json:"name,omitempty" yaml:",omitempty"`
	Start     time.Time              `json:"-" yaml:"-"`
	StartStr  string                 `json:"start,omitempty" yaml:"start,omitempty"`
	Finish    time.Time              `json:"-" yaml:"-"`
	FinishStr string                 `json:"finish,omitempty" yaml:"finish,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty" yaml:",omitempty,inline"`
}

// Recorder persists a Span to an external file or datastore.
type Recorder interface {
	Record(s *Span) error
}

type jsonRecorder struct {
	io.Writer
}

func NewJSONRecorder(writer io.Writer) Recorder {
	return &jsonRecorder{writer}
}

func (r *jsonRecorder) Record(s *Span) error {
	buf, err := json.Marshal(s)
	if err != nil {
		return err
	}

	if _, err = r.Write(buf); err != nil {
		return err
	}

	_, err = r.Write(newlineBuffer)
	return err
}
