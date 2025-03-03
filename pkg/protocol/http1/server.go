package http1

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/telecom-cloud/client-go/internal/bytestr"
	internalStats "github.com/telecom-cloud/client-go/internal/stats"
	"github.com/telecom-cloud/client-go/pkg/app"
	errs "github.com/telecom-cloud/client-go/pkg/common/errors"
	"github.com/telecom-cloud/client-go/pkg/common/render"
	"github.com/telecom-cloud/client-go/pkg/common/tracer/stats"
	"github.com/telecom-cloud/client-go/pkg/common/tracer/traceinfo"
	"github.com/telecom-cloud/client-go/pkg/common/utils"
	"github.com/telecom-cloud/client-go/pkg/network"
	"github.com/telecom-cloud/client-go/pkg/protocol"
	"github.com/telecom-cloud/client-go/pkg/protocol/consts"
	"github.com/telecom-cloud/client-go/pkg/protocol/http1/ext"
	"github.com/telecom-cloud/client-go/pkg/protocol/http1/req"
	"github.com/telecom-cloud/client-go/pkg/protocol/http1/resp"
	"github.com/telecom-cloud/client-go/pkg/protocol/suite"
)

func init() {
	if b, err := utils.GetBoolFromEnv("HERTZ_DISABLE_REQUEST_CONTEXT_POOL"); err == nil {
		disabaleRequestContextPool = b
	}
}

// NextProtoTLS is the NPN/ALPN protocol negotiated during
// HTTP/1.1's TLS setup.
// Also used for server addressing
const NextProtoTLS = suite.HTTP1

var (
	errHijacked        = errs.New(errs.ErrHijacked, errs.ErrorTypePublic, nil)
	errIdleTimeout     = errs.New(errs.ErrIdleTimeout, errs.ErrorTypePrivate, nil)
	errShortConnection = errs.New(errs.ErrShortConnection, errs.ErrorTypePublic, "server is going to close the connection")
	errUnexpectedEOF   = errs.NewPublic(io.ErrUnexpectedEOF.Error() + " when reading request")

	disabaleRequestContextPool = false
)

type Option struct {
	StreamRequestBody             bool
	GetOnly                       bool
	NoDefaultDate                 bool
	NoDefaultContentType          bool
	DisablePreParseMultipartForm  bool
	DisableKeepalive              bool
	NoDefaultServerHeader         bool
	DisableHeaderNamesNormalizing bool
	MaxRequestBodySize            int
	IdleTimeout                   time.Duration
	ReadTimeout                   time.Duration
	ServerName                    []byte
	TLS                           *tls.Config
	HTMLRender                    render.HTMLRender
	EnableTrace                   bool
	ContinueHandler               func(header *protocol.RequestHeader) bool
	HijackConnHandle              func(c network.Conn, h app.HijackHandler)
}

type Server struct {
	Option
	Core suite.Core

	eventStackPool *sync.Pool
}

func (s Server) getRequestContext() *app.RequestContext {
	if disabaleRequestContextPool {
		return &app.RequestContext{}
	}
	return s.Core.GetCtxPool().Get().(*app.RequestContext)
}

func (s Server) putRequestContext(ctx *app.RequestContext) {
	if disabaleRequestContextPool {
		return
	}
	ctx.Reset()
	s.Core.GetCtxPool().Put(ctx)
}

