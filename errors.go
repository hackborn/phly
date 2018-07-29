package phly

import (
	"errors"
)

var (
	BadRequestErr            = errors.New("Bad request")
	corruptFileErr           = errors.New("Corrupt file")
	MissingDocErr            = errors.New("There is no doc")
	missingSourcesErr        = errors.New("There are no source nodes")
	unknownBlockTypeErr      = errors.New("Unknown block type")
	unfinishedPipelineErr    = errors.New("The pipeline hasn't finished but can't continue")
	unsupportedConversionErr = errors.New("Unsupported conversion")
	wrongFormatPinsErr       = errors.New("Pins in the wrong format")
	wrongMagicErr            = errors.New("Wrong magic")
)

func MergeErrors(err ...error) error {
	for _, a := range err {
		if a != nil {
			return a
		}
	}
	return nil
}
