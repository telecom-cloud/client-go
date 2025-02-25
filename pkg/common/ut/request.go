package ut

import (
	"context"
	"io"

	"github.com/telecom-cloud/client-go/pkg/route"
)

// Header is a key-value pair indicating one http header
type Header struct {
	Key   string
	Value string
}

// Body is for setting Request.Body
type Body struct {
	Body io.Reader
	Len  int
}

// PerformRequest send a constructed request to given engine without network transporting
//
// # Url can be a standard relative URI or a simple absolute path
//
// If engine.streamRequestBody is true, it sets body as bodyStream
// if not, it sets body as bodyBytes
//
// ResponseRecorder returned are flushed, which means its StatusCode is always set (default 200)
//
// See ./request_test.go for more examples
func PerformRequest(engine *route.Engine, method, url string, body *Body, headers ...Header) *ResponseRecorder {
	ctx := createUtRequestContext(engine, method, url, body, headers...)
	engine.ServeHTTP(context.Background(), ctx)

	w := NewRecorder()
	h := w.Header()
	ctx.Response.Header.CopyTo(h)

	w.WriteHeader(ctx.Response.StatusCode())
	w.Write(ctx.Response.Body())
	w.Flush()
	return w
}
