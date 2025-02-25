package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/telecom-cloud/client-go/internal/bytestr"
	"github.com/telecom-cloud/client-go/pkg/protocol"
)

var firstTime = true

type MockDoer struct {
	mock.Mock
}

func (m *MockDoer) Do(ctx context.Context, req *protocol.Request, resp *protocol.Response) error {

	// this is the real logic in (c *HostClient) doNonNilReqResp method
	if len(req.Header.Host()) == 0 {
		req.Header.SetHostBytes(req.URI().Host())
	}

	if firstTime {
		// req.Header.Host() is the real host writing to the wire
		if string(req.Header.Host()) != "example.com" {
			return errors.New("host not match")
		}
		// this is the real logic in (c *HostClient) doNonNilReqResp method
		if len(req.Header.Host()) == 0 {
			req.Header.SetHostBytes(req.URI().Host())
		}
		resp.Header.SetCanonical(bytestr.StrLocation, []byte("https://a.b.c/foo"))
		resp.SetStatusCode(301)
		firstTime = false
		return nil
	}

	if string(req.Header.Host()) != "a.b.c" {
		resp.SetStatusCode(400)
		return errors.New("host not match")
	}

	resp.SetStatusCode(200)

	return nil
}

func TestDoRequestFollowRedirects(t *testing.T) {
	mockDoer := new(MockDoer)
	mockDoer.On("Do", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	statusCode, _, err := DoRequestFollowRedirects(context.Background(), &protocol.Request{}, &protocol.Response{}, "https://example.com", defaultMaxRedirectsCount, mockDoer)
	assert.NoError(t, err)
	assert.Equal(t, 200, statusCode)
}
