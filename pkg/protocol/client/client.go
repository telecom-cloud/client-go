package client

import (
	"context"
	"sync"
	"time"

	"github.com/telecom-cloud/client-go/internal/bytestr"
	"github.com/telecom-cloud/client-go/pkg/common/config"
	"github.com/telecom-cloud/client-go/pkg/common/errors"
	"github.com/telecom-cloud/client-go/pkg/common/timer"
	"github.com/telecom-cloud/client-go/pkg/protocol"
	"github.com/telecom-cloud/client-go/pkg/protocol/consts"
)

const defaultMaxRedirectsCount = 16

var (
	errTimeout          = errors.New(errors.ErrTimeout, errors.ErrorTypePublic, "host client")
	errMissingLocation  = errors.NewPublic("missing Location header for http redirect")
	errTooManyRedirects = errors.NewPublic("too many redirects detected when doing the request")

	clientURLResponseChPool sync.Pool
)

type HostClient interface {
	Doer
	SetDynamicConfig(dc *DynamicConfig)
	CloseIdleConnections()
	ShouldRemove() bool
	ConnectionCount() int
}

type Doer interface {
	Do(ctx context.Context, req *protocol.Request, resp *protocol.Response) error
}

// DefaultRetryIf Default retry condition, mainly used for idempotent requests.
// If this cannot be satisfied, you can implement your own retry condition.
func DefaultRetryIf(req *protocol.Request, resp *protocol.Response, err error) bool {
	// cannot retry if the request body is not rewindable
	if req.IsBodyStream() {
		return false
	}

	if isIdempotent(req, resp, err) {
		return true
	}

	return false
}

func isIdempotent(req *protocol.Request, resp *protocol.Response, err error) bool {
	return req.Header.IsGet() ||
		req.Header.IsHead() ||
		req.Header.IsPut() ||
		req.Header.IsDelete() ||
		req.Header.IsOptions() ||
		req.Header.IsTrace()
}

// DynamicConfig is config set which will be confirmed when starts a request.
type DynamicConfig struct {
	Addr     string
	ProxyURI *protocol.URI
	IsTLS    bool
}

// RetryIfFunc signature of retry if function
// Judge whether to retry by request,response or error , return true is retry
type RetryIfFunc func(req *protocol.Request, resp *protocol.Response, err error) bool

type clientURLResponse struct {
	statusCode int
	body       []byte
	err        error
}

func GetURL(ctx context.Context, dst []byte, url string, c Doer, requestOptions ...config.RequestOption) (statusCode int, body []byte, err error) {
	req := protocol.AcquireRequest()
	req.SetOptions(requestOptions...)

	statusCode, body, err = doRequestFollowRedirectsBuffer(ctx, req, dst, url, c)

	protocol.ReleaseRequest(req)
	return statusCode, body, err
}

func GetURLTimeout(ctx context.Context, dst []byte, url string, timeout time.Duration, c Doer, requestOptions ...config.RequestOption) (statusCode int, body []byte, err error) {
	deadline := time.Now().Add(timeout)
	return GetURLDeadline(ctx, dst, url, deadline, c, requestOptions...)
}

func GetURLDeadline(ctx context.Context, dst []byte, url string, deadline time.Time, c Doer, requestOptions ...config.RequestOption) (statusCode int, body []byte, err error) {
	timeout := -time.Since(deadline)
	if timeout <= 0 {
		return 0, dst, errTimeout
	}

	var ch chan clientURLResponse
	chv := clientURLResponseChPool.Get()
	if chv == nil {
		chv = make(chan clientURLResponse, 1)
	}
	ch = chv.(chan clientURLResponse)

	req := protocol.AcquireRequest()
	req.SetOptions(requestOptions...)

	// Note that the request continues execution on errTimeout until
	// client-specific ReadTimeout exceeds. This helps to limit load
	// on slow hosts by MaxConns* concurrent requests.
	//
	// Without this 'hack' the load on slow host could exceed MaxConns*
	// concurrent requests, since timed out requests on client side
	// usually continue execution on the host.
	go func() {
		statusCodeCopy, bodyCopy, errCopy := doRequestFollowRedirectsBuffer(ctx, req, dst, url, c)
		ch <- clientURLResponse{
			statusCode: statusCodeCopy,
			body:       bodyCopy,
			err:        errCopy,
		}
	}()

	tc := timer.AcquireTimer(timeout)
	select {
	case resp := <-ch:
		protocol.ReleaseRequest(req)
		clientURLResponseChPool.Put(chv)
		statusCode = resp.statusCode
		body = resp.body
		err = resp.err
	case <-tc.C:
		body = dst
		err = errTimeout
	}
	timer.ReleaseTimer(tc)

	return statusCode, body, err
}

