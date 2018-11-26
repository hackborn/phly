package phly

import (
	"errors"
	"strconv"
)

var (
	BadRequestErr      = errors.New("Bad request")
	emptyErr           = errors.New("Empty")
	wrongFormatPinsErr = errors.New("Pins in the wrong format")
)

// --------------------------------
// PHLY-ERROR

func NewBadRequestError(msg string) error {
	return &PhlyError{BadRequestErrCode, msg, nil}
}

func NewIllegalError(msg string) error {
	return &PhlyError{IllegalErrCode, msg, nil}
}

func NewMissingError(msg string) error {
	return &PhlyError{MissingErrCode, msg, nil}
}

func NewParseError(err error) error {
	return &PhlyError{ParseErrCode, "", err}
}

type PhlyError struct {
	code int
	msg  string
	err  error
}

func (e *PhlyError) ErrorCode() int {
	return e.code
}

func (e *PhlyError) Error() string {
	label := "Error"
	switch e.code {
	case BadRequestErrCode:
		label = "Bad request"
	case IllegalErrCode:
		label = "Illegal"
	case MissingErrCode:
		label = "Missing"
	case ParseErrCode:
		label = "Parse"
	}
	label += " (" + strconv.Itoa(e.code) + ")"
	if e.msg != "" {
		label += ": " + e.msg
	}
	if e.err != nil {
		label += ": " + e.err.Error()
	}
	return label
}

// --------------------------------
// MISC

// MergeErrors() answers the first non-nil error in the list.
func MergeErrors(err ...error) error {
	for _, a := range err {
		if a != nil {
			return a
		}
	}
	return nil
}

// ErrorsEqual() returns true if the errors are equivalent.
func ErrorsEqual(a, b error) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// Our internal error class only needs to match to the type
	aerr, aok := a.(*PhlyError)
	berr, bok := b.(*PhlyError)
	if aok && bok {
		return aerr.code == berr.code
	}
	return a.Error() == b.Error()
}

// --------------------------------
// CONST and VAR

const (
	BadRequestErrCode = 1000 + iota
	IllegalErrCode
	MissingErrCode
	ParseErrCode
)
