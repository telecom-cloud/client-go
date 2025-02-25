package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/telecom-cloud/client-go/pkg/common/bytebufferpool"
	"github.com/telecom-cloud/client-go/pkg/common/compress"
	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
	"github.com/telecom-cloud/client-go/pkg/common/test/mock"
	"github.com/telecom-cloud/client-go/pkg/network"
	"github.com/telecom-cloud/client-go/pkg/protocol/consts"
)

func TestResponseCopyTo(t *testing.T) {
	t.Parallel()

	var resp Response

	// empty copy
	testResponseCopyTo(t, &resp)

	// init resp
	// resp.laddr = zeroTCPAddr
	resp.SkipBody = true
	resp.Header.SetStatusCode(consts.StatusOK)
	resp.SetBodyString("test")
	testResponseCopyTo(t, &resp)
}

func TestResponseBodyStreamMultipleBodyCalls(t *testing.T) {
	t.Parallel()

	var r Response

	s := "foobar baz abc"
	if r.IsBodyStream() {
		t.Fatalf("IsBodyStream must return false")
	}
	r.SetBodyStream(bytes.NewBufferString(s), len(s))
	if !r.IsBodyStream() {
		t.Fatalf("IsBodyStream must return true")
	}
	for i := 0; i < 10; i++ {
		body := r.Body()
		if string(body) != s {
			t.Fatalf("unexpected body %q. Expecting %q. iteration %d", body, s, i)
		}
	}
}

func TestResponseBodyWriteToPlain(t *testing.T) {
	t.Parallel()

	var r Response

	expectedS := "foobarbaz"
	r.AppendBodyString(expectedS)

	testBodyWriteTo(t, &r, expectedS, true)
}

func TestResponseBodyWriteToStream(t *testing.T) {
	t.Parallel()

	var r Response

	expectedS := "aaabbbccc"
	buf := bytes.NewBufferString(expectedS)
	if r.IsBodyStream() {
		t.Fatalf("IsBodyStream must return false")
	}
	r.SetBodyStream(buf, len(expectedS))
	if !r.IsBodyStream() {
		t.Fatalf("IsBodyStream must return true")
	}

	testBodyWriteTo(t, &r, expectedS, false)
}

func TestResponseBodyWriter(t *testing.T) {
	t.Parallel()

	var r Response
	w := r.BodyWriter()
	for i := 0; i < 10; i++ {
		fmt.Fprintf(w, "%d", i)
	}
	if string(r.Body()) != "0123456789" {
		t.Fatalf("unexpected body %q. Expecting %q", r.Body(), "0123456789")
	}
}

func TestResponseRawBodySet(t *testing.T) {
	t.Parallel()

	var resp Response

	expectedS := "test"
	body := []byte(expectedS)
	resp.SetBodyRaw(body)

	testBodyWriteTo(t, &resp, expectedS, true)
}

func TestResponseRawBodyReset(t *testing.T) {
	t.Parallel()

	var resp Response

	body := []byte("test")
	resp.SetBodyRaw(body)
	resp.ResetBody()

	testBodyWriteTo(t, &resp, "", true)
}

func TestResponseResetBody(t *testing.T) {
	resp := Response{}
	resp.BodyBuffer()
	assert.NotNil(t, resp.body)
	resp.maxKeepBodySize = math.MaxUint32
	resp.ResetBody()
	assert.NotNil(t, resp.body)
	resp.maxKeepBodySize = -1
	resp.ResetBody()
	assert.Nil(t, resp.body)
}

func testResponseCopyTo(t *testing.T, src *Response) {
	var dst Response
	src.CopyTo(&dst)

	if !reflect.DeepEqual(src, &dst) { //nolint:govet
		t.Fatalf("ResponseCopyTo fail, src: \n%+v\ndst: \n%+v\n", src, &dst) //nolint:govet
	}
}

func TestResponseMustSkipBody(t *testing.T) {
	resp := Response{}
	resp.SetStatusCode(consts.StatusOK)
	resp.SetBodyString("test")
	assert.False(t, resp.MustSkipBody())
	// no content 204 means that skip body is necessary
	resp.SetStatusCode(consts.StatusNoContent)
	resp.ResetBody()
	assert.True(t, resp.MustSkipBody())
}

func TestResponseBodyGunzip(t *testing.T) {
	t.Parallel()
	dst1 := []byte("")
	src1 := []byte("hello")
	res1 := compress.AppendGzipBytes(dst1, src1)
	resp := Response{}
	resp.SetBody(res1)
	zipData, err := resp.BodyGunzip()
	assert.Nil(t, err)
	assert.DeepEqual(t, zipData, src1)
}

