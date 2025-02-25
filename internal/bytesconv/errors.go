package bytesconv

import "errors"

var (
	errEmptyInt               = errors.New("empty integer")
	errUnexpectedFirstChar    = errors.New("unexpected first char found. Expecting 0-9")
	errUnexpectedTrailingChar = errors.New("unexpected trailing char found. Expecting 0-9")
	errTooLongInt             = errors.New("too long int")
	errEmptyHexNum            = errors.New("empty hex number")
	errTooLargeHexNum         = errors.New("too large hex number")
)
