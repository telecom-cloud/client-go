package http1

import (
	"context"
	"fmt"
	"sync"
	"testing"

	inStats "github.com/telecom-cloud/client-go/internal/stats"
	"github.com/telecom-cloud/client-go/pkg/app"
	"github.com/telecom-cloud/client-go/pkg/common/test/mock"
	"github.com/telecom-cloud/client-go/pkg/common/tracer/traceinfo"
)

func BenchmarkServer_Serve(b *testing.B) {
	server := &Server{}
	server.eventStackPool = &sync.Pool{
		New: func() interface{} {
			return &eventStack{}
		},
	}
	server.EnableTrace = true
	reqCtx := &app.RequestContext{}
	server.Core = &mockCore{
		ctxPool: &sync.Pool{New: func() interface{} {
			ti := traceinfo.NewTraceInfo()
			ti.Stats().SetLevel(2)
			reqCtx.SetTraceInfo(&mockTraceInfo{ti})
			return reqCtx
		}},
		controller: &inStats.Controller{},
	}
	err := server.Serve(context.TODO(), mock.NewConn("GET /aaa HTTP/1.1\nHost: foobar.com\n\n"))
	if err != nil {
		fmt.Println(err.Error())
	}
	for i := 0; i < b.N; i++ {
		server.Serve(context.TODO(), mock.NewConn("GET /aaa HTTP/1.1\nHost: foobar.com\n\n"))
	}
}