func (s Server) Serve(c context.Context, conn network.Conn) (err error) {
	var (
		zr network.Reader
		zw network.Writer

		serverName      []byte
		isHTTP11        bool
		connectionClose bool

		continueReadingRequest = true

		hijackHandler app.HijackHandler

		// HTTP1 path
		// 1. Get a request context
		// 2. Prepare it
		// 3. Process it
		// 4. Reset and recycle(in pooled mode)
		ctx = s.getRequestContext()

		traceCtl        = s.Core.GetTracer()
		eventsToTrigger *eventStack

		// Use a new variable to hold the standard context to avoid modify the initial
		// context.
		cc = c
	)

	if s.EnableTrace {
		eventsToTrigger = s.eventStackPool.Get().(*eventStack)
	}

	defer func() {
		if s.EnableTrace {
			// in case of error, we need to trigger all events
			if eventsToTrigger != nil {
				for last := eventsToTrigger.pop(); last != nil; last = eventsToTrigger.pop() {
					last(ctx.GetTraceInfo(), err)
				}
				s.eventStackPool.Put(eventsToTrigger)
			}
			if shouldRecordInTraceError(err) {
				traceCtl.DoFinish(cc, ctx, err)
			} else {
				traceCtl.DoFinish(cc, ctx, nil)
			}
		}

		// Hijack may release and close the connection already
		if zr != nil && !errors.Is(err, errs.ErrHijacked) {
			zr.Release() //nolint:errcheck
			zr = nil
		}

		if ctx.IsExiled() {
			return
		}

		s.putRequestContext(ctx)
	}()

	ctx.HTMLRender = s.HTMLRender
	ctx.SetConn(conn)
	ctx.Request.SetIsTLS(s.TLS != nil)
	ctx.SetEnableTrace(s.EnableTrace)

	if !s.NoDefaultServerHeader {
		serverName = s.ServerName
	}

	connRequestNum := uint64(0)

	for {
		connRequestNum++

		if zr == nil {
			zr = ctx.GetReader()
		}

		// If this is a keep-alive connection we want to try and read the first bytes
		// within the idle time.
		if connRequestNum > 1 {
			ctx.GetConn().SetReadTimeout(s.IdleTimeout) //nolint:errcheck

			_, err = zr.Peek(4)
			// This is not the first request, and we haven't read a single byte
			// of a new request yet. This means it's just a keep-alive connection
			// closing down either because the remote closed it or because
			// or a read timeout on our side. Either way just close the connection
			// and don't return any error response.
			if err != nil {
				err = errIdleTimeout
				return
			}

			// Reset the real read timeout for the coming request
			ctx.GetConn().SetReadTimeout(s.ReadTimeout) //nolint:errcheck
		}

		if s.EnableTrace {
			cc = traceCtl.DoStart(c, ctx)
			internalStats.Record(ctx.GetTraceInfo(), stats.ReadHeaderStart, err)
			eventsToTrigger.push(func(ti traceinfo.TraceInfo, err error) {
				internalStats.Record(ti, stats.ReadHeaderFinish, err)
			})
		}

		ctx.Response.Header.SetNoDefaultDate(s.NoDefaultDate)
		ctx.Response.Header.SetNoDefaultContentType(s.NoDefaultContentType)

		if s.DisableHeaderNamesNormalizing {
			ctx.Request.Header.DisableNormalizing()
			ctx.Response.Header.DisableNormalizing()
		}

		// Read Headers
		if err = req.ReadHeader(&ctx.Request.Header, zr); err == nil {
			if s.EnableTrace {
				// read header finished
				if last := eventsToTrigger.pop(); last != nil {
					last(ctx.GetTraceInfo(), err)
				}
				internalStats.Record(ctx.GetTraceInfo(), stats.ReadBodyStart, err)
				eventsToTrigger.push(func(ti traceinfo.TraceInfo, err error) {
					internalStats.Record(ti, stats.ReadBodyFinish, err)
				})
			}
			// Read body
			if s.StreamRequestBody {
				err = req.ReadBodyStream(&ctx.Request, zr, s.MaxRequestBodySize, s.GetOnly, !s.DisablePreParseMultipartForm)
			} else {
				err = req.ReadLimitBody(&ctx.Request, zr, s.MaxRequestBodySize, s.GetOnly, !s.DisablePreParseMultipartForm)
			}
		}

		if s.EnableTrace {
			if ctx.Request.Header.ContentLength() >= 0 {
				ctx.GetTraceInfo().Stats().SetRecvSize(len(ctx.Request.Header.RawHeaders()) + ctx.Request.Header.ContentLength())
			} else {
				ctx.GetTraceInfo().Stats().SetRecvSize(0)
			}
			// read body finished
			if last := eventsToTrigger.pop(); last != nil {
				last(ctx.GetTraceInfo(), err)
			}
		}

		if err != nil {
			if errors.Is(err, errs.ErrNothingRead) {
				return nil
			}

			if err == io.EOF {
				return errUnexpectedEOF
			}
			writeErrorResponse(zw, ctx, serverName, err)
			return
		}

		// 'Expect: 100-continue' request handling.
		// See https://www.w3.org/Protocols/rfc2616/rfc2616-sec8.html#sec8.2.3 for details.
		if ctx.Request.MayContinue() {
			// Allow the ability to deny reading the incoming request body
			if s.ContinueHandler != nil {
				if continueReadingRequest = s.ContinueHandler(&ctx.Request.Header); !continueReadingRequest {
					ctx.SetStatusCode(consts.StatusExpectationFailed)
				}
			}

			if continueReadingRequest {
				zw = ctx.GetWriter()
				// Send 'HTTP/1.1 100 Continue' response.
				_, err = zw.WriteBinary(bytestr.StrResponseContinue)
				if err != nil {
					return
				}
				err = zw.Flush()
				if err != nil {
					return
				}

				// Read body.
				if zr == nil {
					zr = ctx.GetReader()
				}
				if s.StreamRequestBody {
					err = req.ContinueReadBodyStream(&ctx.Request, zr, s.MaxRequestBodySize, !s.DisablePreParseMultipartForm)
				} else {
					err = req.ContinueReadBody(&ctx.Request, zr, s.MaxRequestBodySize, !s.DisablePreParseMultipartForm)
				}
				if err != nil {
					writeErrorResponse(zw, ctx, serverName, err)
					return
				}
			}
		}

		connectionClose = s.DisableKeepalive || ctx.Request.Header.ConnectionClose()
		isHTTP11 = ctx.Request.Header.IsHTTP11()

		if serverName != nil {
			ctx.Response.Header.SetServerBytes(serverName)
		}
		if s.EnableTrace {
			internalStats.Record(ctx.GetTraceInfo(), stats.ServerHandleStart, err)
			eventsToTrigger.push(func(ti traceinfo.TraceInfo, err error) {
				internalStats.Record(ti, stats.ServerHandleFinish, err)
			})
		}
		// Handle the request
		//
		// NOTE: All middlewares and business handler will be executed in this. And at this point, the request has been parsed
		// and the route has been matched.
		s.Core.ServeHTTP(cc, ctx)
		if s.EnableTrace {
			// application layer handle finished
			if last := eventsToTrigger.pop(); last != nil {
				last(ctx.GetTraceInfo(), err)
			}
		}

		// exit check
		if !s.Core.IsRunning() {
			connectionClose = true
		}

		if !ctx.IsGet() && ctx.IsHead() {
			ctx.Response.SkipBody = true
		}

		hijackHandler = ctx.GetHijackHandler()
		ctx.SetHijackHandler(nil)

		connectionClose = connectionClose || ctx.Response.ConnectionClose()
		if connectionClose {
			ctx.Response.Header.SetCanonical(bytestr.StrConnection, bytestr.StrClose)
		} else if !isHTTP11 {
			ctx.Response.Header.SetCanonical(bytestr.StrConnection, bytestr.StrKeepAlive)
		}

		if zw == nil {
			zw = ctx.GetWriter()
		}
		if s.EnableTrace {
			internalStats.Record(ctx.GetTraceInfo(), stats.WriteStart, err)
			eventsToTrigger.push(func(ti traceinfo.TraceInfo, err error) {
				internalStats.Record(ti, stats.WriteFinish, err)
			})
		}
		if err = writeResponse(ctx, zw); err != nil {
			return
		}

		if s.EnableTrace {
			if ctx.Response.Header.ContentLength() > 0 {
				ctx.GetTraceInfo().Stats().SetSendSize(ctx.Response.Header.GetHeaderLength() + ctx.Response.Header.ContentLength())
			} else {
				ctx.GetTraceInfo().Stats().SetSendSize(0)
			}
		}

		// Release the zeroCopyReader before flush to prevent data race
		if zr != nil {
			zr.Release() //nolint:errcheck
			zr = nil
		}
		// Flush the response.
		if err = zw.Flush(); err != nil {
			return
		}
		if s.EnableTrace {
			// write finished
			if last := eventsToTrigger.pop(); last != nil {
				last(ctx.GetTraceInfo(), err)
			}
		}

		// Release request body stream
		if ctx.Request.IsBodyStream() {
			err = ext.ReleaseBodyStream(ctx.RequestBodyStream())
			if err != nil {
				return
			}
		}

		if hijackHandler != nil {
			// Hijacked conn process the timeout by itself
			err = ctx.GetConn().SetReadTimeout(0)
			if err != nil {
				return
			}

			// Hijack and block the connection until the hijackHandler return
			s.HijackConnHandle(ctx.GetConn(), hijackHandler)
			err = errHijacked
			return
		}

		if connectionClose {
			return errShortConnection
		}
		// Back to network layer to trigger.
		// For now, only netpoll network mode has this feature.
		if s.IdleTimeout == 0 {
			return
		}
		// general case
		if s.EnableTrace {
			if shouldRecordInTraceError(err) {
				traceCtl.DoFinish(cc, ctx, err)
			} else {
				traceCtl.DoFinish(cc, ctx, nil)
			}
		}

		ctx.ResetWithoutConn()
	}
}