func PostURL(ctx context.Context, dst []byte, url string, postArgs *protocol.Args, c Doer, requestOptions ...config.RequestOption) (statusCode int, body []byte, err error) {
	req := protocol.AcquireRequest()
	req.Header.SetMethodBytes(bytestr.StrPost)
	req.Header.SetContentTypeBytes(bytestr.StrPostArgsContentType)
	req.SetOptions(requestOptions...)

	if postArgs != nil {
		if _, err := postArgs.WriteTo(req.BodyWriter()); err != nil {
			return 0, nil, err
		}
	}

	statusCode, body, err = doRequestFollowRedirectsBuffer(ctx, req, dst, url, c)

	protocol.ReleaseRequest(req)
	return statusCode, body, err
}

func doRequestFollowRedirectsBuffer(ctx context.Context, req *protocol.Request, dst []byte, url string, c Doer) (statusCode int, body []byte, err error) {
	resp := protocol.AcquireResponse()
	bodyBuf := resp.BodyBuffer()
	oldBody := bodyBuf.B
	bodyBuf.B = dst

	statusCode, _, err = DoRequestFollowRedirects(ctx, req, resp, url, defaultMaxRedirectsCount, c)

	// In HTTP2 scenario, client use stream mode to create a request and its body is in body stream.
	// In HTTP1, only client recv body exceed max body size and client is in stream mode can trig it.
	body = resp.Body()
	bodyBuf.B = oldBody
	protocol.ReleaseResponse(resp)

	return statusCode, body, err
}

func DoRequestFollowRedirects(ctx context.Context, req *protocol.Request, resp *protocol.Response, url string, maxRedirectsCount int, c Doer) (statusCode int, body []byte, err error) {
	redirectsCount := 0

	for {
		req.SetRequestURI(url)
		req.ParseURI()

		if err = c.Do(ctx, req, resp); err != nil {
			break
		}
		statusCode = resp.Header.StatusCode()
		if !StatusCodeIsRedirect(statusCode) {
			break
		}

		redirectsCount++
		if redirectsCount > maxRedirectsCount {
			err = errTooManyRedirects
			break
		}
		location := resp.Header.PeekLocation()
		if len(location) == 0 {
			err = errMissingLocation
			break
		}
		url = getRedirectURL(url, location)

		// Remove the former host header.
		req.Header.Del(consts.HeaderHost)
	}

	return statusCode, body, err
}

// StatusCodeIsRedirect returns true if the status code indicates a redirect.
func StatusCodeIsRedirect(statusCode int) bool {
	return statusCode == consts.StatusMovedPermanently ||
		statusCode == consts.StatusFound ||
		statusCode == consts.StatusSeeOther ||
		statusCode == consts.StatusTemporaryRedirect ||
		statusCode == consts.StatusPermanentRedirect
}

func getRedirectURL(baseURL string, location []byte) string {
	u := protocol.AcquireURI()
	u.Update(baseURL)
	u.UpdateBytes(location)
	redirectURL := u.String()
	protocol.ReleaseURI(u)
	return redirectURL
}

func DoTimeout(ctx context.Context, req *protocol.Request, resp *protocol.Response, timeout time.Duration, c Doer) error {
	if timeout <= 0 {
		return errTimeout
	}
	// Note: it will overwrite the reqTimeout.
	req.SetOptions(config.WithRequestTimeout(timeout))
	return c.Do(ctx, req, resp)
}

func DoDeadline(ctx context.Context, req *protocol.Request, resp *protocol.Response, deadline time.Time, c Doer) error {
	timeout := time.Until(deadline)
	if timeout <= 0 {
		return errTimeout
	}
	// Note: it will overwrite the reqTimeout.
	req.SetOptions(config.WithRequestTimeout(timeout))
	return c.Do(ctx, req, resp)
}
