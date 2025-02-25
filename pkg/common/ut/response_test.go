package ut

import (
	"testing"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
	"github.com/telecom-cloud/client-go/pkg/protocol/consts"
)

func TestResult(t *testing.T) {
	r := new(ResponseRecorder)
	ret := r.Result()
	assert.DeepEqual(t, consts.StatusOK, ret.StatusCode())
}

func TestFlush(t *testing.T) {
	r := new(ResponseRecorder)
	r.Flush()
	ret := r.Result()
	assert.DeepEqual(t, consts.StatusOK, ret.StatusCode())
}

func TestWriterHeader(t *testing.T) {
	r := NewRecorder()
	r.WriteHeader(consts.StatusCreated)
	r.WriteHeader(consts.StatusOK)
	ret := r.Result()
	assert.DeepEqual(t, consts.StatusCreated, ret.StatusCode())
}

func TestWriteString(t *testing.T) {
	r := NewRecorder()
	r.WriteString("hello")
	ret := r.Result()
	assert.DeepEqual(t, "hello", string(ret.Body()))
}

func TestWrite(t *testing.T) {
	r := NewRecorder()
	r.Write([]byte("hello"))
	ret := r.Result()
	assert.DeepEqual(t, "hello", string(ret.Body()))
}
