package resp

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"sync"

	"github.com/telecom-cloud/client-go/pkg/common/bytebufferpool"
	errs "github.com/telecom-cloud/client-go/pkg/common/errors"
	"github.com/telecom-cloud/client-go/pkg/common/logger"
	"github.com/telecom-cloud/client-go/pkg/network"
	"github.com/telecom-cloud/client-go/pkg/protocol"
	"github.com/telecom-cloud/client-go/pkg/protocol/consts"
	"github.com/telecom-cloud/client-go/pkg/protocol/http1/ext"
)

// ErrBodyStreamWritePanic is returned when panic happens during writing body stream.
type ErrBodyStreamWritePanic struct {
	error
}

type h1Response struct {
	*protocol.Response
}

// String returns request representation.
//
// Returns error message instead of request representation on error.
//
// Use Write instead of String for performance-critical code.
func (h1Resp *h1Response) String() string {
	w := bytebufferpool.Get()
	zw := network.NewWriter(w)
	if err := Write(h1Resp.Response, zw); err != nil {
		return err.Error()
	}
	if err := zw.Flush(); err != nil {
		return err.Error()
	}
	s := string(w.B)
	bytebufferpool.Put(w)
	return s
}

func GetHTTP1Response(resp *protocol.Response) fmt.Stringer {
	return &h1Response{resp}
}

// ReadHeaderAndLimitBody reads response from the given r, limiting the body size.
//
// If maxBodySize > 0 and the body size exceeds maxBodySize,
// then ErrBodyTooLarge is returned.
//
// io.EOF is returned if r is closed before reading the first header byte.
func ReadHeaderAndLimitBody(resp *protocol.Response, r network.Reader, maxBodySize int) error {
	resp.ResetBody()
	err := ReadHeader(&resp.Header, r)
	if err != nil {
		return err
	}
	if resp.Header.StatusCode() == consts.StatusContinue {
		// Read the next response according to http://www.w3.org/Protocols/rfc2616/rfc2616-sec8.html .
		if err = ReadHeader(&resp.Header, r); err != nil {
			return err
		}
	}

	if !resp.MustSkipBody() {
		bodyBuf := resp.BodyBuffer()
		bodyBuf.Reset()
		bodyBuf.B, err = ext.ReadBody(r, resp.Header.ContentLength(), maxBodySize, bodyBuf.B)
		if err != nil {
			return err
		}
		if resp.Header.ContentLength() == -1 {
			err = ext.ReadTrailer(resp.Header.Trailer(), r)
			if err != nil && err != io.EOF {
				return err
			}
		}
		resp.Header.SetContentLength(len(bodyBuf.B))
	}

	return nil
}

type clientRespStream struct {
	r             io.Reader
	closeCallback func(shouldClose bool) error
}

func (c *clientRespStream) Close() (err error) {
	runtime.SetFinalizer(c, nil)
	// If error happened in release, the connection may be in abnormal state.
	// Close it in the callback in order to avoid other unexpected problems.
	err = ext.ReleaseBodyStream(c.r)
	shouldClose := false
	if err != nil {
		shouldClose = true
		logger.Warnf("connection will be closed instead of recycled because an error occurred during the stream body release: %s", err.Error())
	}
	if c.closeCallback != nil {
		err = c.closeCallback(shouldClose)
	}
	c.reset()
	return
}

func (c *clientRespStream) Read(p []byte) (n int, err error) {
	return c.r.Read(p)
}

func (c *clientRespStream) reset() {
	c.closeCallback = nil
	c.r = nil
	clientRespStreamPool.Put(c)
}

var clientRespStreamPool = sync.Pool{
	New: func() interface{} {
		return &clientRespStream{}
	},
}

func convertClientRespStream(bs io.Reader, fn func(shouldClose bool) error) *clientRespStream {
	clientStream := clientRespStreamPool.Get().(*clientRespStream)
	clientStream.r = bs
	clientStream.closeCallback = fn
	runtime.SetFinalizer(clientStream, (*clientRespStream).Close)
	return clientStream
}

