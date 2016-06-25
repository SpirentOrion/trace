package trace

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	// Kind is the kind value used when starting new traces.
	Kind string

	// HeaderKey is the key used when ServeHTTP inserts an id header in requests or responses.
	HeaderKey string

	// HonorReqHeader determines whether or not ServeHTTP honors id headers in requests.
	HonorReqHeader bool

	// AddRespHeader determines whether or not ServeHTTP adds id headers to responses.
	AddRespHeader bool

	ctx context.Context
}

// NewHandler creates a middleware handler that facilitates HTTP
// request tracing.
//
// If the request contains an id header and HonorReqHeader is true,
// then the id values are used (allowing trace contexts to span
// services).  Otherwise a new trace id is generated. An id header is
// optionally added to the response.
func NewHandler(ctx context.Context) *Handler {
	return &Handler{
		Kind:           "request",
		HeaderKey:      "X-Request-Id",
		HonorReqHeader: false,
		AddRespHeader:  true,
		ctx:            ctx,
	}
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	ctx0 := h.ctx

	// Optionally honor incoming id headers. If present, header must be in the form "traceID:parentID".
	if h.HonorReqHeader {
		if hdr := req.Header.Get(h.HeaderKey); hdr != "" {
			parts := strings.Split(hdr, ":")
			if len(parts) == 2 {
				traceID, _ := strconv.ParseInt(parts[0], 10, 64)
				parentID, _ := strconv.ParseInt(parts[1], 10, 64)
				if traceID > 0 && parentID > 0 {
					req = req.WithContext(WithParentID(WithTraceID(ctx0, traceID), parentID))
				}
			}
		}
	}

	// Start a new trace span wrapping request processing
	Do(ctx0, h.Kind, req.Method+" "+req.URL.Path, func(ctx1 context.Context) {
		// Add headers
		if traceID := CurrentTraceID(ctx1); traceID != 0 {
			req.Header.Set(h.HeaderKey, fmt.Sprintf("%d:%d", traceID, CurrentSpanID(ctx1)))
			if h.AddRespHeader {
				rw.Header().Set(h.HeaderKey, fmt.Sprint(traceID))
			}
		}

		// Invoke the next handler
		next(rw, req.WithContext(ctx1))
	})
}
