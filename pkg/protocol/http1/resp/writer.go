package resp

import (
	"runtime"
	"sync"

	"github.com/telecom-cloud/client-go/pkg/network"
	"github.com/telecom-cloud/client-go/pkg/protocol"
	"github.com/telecom-cloud/client-go/pkg/protocol/http1/ext"
)

var chunkReaderPool sync.Pool

func init() {
	chunkReaderPool = sync.Pool{
		New: func() interface{} {
			return &chunkedBodyWriter{}
		},
	}
}

type chunkedBodyWriter struct {
	sync.Once
	finalizeErr error
	wroteHeader bool
	r           *protocol.Response
	w           network.Writer
}

// Write will encode chunked p before writing
// It will only return the length of p and a nil error if the writing is successful or 0, error otherwise.
//
// NOTE: Write will use the user buffer to flush.
// Before flush successfully, the buffer b should be valid.
func (c *chunkedBodyWriter) Write(p []byte) (n int, err error) {
	if !c.wroteHeader {
		c.r.Header.SetContentLength(-1)
		if err = WriteHeader(&c.r.Header, c.w); err != nil {
			return
		}
		c.wroteHeader = true
	}
	if err = ext.WriteChunk(c.w, p, false); err != nil {
		return
	}
	return len(p), nil
}

func (c *chunkedBodyWriter) Flush() error {
	return c.w.Flush()
}

// Finalize will write the ending chunk as well as trailer and flush the writer.
// Warning: do not call this method by yourself, unless you know what you are doing.
func (c *chunkedBodyWriter) Finalize() error {
	c.Do(func() {
		// in case no actual data from user
		if !c.wroteHeader {
			c.r.Header.SetContentLength(-1)
			if c.finalizeErr = WriteHeader(&c.r.Header, c.w); c.finalizeErr != nil {
				return
			}
			c.wroteHeader = true
		}
		c.finalizeErr = ext.WriteChunk(c.w, nil, true)
		if c.finalizeErr != nil {
			return
		}
		c.finalizeErr = ext.WriteTrailer(c.r.Header.Trailer(), c.w)
	})
	return c.finalizeErr
}

func (c *chunkedBodyWriter) release() {
	c.r = nil
	c.w = nil
	c.finalizeErr = nil
	c.wroteHeader = false
	chunkReaderPool.Put(c)
}

func NewChunkedBodyWriter(r *protocol.Response, w network.Writer) network.ExtWriter {
	extWriter := chunkReaderPool.Get().(*chunkedBodyWriter)
	extWriter.r = r
	extWriter.w = w
	extWriter.Once = sync.Once{}
	runtime.SetFinalizer(extWriter, (*chunkedBodyWriter).release)
	return extWriter
}