func NewServer() *Server {
	return &Server{
		eventStackPool: &sync.Pool{
			New: func() interface{} {
				return &eventStack{}
			},
		},
	}
}

func writeErrorResponse(zw network.Writer, ctx *app.RequestContext, serverName []byte, err error) network.Writer {
	errorHandler := defaultErrorHandler

	errorHandler(ctx, err)

	if serverName != nil {
		ctx.Response.Header.SetServerBytes(serverName)
	}
	ctx.SetConnectionClose()
	if zw == nil {
		zw = ctx.GetWriter()
	}
	writeResponse(ctx, zw) //nolint:errcheck
	zw.Flush()             //nolint:errcheck
	return zw
}

func writeResponse(ctx *app.RequestContext, w network.Writer) error {
	// Skip default response writing logic if it has been hijacked
	if ctx.Response.GetHijackWriter() != nil {
		return ctx.Response.GetHijackWriter().Finalize()
	}

	err := resp.Write(&ctx.Response, w)
	if err != nil {
		return err
	}

	return err
}

func defaultErrorHandler(ctx *app.RequestContext, err error) {
	if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
		ctx.AbortWithMsg("Request timeout", consts.StatusRequestTimeout)
	} else if errors.Is(err, errs.ErrBodyTooLarge) {
		ctx.AbortWithMsg("Request Entity Too Large", consts.StatusRequestEntityTooLarge)
	} else {
		ctx.AbortWithMsg("Error when parsing request", consts.StatusBadRequest)
	}
}

type eventStack []func(ti traceinfo.TraceInfo, err error)

func (e *eventStack) isEmpty() bool {
	return len(*e) == 0
}

func (e *eventStack) push(f func(ti traceinfo.TraceInfo, err error)) {
	*e = append(*e, f)
}

func (e *eventStack) pop() func(ti traceinfo.TraceInfo, err error) {
	if e.isEmpty() {
		return nil
	}
	last := (*e)[len(*e)-1]
	*e = (*e)[:len(*e)-1]
	return last
}

func shouldRecordInTraceError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, errs.ErrIdleTimeout) {
		return false
	}

	if errors.Is(err, errs.ErrHijacked) {
		return false
	}

	if errors.Is(err, errs.ErrShortConnection) {
		return false
	}

	return true
}
