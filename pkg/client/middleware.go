package client

import (
	"context"

	"github.com/telecom-cloud/client-go/pkg/protocol"
)

// Endpoint represent one method for calling from remote.
type Endpoint func(ctx context.Context, req *protocol.Request, resp *protocol.Response) (err error)

// Middleware deal with input Endpoint and output Endpoint.
type Middleware func(Endpoint) Endpoint

// Chain connect middlewares into one middleware.
func chain(mws ...Middleware) Middleware {
	return func(next Endpoint) Endpoint {
		for i := len(mws) - 1; i >= 0; i-- {
			next = mws[i](next)
		}
		return next
	}
}
