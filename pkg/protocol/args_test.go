package protocol

import (
	"testing"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

func TestArgsDeleteAll(t *testing.T) {
	t.Parallel()
	var a Args
	a.Add("q1", "foo")
	a.Add("q1", "bar")
	a.Add("q1", "baz")
	a.Add("q1", "quux")
	a.Add("q2", "1234")
	a.Del("q1")
	if a.Len() != 1 || a.Has("q1") {
		t.Fatalf("Expected q1 arg to be completely deleted. Current Args: %s", a.String())
	}
}

func TestArgsBytesOperation(t *testing.T) {
	var a Args
	a.Add("q1", "foo")
	a.Add("q2", "bar")
	setArgBytes(a.args, a.args[0].key, a.args[0].value, false)
	assert.DeepEqual(t, []byte("foo"), peekArgBytes(a.args, []byte("q1")))
	setArgBytes(a.args, a.args[1].key, a.args[1].value, true)
	assert.DeepEqual(t, []byte(""), peekArgBytes(a.args, []byte("q2")))
}

func TestArgsPeekExists(t *testing.T) {
	var a Args
	a.Add("q1", "foo")
	a.Add("", "")
	a.Add("?", "=")
	v1, b1 := a.PeekExists("q1")
	assert.DeepEqual(t, []byte("foo"), []byte(v1))
	assert.True(t, b1)
	v2, b2 := a.PeekExists("")
	assert.DeepEqual(t, []byte(""), []byte(v2))
	assert.True(t, b2)
	v3, b3 := a.PeekExists("q3")
	assert.DeepEqual(t, "", v3)
	assert.False(t, b3)
	v4, b4 := a.PeekExists("?")
	assert.DeepEqual(t, "=", v4)
	assert.True(t, b4)
}

func TestSetArg(t *testing.T) {
	a := Args{args: setArg(nil, "q1", "foo", true)}
	a.Add("", "")
	setArgBytes(a.args, []byte("q3"), []byte("bar"), false)
	s := a.String()
	assert.DeepEqual(t, []byte("q1&="), []byte(s))
}

// Test the encoding of special parameters
func TestArgsParseBytes(t *testing.T) {
	var ta1 Args
	ta1.Add("q1", "foo")
	ta1.Add("q1", "bar")
	ta1.Add("q2", "123")
	ta1.Add("q3", "")
	var a1 Args
	a1.ParseBytes([]byte("q1=foo&q1=bar&q2=123&q3="))
	assert.DeepEqual(t, &ta1, &a1)

	var ta2 Args
	ta2.Add("?", "foo")
	ta2.Add("&", "bar")
	ta2.Add("&", "?")
	ta2.Add("=", "=")
	var a2 Args
	a2.ParseBytes([]byte("%3F=foo&%26=bar&%26=%3F&%3D=%3D"))
	assert.DeepEqual(t, &ta2, &a2)
}

func TestArgsVisitAll(t *testing.T) {
	var a Args
	var s []string
	a.Add("cloudwego", "hertz")
	a.Add("hello", "world")
	a.VisitAll(func(key, value []byte) {
		s = append(s, string(key), string(value))
	})
	assert.DeepEqual(t, []string{"cloudwego", "hertz", "hello", "world"}, s)
}

func TestArgsPeekMulti(t *testing.T) {
	var a Args
	a.Add("cloudwego", "hertz")
	a.Add("cloudwego", "kitex")
	a.Add("cloudwego", "")
	a.Add("hello", "world")

	vv := a.PeekAll("cloudwego")
	expectedVV := [][]byte{
		[]byte("hertz"),
		[]byte("kitex"),
		[]byte(nil),
	}
	assert.DeepEqual(t, expectedVV, vv)

	vv = a.PeekAll("aaaa")
	assert.DeepEqual(t, 0, len(vv))

	vv = a.PeekAll("hello")
	expectedVV = [][]byte{[]byte("world")}
	assert.DeepEqual(t, expectedVV, vv)
}
