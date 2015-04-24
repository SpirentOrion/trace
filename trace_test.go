package trace

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type nullRecorder struct {
}

func (r *nullRecorder) Record(s *Span) error {
	return nil
}

func TestMain(m *testing.M) {
	Record(&nullRecorder{}, 1, nil)
	os.Exit(m.Run())
}

func TestNoContext(t *testing.T) {
	if spanId := CurrentSpanId(); spanId != 0 {
		t.Error("non-zero CurrentSpanId() value")
	}

	if traceId := CurrentTraceId(); traceId != 0 {
		t.Error("non-zero CurrentTraceId() value")
	}
}

func TestNew(t *testing.T) {
	s, err := New(1, "testing", "TestNew")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if s.SpanId < 1 {
		t.Error("New() allocated span with invalid span id")
	}

	if s.TraceId != 1 {
		t.Error("New() ignored caller-provided trace id")
	}

	if len(s.Process) == 0 {
		t.Error("New() allocated span without process")
	}

	s, err = New(0, "testing", "TestNew")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if s.SpanId < 1 {
		t.Error("New() allocated span with invalid span id")
	}

	if s.TraceId < 1 {
		t.Error("New() allocated span with invalid trace id")
	}
}

func TestContinue(t *testing.T) {
	s0, err := New(0, "testing", "TestContinue")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	Run(s0, func() {
		if spanId := CurrentSpanId(); spanId != s0.SpanId {
			t.Errorf("CurrentSpanId() returned invalid id %x, expected %x", spanId, s0.SpanId)
		}

		if traceId := CurrentTraceId(); traceId != s0.TraceId {
			t.Errorf("CurrentTraceId() returned invalid id %x, expected %x", traceId, s0.TraceId)
		}

		s1, err := Continue("testing", "TestContinue")
		if err != nil {
			t.Log(err)
			t.FailNow()
		}

		if s1.SpanId < 1 {
			t.Error("Continue() allocated span with invalid span id")
		}

		if s1.SpanId == s0.SpanId {
			t.Error("Continue() allocated span with parent's span id")
		}

		if s1.TraceId != s0.TraceId {
			t.Errorf("Continue() allocated span with invalid trace id %x, expected %x", s1.TraceId, s0.TraceId)
		}

		if len(s1.Process) == 0 {
			t.Error("Continue() allocated span without process")
		}
	})
}

func TestGenerateId(t *testing.T) {
	var (
		id0, id1 int64
		err      error
	)

	for i := 0; i < 1000; i++ {
		id1, err = GenerateId()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}

		if id1 < 1 {
			t.Logf("generated negative or zero value: %d", id1)
			t.FailNow()
		}

		if id0 != 0 && id0 == id1 {
			t.Logf("generated sequentially duplicate value: %d", id1)
			t.FailNow()
		}

		id0 = id1
	}
}

func TestMiddleware(t *testing.T) {
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	h := NewHandler()
	h.ServeHTTP(recorder, req, func(rw http.ResponseWriter, req *http.Request) {
		if _, ok := req.Header[h.HeaderKey]; !ok {
			t.Error("ServeHTTP() failed to add request header")
		}
		if _, ok := rw.Header()[h.HeaderKey]; !ok {
			t.Error("ServeHTTP() failed to add response header")
		}
	})
}
