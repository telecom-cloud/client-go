package utils

import (
	"io"

	"github.com/telecom-cloud/client-go/pkg/network"
)

func CopyBuffer(dst network.Writer, src io.Reader, buf []byte) (written int64, err error) {
	if buf != nil && len(buf) == 0 {
		panic("empty buffer in io.CopyBuffer")
	}
	return copyBuffer(dst, src, buf)
}

// copyBuffer is the actual implementation of Copy and CopyBuffer.
// If buf is nil, one is allocated.
func copyBuffer(dst network.Writer, src io.Reader, buf []byte) (written int64, err error) {
	if wt, ok := src.(io.WriterTo); ok {
		if w, ok := dst.(io.Writer); ok {
			return wt.WriteTo(w)
		}
	}

	// Sendfile impl
	if rf, ok := dst.(io.ReaderFrom); ok {
		return rf.ReadFrom(src)
	}

	if buf == nil {
		size := 32 * 1024
		if l, ok := src.(*io.LimitedReader); ok && int64(size) > l.N {
			if l.N < 1 {
				size = 1
			} else {
				size = int(l.N)
			}
		}
		buf = make([]byte, size)
	}
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, eb := dst.WriteBinary(buf[:nr])
			if eb != nil {
				err = eb
				return
			}

			if nw > 0 {
				written += int64(nw)
			}
			if nr != nw {
				err = io.ErrShortWrite
				return
			}
			if err = dst.Flush(); err != nil {
				return
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return
}

func CopyZeroAlloc(w network.Writer, r io.Reader) (int64, error) {
	vbuf := CopyBufPool.Get()
	buf := vbuf.([]byte)
	n, err := CopyBuffer(w, r, buf)
	CopyBufPool.Put(vbuf)
	return n, err
}
