package bytesconv

import (
	"net/http"
	"reflect"
	"sync"
	"time"
	"unsafe"

	"github.com/telecom-cloud/client-go/pkg/network"
)

const (
	upperhex = "0123456789ABCDEF"
	lowerhex = "0123456789abcdef"
)

var hexIntBufPool sync.Pool

func LowercaseBytes(b []byte) {
	for i := 0; i < len(b); i++ {
		p := &b[i]
		*p = ToLowerTable[*p]
	}
}

// B2s converts byte slice to a string without memory allocation.
// See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func B2s(b []byte) string {
	/* #nosec G103 */
	return *(*string)(unsafe.Pointer(&b))
}

// S2b converts string to a byte slice without memory allocation.
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func S2b(s string) (b []byte) {
	/* #nosec G103 */
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	/* #nosec G103 */
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return b
}

func WriteHexInt(w network.Writer, n int) error {
	if n < 0 {
		panic("BUG: int must be positive")
	}

	v := hexIntBufPool.Get()
	if v == nil {
		v = make([]byte, maxHexIntChars+1)
	}
	buf := v.([]byte)
	i := len(buf) - 1
	for {
		buf[i] = lowerhex[n&0xf]
		n >>= 4
		if n == 0 {
			break
		}
		i--
	}
	safeBuf, err := w.Malloc(maxHexIntChars + 1 - i)
	copy(safeBuf, buf[i:])
	hexIntBufPool.Put(v)
	return err
}

func ReadHexInt(r network.Reader) (int, error) {
	n := 0
	i := 0
	var k int
	for {
		buf, err := r.Peek(1)
		if err != nil {
			r.Skip(1)

			if i > 0 {
				return n, nil
			}
			return -1, err
		}

		c := buf[0]
		k = int(Hex2intTable[c])
		if k == 16 {
			if i == 0 {
				r.Skip(1)
				return -1, errEmptyHexNum
			}
			return n, nil
		}
		if i >= maxHexIntChars {
			r.Skip(1)
			return -1, errTooLargeHexNum
		}

		r.Skip(1)
		n = (n << 4) | k
		i++
	}
}

func ParseUintBuf(b []byte) (int, int, error) {
	n := len(b)
	if n == 0 {
		return -1, 0, errEmptyInt
	}
	v := 0
	for i := 0; i < n; i++ {
		c := b[i]
		k := c - '0'
		if k > 9 {
			if i == 0 {
				return -1, i, errUnexpectedFirstChar
			}
			return v, i, nil
		}
		vNew := 10*v + int(k)
		// Test for overflow.
		if vNew < v {
			return -1, i, errTooLongInt
		}
		v = vNew
	}
	return v, n, nil
}

// AppendUint appends n to dst and returns the extended dst.
func AppendUint(dst []byte, n int) []byte {
	if n < 0 {
		panic("BUG: int must be positive")
	}

	var b [20]byte
	buf := b[:]
	i := len(buf)
	var q int
	for n >= 10 {
		i--
		q = n / 10
		buf[i] = '0' + byte(n-q*10)
		n = q
	}
	i--
	buf[i] = '0' + byte(n)

	dst = append(dst, buf[i:]...)
	return dst
}

// AppendHTTPDate appends HTTP-compliant representation of date
// to dst and returns the extended dst.
func AppendHTTPDate(dst []byte, date time.Time) []byte {
	return date.UTC().AppendFormat(dst, http.TimeFormat)
}

func AppendQuotedPath(dst, src []byte) []byte {
	// Fix issue in https://github.com/golang/go/issues/11202
	if len(src) == 1 && src[0] == '*' {
		return append(dst, '*')
	}

	for _, c := range src {
		if QuotedPathShouldEscapeTable[int(c)] != 0 {
			dst = append(dst, '%', upperhex[c>>4], upperhex[c&15])
		} else {
			dst = append(dst, c)
		}
	}
	return dst
}

// AppendQuotedArg appends url-encoded src to dst and returns appended dst.
func AppendQuotedArg(dst, src []byte) []byte {
	for _, c := range src {
		switch {
		case c == ' ':
			dst = append(dst, '+')
		case QuotedArgShouldEscapeTable[int(c)] != 0:
			dst = append(dst, '%', upperhex[c>>4], upperhex[c&0xf])
		default:
			dst = append(dst, c)
		}
	}
	return dst
}

// ParseHTTPDate parses HTTP-compliant (RFC1123) date.
func ParseHTTPDate(date []byte) (time.Time, error) {
	return time.Parse(time.RFC1123, B2s(date))
}

// ParseUint parses uint from buf.
func ParseUint(buf []byte) (int, error) {
	v, n, err := ParseUintBuf(buf)
	if n != len(buf) {
		return -1, errUnexpectedTrailingChar
	}
	return v, err
}
