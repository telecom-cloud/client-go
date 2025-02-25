package mock

import (
	"bufio"
	"bytes"
	"io"
)

// ZeroCopyReader is used to create ZeroCopyReader for testing.
//
// NOTE: In principle, ut should use the zcReader created by netpoll.NewReader() for mock testing,
// but because zcReader does not implement the io.Reader interface, the test requirements of
// io.Reader involved are replaced with MockZeroCopyReader
type ZeroCopyReader struct {
	*bufio.Reader
}

func (m ZeroCopyReader) Peek(n int) ([]byte, error) {
	b, err := m.Reader.Peek(n)
	// if n is bigger than the buffer in m.Reader,
	// it will only return bufio.ErrBufferFull even if the underline reader return io.EOF.
	// so we make another Peek to get the real error.
	// for more info: https://github.com/golang/go/issues/50569
	if err == bufio.ErrBufferFull && len(b) == 0 {
		return m.Reader.Peek(1)
	}
	return b, err
}

func (m ZeroCopyReader) Skip(n int) (err error) {
	_, err = m.Reader.Discard(n)
	return
}

func (m ZeroCopyReader) Release() (err error) {
	return nil
}

func (m ZeroCopyReader) Len() (length int) {
	return m.Reader.Buffered()
}

func (m ZeroCopyReader) ReadBinary(n int) (p []byte, err error) {
	panic("implement me")
}

func NewZeroCopyReader(r string) ZeroCopyReader {
	br := bufio.NewReaderSize(bytes.NewBufferString(r), len(r))
	return ZeroCopyReader{br}
}

func NewLimitReader(r *bytes.Buffer) io.LimitedReader {
	return io.LimitedReader{
		R: r,
		N: int64(r.Len()),
	}
}

type EOFReader struct{}

func (e *EOFReader) Peek(n int) ([]byte, error) {
	return []byte{}, io.EOF
}

func (e *EOFReader) Skip(n int) error {
	return nil
}

func (e *EOFReader) Release() error {
	return nil
}

func (e *EOFReader) Len() int {
	return 0
}

func (e *EOFReader) ReadByte() (byte, error) {
	return ' ', io.EOF
}

func (e *EOFReader) ReadBinary(n int) (p []byte, err error) {
	return p, io.EOF
}

func (e *EOFReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}
