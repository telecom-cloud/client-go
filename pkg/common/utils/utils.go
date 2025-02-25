package utils

import (
	"bytes"
	"reflect"
	"runtime"
	"strings"

	"github.com/telecom-cloud/client-go/internal/bytesconv"
	errs "github.com/telecom-cloud/client-go/pkg/common/errors"
)

var errNeedMore = errs.New(errs.ErrNeedMore, errs.ErrorTypePublic, "cannot find trailing lf")

func Assert(guard bool, text string) {
	if !guard {
		panic(text)
	}
}

// H is a shortcut for map[string]interface{}
type H map[string]interface{}

func IsTrueString(str string) bool {
	return strings.ToLower(str) == "true"
}

func NameOfFunction(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func CaseInsensitiveCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i]|0x20 != b[i]|0x20 {
			return false
		}
	}
	return true
}

func NormalizeHeaderKey(b []byte, disableNormalizing bool) {
	if disableNormalizing {
		return
	}

	n := len(b)
	if n == 0 {
		return
	}

	b[0] = bytesconv.ToUpperTable[b[0]]
	for i := 1; i < n; i++ {
		p := &b[i]
		if *p == '-' {
			i++
			if i < n {
				b[i] = bytesconv.ToUpperTable[b[i]]
			}
			continue
		}
		*p = bytesconv.ToLowerTable[*p]
	}
}

func NextLine(b []byte) ([]byte, []byte, error) {
	nNext := bytes.IndexByte(b, '\n')
	if nNext < 0 {
		return nil, nil, errNeedMore
	}
	n := nNext
	if n > 0 && b[n-1] == '\r' {
		n--
	}
	return b[:n], b[nNext+1:], nil
}

func FilterContentType(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}
