package http1

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	inStats "github.com/telecom-cloud/client-go/internal/stats"
	"github.com/telecom-cloud/client-go/pkg/app"
	errs "github.com/telecom-cloud/client-go/pkg/common/errors"
	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
	"github.com/telecom-cloud/client-go/pkg/common/test/mock"
	"github.com/telecom-cloud/client-go/pkg/common/tracer"
	"github.com/telecom-cloud/client-go/pkg/common/tracer/stats"
	"github.com/telecom-cloud/client-go/pkg/common/tracer/traceinfo"
	"github.com/telecom-cloud/client-go/pkg/network"
	"github.com/telecom-cloud/client-go/pkg/protocol"
	"github.com/telecom-cloud/client-go/pkg/protocol/consts"
	"github.com/telecom-cloud/client-go/pkg/protocol/http1/resp"
)

var pool = &sync.Pool{New: func() interface{} {
	return &eventStack{}
}}

func TestTraceEventCompleted(t *testing.T) {
	server := &Server{}
	server.eventStackPool = pool
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
	assert.True(t, errors.Is(err, errs.ErrShortConnection))
	traceInfo := reqCtx.GetTraceInfo()
	assert.False(t, traceInfo.Stats().GetEvent(stats.HTTPStart).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.ReadHeaderStart).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.ReadHeaderFinish).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.ReadBodyStart).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.ReadBodyFinish).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.ServerHandleStart).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.ServerHandleFinish).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.WriteStart).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.WriteFinish).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.HTTPFinish).IsNil())
	assert.Nil(t, traceInfo.Stats().Error())
}

func TestTraceEventReadHeaderError(t *testing.T) {
	server := &Server{}
	server.eventStackPool = pool
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
	err := server.Serve(context.TODO(), mock.NewConn("ErrorFirstLine\r\n\r\n"))
	assert.NotNil(t, err)
	traceInfo := reqCtx.GetTraceInfo()
	assert.False(t, traceInfo.Stats().GetEvent(stats.HTTPStart).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.ReadHeaderStart).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.ReadHeaderFinish).IsNil())
	assert.Nil(t, traceInfo.Stats().GetEvent(stats.ReadBodyStart))
	assert.Nil(t, traceInfo.Stats().GetEvent(stats.ReadBodyFinish))
	assert.Nil(t, traceInfo.Stats().GetEvent(stats.ServerHandleStart))
	assert.Nil(t, traceInfo.Stats().GetEvent(stats.ServerHandleFinish))
	assert.Nil(t, traceInfo.Stats().GetEvent(stats.WriteStart))
	assert.Nil(t, traceInfo.Stats().GetEvent(stats.WriteFinish))
	assert.False(t, traceInfo.Stats().GetEvent(stats.HTTPFinish).IsNil())
}

func TestTraceEventReadBodyError(t *testing.T) {
	server := &Server{}
	server.eventStackPool = pool
	server.EnableTrace = true
	server.GetOnly = true
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
	err := server.Serve(context.TODO(), mock.NewConn("POST /aaa HTTP/1.1\nHost: foobar.com\nContent-Length: 5\nContent-Type: foo/bar\n\n12346\n\n"))
	assert.NotNil(t, err)

	traceInfo := reqCtx.GetTraceInfo()
	assert.False(t, traceInfo.Stats().GetEvent(stats.HTTPStart).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.ReadHeaderStart).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.ReadHeaderFinish).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.ReadBodyStart).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.ReadBodyFinish).IsNil())
	assert.Nil(t, traceInfo.Stats().GetEvent(stats.ServerHandleStart))
	assert.Nil(t, traceInfo.Stats().GetEvent(stats.ServerHandleFinish))
	assert.Nil(t, traceInfo.Stats().GetEvent(stats.WriteStart))
	assert.Nil(t, traceInfo.Stats().GetEvent(stats.WriteFinish))
	assert.False(t, traceInfo.Stats().GetEvent(stats.HTTPFinish).IsNil())
}

