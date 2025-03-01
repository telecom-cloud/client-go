package ext

import (
	"bytes"

	errs "github.com/telecom-cloud/client-go/pkg/common/errors"
	"github.com/telecom-cloud/client-go/pkg/common/utils"
)

var errInvalidName = errs.NewPublic("invalid header name")

type HeaderScanner struct {
	B     []byte
	Key   []byte
	Value []byte
	Err   error

	// HLen stores header subslice len
	HLen int

	DisableNormalizing bool

	// by checking whether the Next line contains a colon or not to tell
	// it's a header entry or a multi line value of current header entry.
	// the side effect of this operation is that we know the index of the
	// Next colon and new line, so this can be used during Next iteration,
	// instead of find them again.
	nextColon   int
	nextNewLine int

	initialized bool
}

type HeaderValueScanner struct {
	B     []byte
	Value []byte
}

func (s *HeaderScanner) Next() bool {
	if !s.initialized {
		s.nextColon = -1
		s.nextNewLine = -1
		s.initialized = true
	}
	bLen := len(s.B)
	if bLen >= 2 && s.B[0] == '\r' && s.B[1] == '\n' {
		s.B = s.B[2:]
		s.HLen += 2
		return false
	}
	if bLen >= 1 && s.B[0] == '\n' {
		s.B = s.B[1:]
		s.HLen++
		return false
	}
	var n int
	if s.nextColon >= 0 {
		n = s.nextColon
		s.nextColon = -1
	} else {
		n = bytes.IndexByte(s.B, ':')

		// There can't be a \n inside the header name, check for this.
		x := bytes.IndexByte(s.B, '\n')
		if x < 0 {
			// A header name should always at some point be followed by a \n
			// even if it's the one that terminates the header block.
			s.Err = errNeedMore
			return false
		}
		if x < n {
			// There was a \n before the :
			s.Err = errInvalidName
			return false
		}
	}
	if n < 0 {
		s.Err = errNeedMore
		return false
	}
	s.Key = s.B[:n]
	utils.NormalizeHeaderKey(s.Key, s.DisableNormalizing)
	n++
	for len(s.B) > n && s.B[n] == ' ' {
		n++
		// the newline index is a relative index, and lines below trimmed `s.b` by `n`,
		// so the relative newline index also shifted forward. it's safe to decrease
		// to a minus value, it means it's invalid, and will find the newline again.
		s.nextNewLine--
	}
	s.HLen += n
	s.B = s.B[n:]
	if s.nextNewLine >= 0 {
		n = s.nextNewLine
		s.nextNewLine = -1
	} else {
		n = bytes.IndexByte(s.B, '\n')
	}
	if n < 0 {
		s.Err = errNeedMore
		return false
	}
	isMultiLineValue := false
	for {
		if n+1 >= len(s.B) {
			break
		}
		if s.B[n+1] != ' ' && s.B[n+1] != '\t' {
			break
		}
		d := bytes.IndexByte(s.B[n+1:], '\n')
		if d <= 0 {
			break
		} else if d == 1 && s.B[n+1] == '\r' {
			break
		}
		e := n + d + 1
		if c := bytes.IndexByte(s.B[n+1:e], ':'); c >= 0 {
			s.nextColon = c
			s.nextNewLine = d - c - 1
			break
		}
		isMultiLineValue = true
		n = e
	}
	if n >= len(s.B) {
		s.Err = errNeedMore
		return false
	}
	oldB := s.B
	s.Value = s.B[:n]
	s.HLen += n + 1
	s.B = s.B[n+1:]

	if n > 0 && s.Value[n-1] == '\r' {
		n--
	}
	for n > 0 && s.Value[n-1] == ' ' {
		n--
	}
	s.Value = s.Value[:n]
	if isMultiLineValue {
		s.Value, s.B, s.HLen = normalizeHeaderValue(s.Value, oldB, s.HLen)
	}
	return true
}

func (s *HeaderValueScanner) next() bool {
	b := s.B
	if len(b) == 0 {
		return false
	}
	n := bytes.IndexByte(b, ',')
	if n < 0 {
		s.Value = stripSpace(b)
		s.B = b[len(b):]
		return true
	}
	s.Value = stripSpace(b[:n])
	s.B = b[n+1:]
	return true
}

func HasHeaderValue(s, value []byte) bool {
	var vs HeaderValueScanner
	vs.B = s
	for vs.next() {
		if utils.CaseInsensitiveCompare(vs.Value, value) {
			return true
		}
	}
	return false
}
