package phly

import (
	"errors"
)

var (
	BadRequestErr            = errors.New("Bad request")
	corruptFileErr           = errors.New("Corrupt file")
	emptyErr                 = errors.New("Empty")
	MissingDocErr            = errors.New("There is no doc")
	missingSourcesErr        = errors.New("There are no source nodes")
	unknownBlockTypeErr      = errors.New("Unknown block type")
	unfinishedPipelineErr    = errors.New("The pipeline hasn't finished but can't continue")
	unsupportedConversionErr = errors.New("Unsupported conversion")
	unsupportedInputErr      = errors.New("Unsupported input")
	wrongFormatPinsErr       = errors.New("Pins in the wrong format")
	wrongMagicErr            = errors.New("Wrong magic")
)

// --------------------------------
// ERR-CODE

type ErrCode interface {
	ErrorCode() int
	Error() string
}

func NewErrCode(errcode int, errmsg string) ErrCode {
	return &errCode{code: errcode, msg: errmsg}
}

type errCode struct {
	code int
	msg  string
}

func (e *errCode) ErrorCode() int {
	return e.code
}

func (e *errCode) Error() string {
	return e.msg
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
