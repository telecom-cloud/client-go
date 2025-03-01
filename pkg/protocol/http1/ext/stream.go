package ext

import (
	"bytes"
	"io"
	"sync"

	"github.com/telecom-cloud/client-go/internal/bytestr"
	"github.com/telecom-cloud/client-go/pkg/common/bytebufferpool"
	errs "github.com/telecom-cloud/client-go/pkg/common/errors"
	"github.com/telecom-cloud/client-go/pkg/common/utils"
	"github.com/telecom-cloud/client-go/pkg/network"
	"github.com/telecom-cloud/client-go/pkg/protocol"
)

var (
	errChunkedStream = errs.New(errs.ErrChunkedStream, errs.ErrorTypePublic, nil)

	bodyStreamPool = sync.Pool{
		New: func() interface{} {
			return &bodyStream{}
		},
	}
)

// Deprecated: Use github.com/telecom-cloud/client-go/pkg/protocol.NoBody instead.
var NoBody = protocol.NoBody

type bodyStream struct {
	prefetchedBytes *bytes.Reader
	reader          network.Reader
	trailer         *protocol.Trailer
	offset          int
	contentLength   int
	chunkLeft       int
	// whether the chunk has reached the EOF
	chunkEOF bool
}

func ReadBodyWithStreaming(zr network.Reader, contentLength, maxBodySize int, dst []byte) (b []byte, err error) {
	if contentLength == -1 {
		// handled in requestStream.Read()
		return b, errChunkedStream
	}
	dst = dst[:0]

	if maxBodySize <= 0 {
		maxBodySize = maxContentLengthInStream
	}
	readN := maxBodySize
	if readN > contentLength {
		readN = contentLength
	}
	if readN > maxContentLengthInStream {
		readN = maxContentLengthInStream
	}

	if contentLength >= 0 && maxBodySize >= contentLength {
		b, err = appendBodyFixedSize(zr, dst, readN)
	} else {
		b, err = readBodyIdentity(zr, readN, dst)
	}

	if err != nil {
		return b, err
	}
	if contentLength > maxBodySize {
		return b, errBodyTooLarge
	}
	return b, nil
}

func AcquireBodyStream(b *bytebufferpool.ByteBuffer, r network.Reader, t *protocol.Trailer, contentLength int) io.Reader {
	rs := bodyStreamPool.Get().(*bodyStream)
	rs.prefetchedBytes = bytes.NewReader(b.B)
	rs.reader = r
	rs.contentLength = contentLength
	rs.trailer = t
	rs.chunkEOF = false

	return rs
}

func (rs *bodyStream) Read(p []byte) (int, error) {
	defer func() {
		if rs.reader != nil {
			rs.reader.Release() //nolint:errcheck
		}
	}()
	if rs.contentLength == -1 {
		if rs.chunkEOF {
			return 0, io.EOF
		}

		if rs.chunkLeft == 0 {
			chunkSize, err := utils.ParseChunkSize(rs.reader)
			if err != nil {
				return 0, err
			}
			if chunkSize == 0 {
				err = ReadTrailer(rs.trailer, rs.reader)
				if err == nil {
					rs.chunkEOF = true
					err = io.EOF
				}
				return 0, err
			}

			rs.chunkLeft = chunkSize
		}
		bytesToRead := len(p)

		if bytesToRead > rs.chunkLeft {
			bytesToRead = rs.chunkLeft
		}

		src, err := rs.reader.Peek(bytesToRead)
		copied := copy(p, src)
		rs.reader.Skip(copied) // nolint: errcheck
		rs.chunkLeft -= copied

		if err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return copied, err
		}

		if rs.chunkLeft == 0 {
			err = utils.SkipCRLF(rs.reader)
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
		}

		return copied, err
	}
	if rs.offset == rs.contentLength {
		return 0, io.EOF
	}
	var n int
	var err error
	// read from the pre-read buffer
	if int(rs.prefetchedBytes.Size()) > rs.offset {
		n, err = rs.prefetchedBytes.Read(p)
		rs.offset += n
		if rs.offset == rs.contentLength {
			return n, io.EOF
		}
		if err != nil || len(p) == n {
			return n, err
		}
	}

	// read from the wire
	m := len(p) - n
	remain := rs.contentLength - rs.offset

	if m > remain {
		m = remain
	}

	if conn, ok := rs.reader.(io.Reader); ok {
		m, err = conn.Read(p[n:])
	} else {
		var tmp []byte
		tmp, err = rs.reader.Peek(m)
		m = copy(p[n:], tmp)
		rs.reader.Skip(m) // nolint: errcheck
	}
	rs.offset += m
	n += m

	if err != nil {
		// the data on stream may be incomplete
		if err == io.EOF {
			if rs.offset != rs.contentLength && rs.contentLength != -2 {
				err = io.ErrUnexpectedEOF
			}
			// ensure that skipRest works fine
			rs.offset = rs.contentLength
		}
		return n, err
	}
	if rs.offset == rs.contentLength {
		err = io.EOF
	}
	return n, err
}

func (rs *bodyStream) skipRest() error {
	// The body length doesn't exceed the maxContentLengthInStream or
	// the bodyStream has been skip rest
	if rs.prefetchedBytes == nil {
		return nil
	}

	// the request is chunked encoding
	if rs.contentLength == -1 {
		if rs.chunkEOF {
			return nil
		}

		strCRLFLen := len(bytestr.StrCRLF)
		for {
			chunkSize, err := utils.ParseChunkSize(rs.reader)
			if err != nil {
				return err
			}

			if chunkSize == 0 {
				rs.chunkEOF = true
				return SkipTrailer(rs.reader)
			}

			err = rs.reader.Skip(chunkSize)
			if err != nil {
				return err
			}

			crlf, err := rs.reader.Peek(strCRLFLen)
			if err != nil {
				return err
			}

			if !bytes.Equal(crlf, bytestr.StrCRLF) {
				return errBrokenChunk
			}

			err = rs.reader.Skip(strCRLFLen)
			if err != nil {
				return err
			}
		}
	}
	// max value of pSize is 8193, it's safe.
	pSize := int(rs.prefetchedBytes.Size())
	if rs.contentLength <= pSize || rs.offset == rs.contentLength {
		return nil
	}

	needSkipLen := 0
	if rs.offset > pSize {
		needSkipLen = rs.contentLength - rs.offset
	} else {
		needSkipLen = rs.contentLength - pSize
	}

	// must skip size
	for {
		skip := rs.reader.Len()
		if skip == 0 {
			_, err := rs.reader.Peek(1)
			if err != nil {
				return err
			}
			skip = rs.reader.Len()
		}
		if skip > needSkipLen {
			skip = needSkipLen
		}
		rs.reader.Skip(skip)
		needSkipLen -= skip
		if needSkipLen == 0 {
			return nil
		}
	}
}

// ReleaseBodyStream releases the body stream.
// Error of skipRest may be returned if there is one.
//
// NOTE: Be careful to use this method unless you know what it's for.
func ReleaseBodyStream(requestReader io.Reader) (err error) {
	if rs, ok := requestReader.(*bodyStream); ok {
		err = rs.skipRest()
		rs.reset()
		bodyStreamPool.Put(rs)
	}
	return
}

func (rs *bodyStream) reset() {
	rs.prefetchedBytes = nil
	rs.offset = 0
	rs.reader = nil
	rs.trailer = nil
	rs.chunkEOF = false
	rs.chunkLeft = 0
	rs.contentLength = 0
}
