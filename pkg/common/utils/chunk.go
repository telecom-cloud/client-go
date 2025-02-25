package utils

import (
	"bytes"
	"fmt"
	"io"

	"github.com/telecom-cloud/client-go/internal/bytesconv"
	"github.com/telecom-cloud/client-go/internal/bytestr"
	"github.com/telecom-cloud/client-go/pkg/common/errors"
	"github.com/telecom-cloud/client-go/pkg/network"
)

var errBrokenChunk = errors.NewPublic("cannot find crlf at the end of chunk")

func ParseChunkSize(r network.Reader) (int, error) {
	n, err := bytesconv.ReadHexInt(r)
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return -1, err
	}
	for {
		c, err := r.ReadByte()
		if err != nil {
			return -1, errors.NewPublic(fmt.Sprintf("cannot read '\r' char at the end of chunk size: %s", err))
		}
		// Skip any trailing whitespace after chunk size.
		if c == ' ' {
			continue
		}
		if c != '\r' {
			return -1, errors.NewPublic(
				fmt.Sprintf("unexpected char %q at the end of chunk size. Expected %q", c, '\r'),
			)
		}
		break
	}
	c, err := r.ReadByte()
	if err != nil {
		return -1, errors.NewPublic(fmt.Sprintf("cannot read '\n' char at the end of chunk size: %s", err))
	}
	if c != '\n' {
		return -1, errors.NewPublic(
			fmt.Sprintf("unexpected char %q at the end of chunk size. Expected %q", c, '\n'),
		)
	}
	return n, nil
}

// SkipCRLF will only skip the next CRLF("\r\n"), otherwise, error will be returned.
func SkipCRLF(reader network.Reader) error {
	p, err := reader.Peek(len(bytestr.StrCRLF))
	if err != nil {
		return err
	}
	if !bytes.Equal(p, bytestr.StrCRLF) {
		return errBrokenChunk
	}

	reader.Skip(len(p)) // nolint: errcheck
	return nil
}
