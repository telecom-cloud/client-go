package errors

import (
	"errors"
	"fmt"
)

type ApiStatus interface {
	Status() Status
}

type Status struct {
	Code    int32  `json:"statusCode"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
	Details *StatusDetails
}

type StatusDetails struct {
	Name              string
	Cause             error
	RetryAfterSeconds int32
}

type StatusError struct {
	ErrStatus Status
}

var _ error = (*StatusError)(nil)

// Error implements the Error interface.
func (e *StatusError) Error() string {
	return fmt.Sprintf("code: %d, reason: %s, message: %s", e.ErrStatus.Code, e.ErrStatus.Reason, e.ErrStatus.Message)
}

// Status allows access to e's status without having to know the detailed workings
// of StatusError.
func (e *StatusError) Status() Status {
	return e.ErrStatus
}

func (e *StatusError) Unwrap() error {
	if e.ErrStatus.Details != nil {
		return e.ErrStatus.Details.Cause
	}
	return nil
}

func (e *StatusError) Is(err error) bool {
	if se := new(StatusError); errors.As(err, &se) {
		return se.ErrStatus.Code == e.ErrStatus.Code && se.ErrStatus.Reason == e.ErrStatus.Reason
	}
	return false
}
