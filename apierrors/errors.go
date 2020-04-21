package apierrors

import (
	"errors"
)

// A list of error messages for Observation API
var (
	ErrDatasetNotFound                   = errors.New("dataset not found")
	ErrEditionNotFound                   = errors.New("edition not found")
	ErrVersionNotFound                   = errors.New("version not found")
	ErrObservationsNotFound              = errors.New("no observations found")
	ErrTooManyWildcards                  = errors.New("only one wildcard (*) is allowed as a value in selected query parameters")
	ErrMissingVersionHeadersOrDimensions = errors.New("missing headers or dimensions or both from version doc")
	ErrIndexOutOfRange                   = errors.New("index out of range")
	ErrInternalServer                    = errors.New("internal error")
	ErrResourceState                     = errors.New("incorrect resource state")
)
