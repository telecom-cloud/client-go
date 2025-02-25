package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"sync"

	"github.com/telecom-cloud/client-go/pkg/common/bytebufferpool"
	"github.com/telecom-cloud/client-go/pkg/common/stackless"
	"github.com/telecom-cloud/client-go/pkg/common/utils"
	"github.com/telecom-cloud/client-go/pkg/network"
)

const CompressDefaultCompression = 6 // flate.DefaultCompression

var gzipReaderPool sync.Pool

var (
	stacklessGzipWriterPoolMap = newCompressWriterPoolMap()
	realGzipWriterPoolMap      = newCompressWriterPoolMap()
)

func newCompressWriterPoolMap() []*sync.Pool {
	// Initialize pools for all the compression levels defined
	// in https://golang.org/pkg/compress/flate/#pkg-constants .
	// Compression levels are normalized with normalizeCompressLevel,
	// so the fit [0..11].
	var m []*sync.Pool
	for i := 0; i < 12; i++ {
		m = append(m, &sync.Pool{})
	}
	return m
}

type compressCtx struct {
	w     io.Writer
	p     []byte
	level int
}

// AppendGunzipBytes appends gunzipped src to dst and returns the resulting dst.
func AppendGunzipBytes(dst, src []byte) ([]byte, error) {
	w := &byteSliceWriter{dst}
	_, err := WriteGunzip(w, src)
	return w.b, err
}

type byteSliceWriter struct {
	b []byte
}

func (w *byteSliceWriter) Write(p []byte) (int, error) {
	w.b = append(w.b, p...)
	return len(p), nil
}

// WriteGunzip writes gunzipped p to w and returns the number of uncompressed
// bytes written to w.
func WriteGunzip(w io.Writer, p []byte) (int, error) {
	r := &byteSliceReader{p}
	zr, err := AcquireGzipReader(r)
	if err != nil {
		return 0, err
	}
	zw := network.NewWriter(w)
	n, err := utils.CopyZeroAlloc(zw, zr)
	ReleaseGzipReader(zr)
	nn := int(n)
	if int64(nn) != n {
		return 0, fmt.Errorf("too much data gunzipped: %d", n)
	}
	return nn, err
}

type byteSliceReader struct {
	b []byte
}

func (r *byteSliceReader) Read(p []byte) (int, error) {
	if len(r.b) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.b)
	r.b = r.b[n:]
	return n, nil
}

func AcquireGzipReader(r io.Reader) (*gzip.Reader, error) {
	v := gzipReaderPool.Get()
	if v == nil {
		return gzip.NewReader(r)
	}
	zr := v.(*gzip.Reader)
	if err := zr.Reset(r); err != nil {
		return nil, err
	}
	return zr, nil
}

func ReleaseGzipReader(zr *gzip.Reader) {
	zr.Close()
	gzipReaderPool.Put(zr)
}

// AppendGzipBytes appends gzipped src to dst and returns the resulting dst.
func AppendGzipBytes(dst, src []byte) []byte {
	return AppendGzipBytesLevel(dst, src, CompressDefaultCompression)
}

// AppendGzipBytesLevel appends gzipped src to dst using the given
// compression level and returns the resulting dst.
//
// Supported compression levels are:
//
//   - CompressNoCompression
//   - CompressBestSpeed
//   - CompressBestCompression
//   - CompressDefaultCompression
//   - CompressHuffmanOnly
func AppendGzipBytesLevel(dst, src []byte, level int) []byte {
	w := &byteSliceWriter{dst}
	WriteGzipLevel(w, src, level) //nolint:errcheck
	return w.b
}

var stacklessWriteGzip = stackless.NewFunc(nonblockingWriteGzip)

func nonblockingWriteGzip(ctxv interface{}) {
	ctx := ctxv.(*compressCtx)
	zw := acquireRealGzipWriter(ctx.w, ctx.level)

	_, err := zw.Write(ctx.p)
	if err != nil {
		panic(fmt.Sprintf("BUG: gzip.Writer.Write for len(p)=%d returned unexpected error: %s", len(ctx.p), err))
	}

	releaseRealGzipWriter(zw, ctx.level)
}

func releaseRealGzipWriter(zw *gzip.Writer, level int) {
	zw.Close()
	nLevel := normalizeCompressLevel(level)
	p := realGzipWriterPoolMap[nLevel]
	p.Put(zw)
}

func acquireRealGzipWriter(w io.Writer, level int) *gzip.Writer {
	nLevel := normalizeCompressLevel(level)
	p := realGzipWriterPoolMap[nLevel]
	v := p.Get()
	if v == nil {
		zw, err := gzip.NewWriterLevel(w, level)
		if err != nil {
			panic(fmt.Sprintf("BUG: unexpected error from gzip.NewWriterLevel(%d): %s", level, err))
		}
		return zw
	}
	zw := v.(*gzip.Writer)
	zw.Reset(w)
	return zw
}

// normalizes compression level into [0..11], so it could be used as an index
// in *PoolMap.
func normalizeCompressLevel(level int) int {
	// -2 is the lowest compression level - CompressHuffmanOnly
	// 9 is the highest compression level - CompressBestCompression
	if level < -2 || level > 9 {
		level = CompressDefaultCompression
	}
	return level + 2
}

// WriteGzipLevel writes gzipped p to w using the given compression level
// and returns the number of compressed bytes written to w.
//
// Supported compression levels are:
//
//   - CompressNoCompression
//   - CompressBestSpeed
//   - CompressBestCompression
//   - CompressDefaultCompression
//   - CompressHuffmanOnly
func WriteGzipLevel(w io.Writer, p []byte, level int) (int, error) {
	switch w.(type) {
	case *byteSliceWriter,
		*bytes.Buffer,
		*bytebufferpool.ByteBuffer:
		// These writers don't block, so we can just use stacklessWriteGzip
		ctx := &compressCtx{
			w:     w,
			p:     p,
			level: level,
		}
		stacklessWriteGzip(ctx)
		return len(p), nil
	default:
		zw := AcquireStacklessGzipWriter(w, level)
		n, err := zw.Write(p)
		ReleaseStacklessGzipWriter(zw, level)
		return n, err
	}
}

func AcquireStacklessGzipWriter(w io.Writer, level int) stackless.Writer {
	nLevel := normalizeCompressLevel(level)
	p := stacklessGzipWriterPoolMap[nLevel]
	v := p.Get()
	if v == nil {
		return stackless.NewWriter(w, func(w io.Writer) stackless.Writer {
			return acquireRealGzipWriter(w, level)
		})
	}
	sw := v.(stackless.Writer)
	sw.Reset(w)
	return sw
}

func ReleaseStacklessGzipWriter(sw stackless.Writer, level int) {
	sw.Close()
	nLevel := normalizeCompressLevel(level)
	p := stacklessGzipWriterPoolMap[nLevel]
	p.Put(sw)
}
