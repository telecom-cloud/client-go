package ut

import (
	"io"
	"io/ioutil"

	"github.com/telecom-cloud/client-go/pkg/app"
	"github.com/telecom-cloud/client-go/pkg/common/config"
	"github.com/telecom-cloud/client-go/pkg/protocol"
	"github.com/telecom-cloud/client-go/pkg/route"
)

// CreateUtRequestContext returns an app.RequestContext for testing purposes
func CreateUtRequestContext(method, url string, body *Body, headers ...Header) *app.RequestContext {
	engine := route.NewEngine(config.NewOptions([]config.Option{}))
	return createUtRequestContext(engine, method, url, body, headers...)
}

func createUtRequestContext(engine *route.Engine, method, url string, body *Body, headers ...Header) *app.RequestContext {
	ctx := engine.NewContext()

	var r *protocol.Request
	if body != nil && body.Body != nil {
		r = protocol.NewRequest(method, url, body.Body)
		r.CopyTo(&ctx.Request)
		if engine.IsStreamRequestBody() || body.Len == -1 {
			ctx.Request.SetBodyStream(body.Body, body.Len)
		} else {
			buf, err := ioutil.ReadAll(&io.LimitedReader{R: body.Body, N: int64(body.Len)})
			ctx.Request.SetBody(buf)
			if err != nil && err != io.EOF {
				panic(err)
			}
		}
	} else {
		r = protocol.NewRequest(method, url, nil)
		r.CopyTo(&ctx.Request)
	}

	for _, v := range headers {
		if ctx.Request.Header.Get(v.Key) != "" {
			ctx.Request.Header.Add(v.Key, v.Value)
		} else {
			ctx.Request.Header.Set(v.Key, v.Value)
		}
	}

	return ctx
}
