//go:build !amd64 && !arm64 && !ppc64
// +build !amd64,!arm64,!ppc64

package bytesconv

import (
	"fmt"
	"testing"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

func TestWriteHexInt(t *testing.T) {
	t.Parallel()

	for _, v := range []struct {
		s string
		n int
	}{
		{"0", 0},
		{"1", 1},
		{"123", 0x123},
		{"7fffffff", 0x7fffffff},
	} {
		testWriteHexInt(t, v.n, v.s)
	}
}

func TestReadHexInt(t *testing.T) {
	t.Parallel()

	for _, v := range []struct {
		s string
		n int
	}{
		//errTooLargeHexNum "too large hex number"
		//{"0123456789abcdef", -1},
		{"0", 0},
		{"fF", 0xff},
		{"00abc", 0xabc},
		{"7fffffff", 0x7fffffff},
		{"000", 0},
		{"1234ZZZ", 0x1234},
	} {
		testReadHexInt(t, v.s, v.n)
	}
}

func TestParseUint(t *testing.T) {
	t.Parallel()

	for _, v := range []struct {
		s string
		i int
	}{
		{"0", 0},
		{"123", 123},
		{"123456789", 123456789},
		{"2147483647", 2147483647},
	} {
		n, err := ParseUint(S2b(v.s))
		if err != nil {
			t.Errorf("unexpected error: %v. s=%q n=%v", err, v.s, n)
		}
		assert.DeepEqual(t, n, v.i)
	}
}

func TestParseUintError(t *testing.T) {
	t.Parallel()

	for _, v := range []struct {
		s string
	}{
		{""},
		{"cloudwego123"},
		{"1234.545"},
		{"-2147483648"},
		{"2147483648"},
		{"4294967295"},
	} {
		n, err := ParseUint(S2b(v.s))
		if err == nil {
			t.Fatalf("Expecting error when parsing %q. obtained %d", v.s, n)
		}
		if n >= 0 {
			t.Fatalf("Unexpected n=%d when parsing %q. Expected negative num", n, v.s)
		}
	}
}

func TestAppendUint(t *testing.T) {
	t.Parallel()

	for _, s := range []struct {
		n int
	}{
		{0},
		{123},
		{0x7fffffff},
	} {
		expectedS := fmt.Sprintf("%d", s.n)
		s := AppendUint(nil, s.n)
		assert.DeepEqual(t, expectedS, B2s(s))
	}
}
