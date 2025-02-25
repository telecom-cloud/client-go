package ut

import (
	"bytes"
	"testing"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

func TestCreateUtRequestContext(t *testing.T) {
	body := "1"
	method := "PUT"
	path := "/hey/dy"
	headerKey := "Connection"
	headerValue := "close"
	ctx := CreateUtRequestContext(method, path, &Body{bytes.NewBufferString(body), len(body)},
		Header{headerKey, headerValue})

	assert.DeepEqual(t, method, string(ctx.Method()))
	assert.DeepEqual(t, path, string(ctx.Path()))
	body1, err := ctx.Body()
	assert.DeepEqual(t, nil, err)
	assert.DeepEqual(t, body, string(body1))
	assert.DeepEqual(t, headerValue, string(ctx.GetHeader(headerKey)))
}
