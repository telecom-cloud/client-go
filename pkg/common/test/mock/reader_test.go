package mock

import (
	"bufio"
	"io"
	"testing"
	"testing/iotest"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

func TestZeroCopyReader(t *testing.T) {
	// raw
	r := "abcdef4343"
	zr := NewZeroCopyReader(r)
	rs := readBytes(zr.Reader)
	assert.DeepEqual(t, rs, r)

	// peek
	zr = NewZeroCopyReader(r)
	s, err := zr.Peek(1)
	assert.DeepEqual(t, nil, err)
	assert.DeepEqual(t, "a", string(s))

	s, err = zr.Peek(4)
	assert.DeepEqual(t, nil, err)
	assert.DeepEqual(t, "abcd", string(s))

	// https://github.com/golang/go/issues/50569
	ezr := NewZeroCopyReader("")
	s, err = ezr.Peek(32)
	assert.DeepEqual(t, io.EOF, err)
	assert.DeepEqual(t, "", string(s))

	// skip
	err = zr.Skip(1)
	assert.DeepEqual(t, nil, err)

	s, err = zr.Peek(4)
	assert.DeepEqual(t, nil, err)
	assert.DeepEqual(t, "bcde", string(s))

	// len
	assert.DeepEqual(t, len(r)-1, zr.Len())
	assert.DeepEqual(t, nil, ezr.Release())
	assert.Panic(t, func() { // not implement
		zr.ReadBinary(10)
	})
}

func TestEOFReader(t *testing.T) {
	r := &EOFReader{}
	s, err := r.Peek(1)
	assert.DeepEqual(t, io.EOF, err)
	assert.DeepEqual(t, "", string(s))
	assert.DeepEqual(t, nil, r.Skip(1))
	assert.DeepEqual(t, 0, r.Len())

	_, err = r.ReadByte()
	assert.DeepEqual(t, io.EOF, err)
	_, err = r.ReadBinary(10)
	assert.DeepEqual(t, io.EOF, err)
	_, err = r.Read(s)
	assert.DeepEqual(t, io.EOF, err)
	assert.DeepEqual(t, nil, r.Release())
}

func readBytes(buf *bufio.Reader) string {
	var b [1000]byte
	nb := 0
	for {
		c, err := buf.ReadByte()
		if err == io.EOF {
			break
		}
		if err == nil {
			b[nb] = c
			nb++
		} else if err != iotest.ErrTimeout {
			panic("Data: " + err.Error())
		}
	}
	return string(b[0:nb])
}
