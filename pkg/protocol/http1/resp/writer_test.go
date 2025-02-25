package resp

import (
	"strings"
	"testing"

	"github.com/telecom-cloud/client-go/internal/bytestr"
	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
	"github.com/telecom-cloud/client-go/pkg/common/test/mock"
	"github.com/telecom-cloud/client-go/pkg/protocol"
)

func TestNewChunkedBodyWriter(t *testing.T) {
	response := protocol.AcquireResponse()
	mockConn := mock.NewConn("")
	w := NewChunkedBodyWriter(response, mockConn)
	w.Write([]byte("hello"))
	w.Flush()
	out, _ := mockConn.WriterRecorder().ReadBinary(mockConn.WriterRecorder().WroteLen())
	assert.True(t, strings.Contains(string(out), "Transfer-Encoding: chunked"))
	assert.True(t, strings.Contains(string(out), "5"+string(bytestr.StrCRLF)+"hello"))
	assert.False(t, strings.Contains(string(out), "0"+string(bytestr.StrCRLF)+string(bytestr.StrCRLF)))
}

func TestNewChunkedBodyWriter1(t *testing.T) {
	response := protocol.AcquireResponse()
	mockConn := mock.NewConn("")
	w := NewChunkedBodyWriter(response, mockConn)
	w.Write([]byte("hello"))
	w.Flush()
	w.Finalize()
	w.Flush()
	out, _ := mockConn.WriterRecorder().ReadBinary(mockConn.WriterRecorder().WroteLen())
	assert.True(t, strings.Contains(string(out), "Transfer-Encoding: chunked"))
	assert.True(t, strings.Contains(string(out), "5"+string(bytestr.StrCRLF)+"hello"))
	assert.True(t, strings.Contains(string(out), "0"+string(bytestr.StrCRLF)+string(bytestr.StrCRLF)))
}

func TestNewChunkedBodyWriterNoData(t *testing.T) {
	response := protocol.AcquireResponse()
	response.Header.Set("Foo", "Bar")
	mockConn := mock.NewConn("")
	w := NewChunkedBodyWriter(response, mockConn)
	w.Finalize()
	w.Flush()
	out, _ := mockConn.WriterRecorder().ReadBinary(mockConn.WriterRecorder().WroteLen())
	assert.True(t, strings.Contains(string(out), "Transfer-Encoding: chunked"))
	assert.True(t, strings.Contains(string(out), "Foo: Bar"))
	assert.True(t, strings.Contains(string(out), "0"+string(bytestr.StrCRLF)+string(bytestr.StrCRLF)))
}