func TestResponseSwapResponseBody(t *testing.T) {
	t.Parallel()
	resp1 := Response{}
	str1 := "resp1"
	byteBuffer1 := &bytebufferpool.ByteBuffer{}
	byteBuffer1.Set([]byte(str1))
	resp1.ConstructBodyStream(byteBuffer1, bytes.NewBufferString(str1))
	assert.True(t, resp1.HasBodyBytes())
	resp2 := Response{}
	str2 := "resp2"
	byteBuffer2 := &bytebufferpool.ByteBuffer{}
	byteBuffer2.Set([]byte(str2))
	resp2.ConstructBodyStream(byteBuffer2, bytes.NewBufferString(str2))
	SwapResponseBody(&resp1, &resp2)
	assert.DeepEqual(t, resp1.body.B, []byte(str2))
	assert.DeepEqual(t, resp1.BodyStream(), bytes.NewBufferString(str2))
	assert.DeepEqual(t, resp2.body.B, []byte(str1))
	assert.DeepEqual(t, resp2.BodyStream(), bytes.NewBufferString(str1))
}

func TestResponseAcquireResponse(t *testing.T) {
	t.Parallel()
	resp1 := AcquireResponse()
	assert.NotNil(t, resp1)
	resp1.SetBody([]byte("test"))
	resp1.SetStatusCode(consts.StatusOK)
	ReleaseResponse(resp1)
	assert.Nil(t, resp1.body)
}

type closeBuffer struct {
	*bytes.Buffer
}

func (b *closeBuffer) Close() error {
	b.Reset()
	return nil
}

func TestSetBodyStreamNoReset(t *testing.T) {
	t.Parallel()
	resp := Response{}
	bsA := &closeBuffer{bytes.NewBufferString("A")}
	bsB := &closeBuffer{bytes.NewBufferString("B")}
	bsC := &closeBuffer{bytes.NewBufferString("C")}

	resp.SetBodyStream(bsA, 1)
	resp.SetBodyStreamNoReset(bsB, 1)
	// resp.Body() has closed bsB
	assert.DeepEqual(t, string(resp.Body()), "B")
	assert.DeepEqual(t, bsA.String(), "A")

	resp.bodyStream = bsA
	resp.SetBodyStream(bsC, 1)
	assert.DeepEqual(t, bsA.String(), "")
}

func TestRespSafeCopy(t *testing.T) {
	resp := AcquireResponse()
	resp.bodyRaw = make([]byte, 1)
	resps := make([]*Response, 10)
	for i := 0; i < 10; i++ {
		resp.bodyRaw[0] = byte(i)
		tmpResq := AcquireResponse()
		resp.CopyTo(tmpResq)
		resps[i] = tmpResq
	}
	for i := 0; i < 10; i++ {
		assert.DeepEqual(t, []byte{byte(i)}, resps[i].Body())
	}
}

func TestResponse_HijackWriter(t *testing.T) {
	resp := AcquireResponse()
	buf := new(bytes.Buffer)
	isFinal := false
	resp.HijackWriter(&mock.ExtWriter{Buf: buf, IsFinal: &isFinal})
	resp.AppendBody([]byte("hello"))
	assert.DeepEqual(t, 0, buf.Len())
	resp.GetHijackWriter().Flush()
	assert.DeepEqual(t, "hello", buf.String())
	resp.AppendBodyString(", world")
	assert.DeepEqual(t, "hello", buf.String())
	resp.GetHijackWriter().Flush()
	assert.DeepEqual(t, "hello, world", buf.String())
	resp.SetBody([]byte("hello, hertz"))
	resp.GetHijackWriter().Flush()
	assert.DeepEqual(t, "hello, hertz", buf.String())
	assert.False(t, isFinal)
	resp.GetHijackWriter().Finalize()
	assert.True(t, isFinal)
}

type HijackerFunc func() (network.Conn, error)

func (h HijackerFunc) Read(_ []byte) (int, error)    { return 0, errors.New("not implemented") }
func (h HijackerFunc) Hijack() (network.Conn, error) { return h() }

func TestResponse_Hijack(t *testing.T) {
	resp := AcquireResponse()
	defer ReleaseResponse(resp)

	_, err := resp.Hijack()
	assert.NotNil(t, err)

	resp.SetBodyStream(HijackerFunc(func() (network.Conn, error) { return nil, nil }), -1)
	_, err = resp.Hijack()
	assert.Nil(t, err)
}
