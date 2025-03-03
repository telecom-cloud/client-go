package network

import (
	"testing"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

const (
	size1K = 1024
)

func TestConvertNetworkWriter(t *testing.T) {
	iw := &mockIOWriter{}
	w := NewWriter(iw)
	nw, _ := w.(*networkWriter)

	// Test malloc
	buf, _ := w.Malloc(size1K)
	assert.DeepEqual(t, len(buf), size1K)
	assert.DeepEqual(t, len(nw.caches), 1)
	assert.DeepEqual(t, len(nw.caches[0].data), size1K)
	assert.DeepEqual(t, cap(nw.caches[0].data), size1K)
	err := w.Flush()
	assert.Nil(t, err)
	assert.DeepEqual(t, size1K, iw.WriteNum)
	assert.DeepEqual(t, len(nw.caches), 0)
	assert.DeepEqual(t, cap(nw.caches), 1)

	// Test malloc left size
	buf, _ = w.Malloc(size1K + 1)
	assert.DeepEqual(t, len(buf), size1K+1)
	assert.DeepEqual(t, len(nw.caches), 1)
	assert.DeepEqual(t, len(nw.caches[0].data), size1K+1)
	assert.DeepEqual(t, cap(nw.caches[0].data), size1K*2)
	buf, _ = w.Malloc(size1K / 2)
	assert.DeepEqual(t, len(buf), size1K/2)
	assert.DeepEqual(t, len(nw.caches), 1)
	assert.DeepEqual(t, len(nw.caches[0].data), size1K+1+size1K/2)
	assert.DeepEqual(t, cap(nw.caches[0].data), size1K*2)
	buf, _ = w.Malloc(size1K / 2)
	assert.DeepEqual(t, len(buf), size1K/2)
	assert.DeepEqual(t, len(nw.caches), 2)
	assert.DeepEqual(t, len(nw.caches[0].data), size1K+1+size1K/2)
	assert.DeepEqual(t, cap(nw.caches[0].data), size1K*2)
	assert.DeepEqual(t, len(nw.caches[1].data), size1K/2)
	assert.DeepEqual(t, cap(nw.caches[1].data), size1K/2)
	err = w.Flush()
	assert.Nil(t, err)
	assert.DeepEqual(t, size1K*3+1, iw.WriteNum)
	assert.DeepEqual(t, len(nw.caches), 0)
	assert.DeepEqual(t, cap(nw.caches), 2)

	// Test WriteBinary after Malloc
	buf, _ = w.Malloc(size1K * 6)
	assert.DeepEqual(t, len(buf), size1K*6)
	assert.DeepEqual(t, len(nw.caches[0].data), size1K*6)
	b := make([]byte, size1K)
	w.WriteBinary(b)
	assert.DeepEqual(t, size1K*3+1, iw.WriteNum)
	assert.DeepEqual(t, len(nw.caches[0].data), size1K*7)
	assert.DeepEqual(t, cap(nw.caches[0].data), size1K*8)

	b = make([]byte, size1K*4)
	w.WriteBinary(b)
	assert.DeepEqual(t, len(nw.caches[1].data), size1K*4)
	assert.DeepEqual(t, cap(nw.caches[1].data), size1K*4)
	assert.DeepEqual(t, nw.caches[1].readOnly, true)
	w.Flush()
	assert.DeepEqual(t, size1K*14+1, iw.WriteNum)
}

type mockIOWriter struct {
	WriteNum int
}

func (m *mockIOWriter) Write(p []byte) (n int, err error) {
	m.WriteNum += len(p)
	return len(p), nil
}