func TestTraceEventWriteError(t *testing.T) {
	server := &Server{}
	server.eventStackPool = pool
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
	err := server.Serve(
		context.TODO(),
		&mockErrorWriter{
			mock.NewConn("POST /aaa HTTP/1.1\nHost: foobar.com\nContent-Length: 5\nContent-Type: foo/bar\n\n12346\n\n"),
		},
	)
	assert.NotNil(t, err)
	traceInfo := reqCtx.GetTraceInfo()
	assert.False(t, traceInfo.Stats().GetEvent(stats.HTTPStart).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.ReadHeaderStart).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.ReadHeaderFinish).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.ReadBodyStart).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.ReadBodyFinish).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.ServerHandleStart).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.ServerHandleFinish).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.WriteStart).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.WriteFinish).IsNil())
	assert.False(t, traceInfo.Stats().GetEvent(stats.HTTPFinish).IsNil())
}

func TestEventStack(t *testing.T) {
	// Create a stack.
	s := &eventStack{}
	assert.True(t, s.isEmpty())

	count := 0

	// Push 10 events.
	for i := 0; i < 10; i++ {
		s.push(func(ti traceinfo.TraceInfo, err error) {
			count += 1
		})
	}

	assert.False(t, s.isEmpty())
	// Pop 10 events and process them.
	for last := s.pop(); last != nil; last = s.pop() {
		last(nil, nil)
	}

	assert.DeepEqual(t, 10, count)

	// Pop an empty stack.
	e := s.pop()
	if e != nil {
		t.Fatalf("should be nil")
	}
}

func TestDefaultWriter(t *testing.T) {
	server := &Server{}
	reqCtx := &app.RequestContext{}
	server.Core = &mockCore{
		ctxPool: &sync.Pool{New: func() interface{} {
			return reqCtx
		}},
		mockHandler: func(c context.Context, ctx *app.RequestContext) {
			ctx.Write([]byte("hello, hertz"))
			ctx.Flush()
		},
	}
	defaultConn := mock.NewConn("GET / HTTP/1.1\nHost: foobar.com\n\n")
	err := server.Serve(context.TODO(), defaultConn)
	assert.True(t, errors.Is(err, errs.ErrShortConnection))
	defaultResponseResult := defaultConn.WriterRecorder()
	assert.DeepEqual(t, 0, defaultResponseResult.Len()) // all data is flushed so the buffer length is 0
	response := protocol.AcquireResponse()
	resp.Read(response, defaultResponseResult)
	assert.DeepEqual(t, "hello, hertz", string(response.Body()))
}

func TestServerDisableReqCtxPool(t *testing.T) {
	server := &Server{}
	reqCtx := &app.RequestContext{}
	server.Core = &mockCore{
		ctxPool: &sync.Pool{New: func() interface{} {
			reqCtx.Set("POOL_KEY", "in pool")
			return reqCtx
		}},
		mockHandler: func(c context.Context, ctx *app.RequestContext) {
			if ctx.GetString("POOL_KEY") != "in pool" {
				t.Fatal("reqCtx is not in pool")
			}
		},
		isRunning: true,
	}
	defaultConn := mock.NewConn("GET / HTTP/1.1\nHost: foobar.com\n\n")
	err := server.Serve(context.TODO(), defaultConn)
	assert.Nil(t, err)
	disabaleRequestContextPool = true
	defer func() {
		// reset global variable
		disabaleRequestContextPool = false
	}()
	server.Core = &mockCore{
		ctxPool: &sync.Pool{New: func() interface{} {
			reqCtx.Set("POOL_KEY", "in pool")
			return reqCtx
		}},
		mockHandler: func(c context.Context, ctx *app.RequestContext) {
			if len(ctx.GetString("POOL_KEY")) != 0 {
				t.Fatal("must not get pool key")
			}
		},
		isRunning: true,
	}
	defaultConn = mock.NewConn("GET / HTTP/1.1\nHost: foobar.com\n\n")
	err = server.Serve(context.TODO(), defaultConn)
	assert.Nil(t, err)
}