// ReadBodyStream reads response body in stream
func ReadBodyStream(resp *protocol.Response, r network.Reader, maxBodySize int, closeCallBack func(shouldClose bool) error) error {
	resp.ResetBody()
	err := ReadHeader(&resp.Header, r)
	if err != nil {
		return err
	}

	if resp.Header.StatusCode() == consts.StatusContinue {
		// Read the next response according to http://www.w3.org/Protocols/rfc2616/rfc2616-sec8.html .
		if err = ReadHeader(&resp.Header, r); err != nil {
			return err
		}
	}

	if resp.MustSkipBody() {
		return nil
	}

	bodyBuf := resp.BodyBuffer()
	bodyBuf.Reset()
	bodyBuf.B, err = ext.ReadBodyWithStreaming(r, resp.Header.ContentLength(), maxBodySize, bodyBuf.B)
	if err != nil {
		if errors.Is(err, errs.ErrBodyTooLarge) {
			bodyStream := ext.AcquireBodyStream(bodyBuf, r, resp.Header.Trailer(), resp.Header.ContentLength())
			resp.ConstructBodyStream(bodyBuf, convertClientRespStream(bodyStream, closeCallBack))
			return nil
		}

		if errors.Is(err, errs.ErrChunkedStream) {
			bodyStream := ext.AcquireBodyStream(bodyBuf, r, resp.Header.Trailer(), -1)
			resp.ConstructBodyStream(bodyBuf, convertClientRespStream(bodyStream, closeCallBack))
			return nil
		}

		resp.Reset()
		return err
	}

	bodyStream := ext.AcquireBodyStream(bodyBuf, r, resp.Header.Trailer(), resp.Header.ContentLength())
	resp.ConstructBodyStream(bodyBuf, convertClientRespStream(bodyStream, closeCallBack))
	return nil
}

// Read reads response (including body) from the given r.
//
// io.EOF is returned if r is closed before reading the first header byte.
func Read(resp *protocol.Response, r network.Reader) error {
	return ReadHeaderAndLimitBody(resp, r, 0)
}

// Write writes response to w.
//
// Write doesn't flush response to w for performance reasons.
//
// See also WriteTo.
func Write(resp *protocol.Response, w network.Writer) error {
	sendBody := !resp.MustSkipBody()

	if resp.IsBodyStream() {
		return writeBodyStream(resp, w, sendBody)
	}

	body := resp.BodyBytes()
	bodyLen := len(body)
	if sendBody || bodyLen > 0 {
		resp.Header.SetContentLength(bodyLen)
	}

	header := resp.Header.Header()
	_, err := w.WriteBinary(header)
	if err != nil {
		return err
	}
	resp.Header.SetHeaderLength(len(header))
	// Write body
	if sendBody && bodyLen > 0 {
		_, err = w.WriteBinary(body)
	}
	return err
}

func writeBodyStream(resp *protocol.Response, w network.Writer, sendBody bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = &ErrBodyStreamWritePanic{
				error: fmt.Errorf("panic while writing body stream: %+v", r),
			}
		}
	}()

	contentLength := resp.Header.ContentLength()
	if contentLength < 0 {
		lrSize := ext.LimitedReaderSize(resp.BodyStream())
		if lrSize >= 0 {
			contentLength = int(lrSize)
			if int64(contentLength) != lrSize {
				contentLength = -1
			}
			if contentLength >= 0 {
				resp.Header.SetContentLength(contentLength)
			}
		}
	}
	if contentLength >= 0 {
		if err = WriteHeader(&resp.Header, w); err == nil && sendBody {
			if resp.ImmediateHeaderFlush {
				err = w.Flush()
			}
			if err == nil {
				err = ext.WriteBodyFixedSize(w, resp.BodyStream(), int64(contentLength))
			}
		}
	} else {
		resp.Header.SetContentLength(-1)
		if err = WriteHeader(&resp.Header, w); err == nil && sendBody {
			if resp.ImmediateHeaderFlush {
				err = w.Flush()
			}
			if err == nil {
				err = ext.WriteBodyChunked(w, resp.BodyStream())
			}
			if err == nil {
				err = ext.WriteTrailer(resp.Header.Trailer(), w)
			}
		}
	}
	err1 := resp.CloseBodyStream()
	if err == nil {
		err = err1
	}
	return err
}
