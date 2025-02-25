package ext

import (
	"errors"
	"fmt"
	"io"

	errs "github.com/telecom-cloud/client-go/pkg/common/errors"
)

var (
	errNeedMore     = errs.New(errs.ErrNeedMore, errs.ErrorTypePublic, "cannot find trailing lf")
	errBodyTooLarge = errs.New(errs.ErrBodyTooLarge, errs.ErrorTypePublic, "ext")
)

func HeaderError(typ string, err, errParse error, b []byte) error {
	if !errors.Is(errParse, errs.ErrNeedMore) {
		return headerErrorMsg(typ, errParse, b)
	}
	if err == nil {
		return errNeedMore
	}

	// Buggy servers may leave trailing CRLFs after http body.
	// Treat this case as EOF.
	if isOnlyCRLF(b) {
		return io.EOF
	}

	return headerErrorMsg(typ, err, b)
}

func headerErrorMsg(typ string, err error, b []byte) error {
	return errs.NewPublic(fmt.Sprintf("error when reading %s headers: %s. Buffer size=%d, contents: %s", typ, err, len(b), BufferSnippet(b)))
}
