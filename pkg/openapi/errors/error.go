package errors

import (
	"errors"
	"fmt"
	"net/http"
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

var knownReasons = map[StatusReason]string{
	StatusReason_Unknown:               StatusReason_name[int32(StatusReason_Unknown)],
	StatusReason_Unauthorized:          StatusReason_name[int32(StatusReason_Unauthorized)],
	StatusReason_PaymentRequired:       StatusReason_name[int32(StatusReason_PaymentRequired)],
	StatusReason_Forbidden:             StatusReason_name[int32(StatusReason_Forbidden)],
	StatusReason_NotFound:              StatusReason_name[int32(StatusReason_NotFound)],
	StatusReason_AlreadyExists:         StatusReason_name[int32(StatusReason_AlreadyExists)],
	StatusReason_Conflict:              StatusReason_name[int32(StatusReason_Conflict)],
	StatusReason_Gone:                  StatusReason_name[int32(StatusReason_Gone)],
	StatusReason_Invalid:               StatusReason_name[int32(StatusReason_Invalid)],
	StatusReason_ServerTimeout:         StatusReason_name[int32(StatusReason_ServerTimeout)],
	StatusReason_Timeout:               StatusReason_name[int32(StatusReason_Timeout)],
	StatusReason_TooManyRequests:       StatusReason_name[int32(StatusReason_TooManyRequests)],
	StatusReason_BadRequest:            StatusReason_name[int32(StatusReason_BadRequest)],
	StatusReason_MethodNotAllowed:      StatusReason_name[int32(StatusReason_MethodNotAllowed)],
	StatusReason_NotAcceptable:         StatusReason_name[int32(StatusReason_NotAcceptable)],
	StatusReason_RequestEntityTooLarge: StatusReason_name[int32(StatusReason_RequestEntityTooLarge)],
	StatusReason_UnsupportedMediaType:  StatusReason_name[int32(StatusReason_UnsupportedMediaType)],
	StatusReason_InternalError:         StatusReason_name[int32(StatusReason_InternalError)],
	StatusReason_Expired:               StatusReason_name[int32(StatusReason_Expired)],
	StatusReason_ServiceUnavailable:    StatusReason_name[int32(StatusReason_ServiceUnavailable)],
}

func ReasonAndCodeForError(err error) (StatusReason, int32) {
	if status, ok := err.(ApiStatus); ok || errors.As(err, &status) {
		return StatusReason(StatusReason_value[status.Status().Reason]), status.Status().Code
	}
	return StatusReason_Unknown, 0
}

// IsUnauthorized returns true if the specified error was created by NewUnauthorized.
// It supports wrapped errors and returns false when the error is nil.
func IsUnauthorized(err error) bool {
	reason, code := ReasonAndCodeForError(err)
	if reason == StatusReason_Unauthorized {
		return true
	}
	if _, ok := knownReasons[reason]; !ok && code == http.StatusUnauthorized {
		return true
	}
	return false
}

// IsForbidden returns true if the specified error was created by NewForbidden.
// It supports wrapped errors and returns false when the error is nil.
func IsForbidden(err error) bool {
	reason, code := ReasonAndCodeForError(err)
	if reason == StatusReason_Forbidden {
		return true
	}
	if _, ok := knownReasons[reason]; !ok && code == http.StatusForbidden {
		return true
	}
	return false
}

// IsBadRequest returns true if the specified error was created by NewBadRequest.
// It supports wrapped errors and returns false when the error is nil.
func IsBadRequest(err error) bool {
	reason, code := ReasonAndCodeForError(err)
	if reason == StatusReason_BadRequest {
		return true
	}
	if _, ok := knownReasons[reason]; !ok && code == http.StatusBadRequest {
		return true
	}
	return false
}

// IsNotFound returns true if the specified error was created by NewNotFound.
// It supports wrapped errors and returns false when the error is nil.
func IsNotFound(err error) bool {
	reason, code := ReasonAndCodeForError(err)
	if reason == StatusReason_NotFound {
		return true
	}
	if _, ok := knownReasons[reason]; !ok && code == http.StatusNotFound {
		return true
	}
	return false
}

// IsAlreadyExists determines if err is an error which indicates that a specified resource already exists.
// It supports wrapped errors and returns false when the error is nil.
func IsAlreadyExists(err error) bool {
	reason, _ := ReasonAndCodeForError(err)
	return reason == StatusReason_AlreadyExists
}

// IsConflict determines if err is an error which indicates the provided update conflicts.
// It supports wrapped errors and returns false when the error is nil.
func IsConflict(err error) bool {
	reason, code := ReasonAndCodeForError(err)
	if reason == StatusReason_Conflict {
		return true
	}
	if _, ok := knownReasons[reason]; !ok && code == http.StatusConflict {
		return true
	}
	return false
}

// IsInvalid determines if err is an error which indicates the provided resource is not valid.
// It supports wrapped errors and returns false when the error is nil.
func IsInvalid(err error) bool {
	reason, code := ReasonAndCodeForError(err)
	if reason == StatusReason_Invalid {
		return true
	}
	if _, ok := knownReasons[reason]; !ok && code == http.StatusUnprocessableEntity {
		return true
	}
	return false
}

// IsGone is true if the error indicates the requested resource is no longer available.
// It supports wrapped errors and returns false when the error is nil.
func IsGone(err error) bool {
	reason, code := ReasonAndCodeForError(err)
	if reason == StatusReason_Gone {
		return true
	}
	if _, ok := knownReasons[reason]; !ok && code == http.StatusGone {
		return true
	}
	return false
}

// IsResourceExpired is true if the error indicates the resource has expired and the current action is
// no longer possible.
// It supports wrapped errors and returns false when the error is nil.
func IsResourceExpired(err error) bool {
	reason, _ := ReasonAndCodeForError(err)
	return reason == StatusReason_Expired
}

// IsTimeout is true if the error indicates the requested resource is no longer available.
// It supports wrapped errors and returns false when the error is nil.
func IsTimeout(err error) bool {
	reason, code := ReasonAndCodeForError(err)
	if reason == StatusReason_Timeout {
		return true
	}
	if _, ok := knownReasons[reason]; !ok && code == http.StatusGatewayTimeout {
		return true
	}
	return false
}