func TestHijackResponseWriter(t *testing.T) {
	server := &Server{}
	reqCtx := &app.RequestContext{}
	buf := new(bytes.Buffer)
	isFinal := false
	server.Core = &mockCore{
		ctxPool: &sync.Pool{New: func() interface{} {
			return reqCtx
		}},
		mockHandler: func(c context.Context, ctx *app.RequestContext) {
			// response before write will be dropped
			ctx.Write([]byte("invalid data"))

			ctx.Response.HijackWriter(&mock.ExtWriter{
				Buf:     buf,
				IsFinal: &isFinal,
			})

			ctx.Write([]byte("hello, hertz"))
			ctx.Flush()
		},
	}
	defaultConn := mock.NewConn("GET / HTTP/1.1\nHost: foobar.com\n\n")
	err := server.Serve(context.TODO(), defaultConn)
	assert.True(t, errors.Is(err, errs.ErrShortConnection))
	defaultResponseResult := defaultConn.WriterRecorder()
	response := protocol.AcquireResponse()
	resp.Read(response, defaultResponseResult)
	assert.DeepEqual(t, 0, len(response.Body()))
	assert.DeepEqual(t, "hello, hertz", buf.String())
	assert.True(t, isFinal)
}

func TestHijackHandler(t *testing.T) {
	server := NewServer()
	reqCtx := &app.RequestContext{}
	originReadTimeout := time.Second
	hijackReadTimeout := 200 * time.Millisecond
	reqCtx.SetHijackHandler(func(c network.Conn) {
		c.SetReadTimeout(hijackReadTimeout) // hijack read timeout
	})

	server.Core = &mockCore{
		ctxPool: &sync.Pool{New: func() interface{} {
			return reqCtx
		}},
	}

	server.HijackConnHandle = func(c network.Conn, h app.HijackHandler) {
		h(c)
	}

	defaultConn := mock.NewConn("GET / HTTP/1.1\nHost: foobar.com\n\n")
	defaultConn.SetReadTimeout(originReadTimeout)
	assert.DeepEqual(t, originReadTimeout, defaultConn.GetReadTimeout())
	err := server.Serve(context.TODO(), defaultConn)
	assert.True(t, errors.Is(err, errs.ErrHijacked))
	assert.DeepEqual(t, hijackReadTimeout, defaultConn.GetReadTimeout())
}

func TestKeepAlive(t *testing.T) {
	server := NewServer()
	reqCtx := &app.RequestContext{}
	times := 0
	server.Core = &mockCore{
		ctxPool: &sync.Pool{New: func() interface{} {
			return reqCtx
		}},
		isRunning: true,
		mockHandler: func(c context.Context, ctx *app.RequestContext) {
			times++
			if string(ctx.Path()) == "/close" {
				ctx.SetConnectionClose()
			}
		},
	}
	server.IdleTimeout = time.Second

	var s strings.Builder
	s.WriteString("GET / HTTP/1.1\r\nHost: aaa\r\nConnection: keep-alive\r\n\r\n")
	s.WriteString("GET /close HTTP/1.0\r\nHost: aaa\r\nConnection: keep-alive\r\n\r\n") // set connection close

	defaultConn := mock.NewConn(s.String())
	err := server.Serve(context.TODO(), defaultConn)
	assert.True(t, errors.Is(err, errs.ErrShortConnection))
	assert.DeepEqual(t, times, 2)
}

func TestExpect100Continue(t *testing.T) {
	server := &Server{}
	reqCtx := &app.RequestContext{}
	server.Core = &mockCore{
		ctxPool: &sync.Pool{New: func() interface{} {
			return reqCtx
		}},
		mockHandler: func(c context.Context, ctx *app.RequestContext) {
			data, err := ctx.Body()
			if err == nil {
				ctx.Write(data)
			}
		},
	}

	defaultConn := mock.NewConn("POST /foo HTTP/1.1\r\nHost: gle.com\r\nExpect: 100-continue\r\nContent-Length: 5\r\nContent-Type: a/b\r\n\r\n12345")
	err := server.Serve(context.TODO(), defaultConn)
	assert.True(t, errors.Is(err, errs.ErrShortConnection))
	defaultResponseResult := defaultConn.WriterRecorder()
	assert.DeepEqual(t, 0, defaultResponseResult.Len())
	response := protocol.AcquireResponse()
	resp.Read(response, defaultResponseResult)
	assert.DeepEqual(t, "12345", string(response.Body()))
}

