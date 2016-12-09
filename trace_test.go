package trace

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type nullRecorder struct {
}

func (r *nullRecorder) Record(s *Span) error {
	return nil
}

func TestCurrentSpanIDGivenNoTrace(t *testing.T) {
	if spanID := CurrentSpanID(context.Background()); spanID != 0 {
		t.Error("got non-zero span id")
	}
}

func TestCurrentTraceIDGivenNoTrace(t *testing.T) {
	if traceID := CurrentTraceID(context.Background()); traceID != 0 {
		t.Error("got non-zero trace id")
	}
}

func TestDo(t *testing.T) {
	ctx0, cancelFunc := context.WithCancel(context.Background())
	defer func() {
		cancelFunc()
	}()

	ctx0, err := Record(ctx0, &nullRecorder{})
	if err != nil {
		t.Fatal(err)
	}

	Do(ctx0, "testing", "TestInOriginalContext", func(ctx1 context.Context) { assertInTestContext(t, ctx1) })

	jctx0, err := Join(context.Background(), ctx0)
	if err != nil {
		t.Fatal(err)
	}

	Do(jctx0, "testing", "TestInJoinedContext", func(ctx1 context.Context) { assertInTestContext(t, ctx1) })
}

func assertInTestContext(t *testing.T, ctx1 context.Context) {
	spanID1 := CurrentSpanID(ctx1)
	if spanID1 <= 0 {
		t.Error("Do() started trace with invalid span id")
	}

	traceID1 := CurrentTraceID(ctx1)
	if traceID1 <= 0 {
		t.Error("Do() started trace with invalid trace id")
	}

	Do(ctx1, "testing", "Test2", func(ctx2 context.Context) {
		spanID2 := CurrentSpanID(ctx2)
		if spanID2 <= 0 {
			t.Error("Do() continued trace with invalid span id")
		}
		if spanID2 == spanID1 {
			t.Error("Do() continued trace with parent's span id")
		}

		traceID2 := CurrentTraceID(ctx2)
		if traceID2 <= 0 {
			t.Error("Do() continued trace with invalid trace id")
		}
		if traceID2 != traceID1 {
			t.Error("Do() continued trace but generated a new trace id")
		}
	})
}

func TestGenerateID(t *testing.T) {
	var (
		id0, id1 int64
		err      error
	)

	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		if id1, err = GenerateID(ctx); err != nil {
			t.Fatal(err)
		}
		if id1 < 1 {
			t.Fatalf("generated negative or zero value: %d", id1)
		}
		if id0 != 0 && id0 == id1 {
			t.Fatalf("generated sequentially duplicate value: %d", id1)
		}
		id0 = id1
	}
}

func TestHandler(t *testing.T) {
	ctx0, cancelFunc := context.WithCancel(context.Background())
	defer func() {
		cancelFunc()
	}()

	var err error
	if ctx0, err = Record(ctx0, &nullRecorder{}); err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req = req.WithContext(ctx0)

	h := NewHandler(ctx0)
	h.ServeHTTP(recorder, req, func(rw http.ResponseWriter, req *http.Request) {
		if _, ok := req.Header[h.HeaderKey]; !ok {
			t.Error("ServeHTTP() failed to add request header")
		}
		if _, ok := rw.Header()[h.HeaderKey]; !ok {
			t.Error("ServeHTTP() failed to add response header")
		}
	})
}