func TestExpect100ContinueHandler(t *testing.T) {
	server := &Server{}
	reqCtx := &app.RequestContext{}
	server.Core = &mockCore{
		ctxPool: &sync.Pool{New: func() interface{} {
			return reqCtx
		}},
		mockHandler: func(c context.Context, ctx *app.RequestContext) {
			data, err := ctx.Body()
			if err == nil {
				ctx.Write(data)
			}
		},
	}
	server.ContinueHandler = func(header *protocol.RequestHeader) bool {
		return false
	}

	defaultConn := mock.NewConn("POST /foo HTTP/1.1\r\nHost: gle.com\r\nExpect: 100-continue\r\nContent-Length: 5\r\nContent-Type: a/b\r\n\r\n12345")
	err := server.Serve(context.TODO(), defaultConn)
	assert.True(t, errors.Is(err, errs.ErrShortConnection))
	defaultResponseResult := defaultConn.WriterRecorder()
	assert.DeepEqual(t, 0, defaultResponseResult.Len())
	response := protocol.AcquireResponse()
	resp.Read(response, defaultResponseResult)
	assert.DeepEqual(t, consts.StatusExpectationFailed, response.StatusCode())
	assert.DeepEqual(t, "", string(response.Body()))
}

type mockController struct {
	FinishTimes int
}

func (m *mockController) Append(col tracer.Tracer) {}

func (m *mockController) DoStart(ctx context.Context, c *app.RequestContext) context.Context {
	return ctx
}

func (m *mockController) DoFinish(ctx context.Context, c *app.RequestContext, err error) {
	m.FinishTimes++
}

func (m *mockController) HasTracer() bool { return true }

func (m *mockController) reset() { m.FinishTimes = 0 }

func TestTraceDoFinishTimes(t *testing.T) {
	server := &Server{}
	server.eventStackPool = pool
	server.EnableTrace = true
	reqCtx := &app.RequestContext{}
	controller := &mockController{}
	server.Core = &mockCore{
		ctxPool: &sync.Pool{New: func() interface{} {
			ti := traceinfo.NewTraceInfo()
			ti.Stats().SetLevel(2)
			reqCtx.SetTraceInfo(&mockTraceInfo{ti})
			return reqCtx
		}},
		controller: controller,
	}
	// for disableKeepAlive case
	server.DisableKeepalive = true
	err := server.Serve(context.TODO(), mock.NewConn("GET /aaa HTTP/1.1\nHost: foobar.com\n\n"))
	assert.True(t, errors.Is(err, errs.ErrShortConnection))
	assert.DeepEqual(t, 1, controller.FinishTimes)
	// for IdleTimeout==0 case
	server.IdleTimeout = 0
	controller.reset()
	err = server.Serve(context.TODO(), mock.NewConn("GET /aaa HTTP/1.1\nHost: foobar.com\n\n"))
	assert.True(t, errors.Is(err, errs.ErrShortConnection))
	assert.DeepEqual(t, 1, controller.FinishTimes)
}

type mockCore struct {
	ctxPool     *sync.Pool
	controller  tracer.Controller
	mockHandler func(c context.Context, ctx *app.RequestContext)
	isRunning   bool
}

func (m *mockCore) IsRunning() bool {
	return m.isRunning
}

func (m *mockCore) GetCtxPool() *sync.Pool {
	return m.ctxPool
}

func (m *mockCore) ServeHTTP(c context.Context, ctx *app.RequestContext) {
	if m.mockHandler != nil {
		m.mockHandler(c, ctx)
	}
}

func (m *mockCore) GetTracer() tracer.Controller {
	return m.controller
}

type mockTraceInfo struct {
	traceinfo.TraceInfo
}

func (m *mockTraceInfo) Reset() {}

type mockErrorWriter struct {
	network.Conn
}

func (errorWriter *mockErrorWriter) Flush() error {
	return errors.New("error")
}

func TestShouldRecordInTraceError(t *testing.T) {
	assert.False(t, shouldRecordInTraceError(nil))
	assert.False(t, shouldRecordInTraceError(errHijacked))
	assert.False(t, shouldRecordInTraceError(errIdleTimeout))
	assert.False(t, shouldRecordInTraceError(errShortConnection))

	assert.True(t, shouldRecordInTraceError(errTimeout))
	assert.True(t, shouldRecordInTraceError(errors.New("foo error")))
}
