package apierrors

import (
	"errors"
	"fmt"
)

// A list of error messages for Observation API
var (
	ErrDatasetNotFound          = errors.New("dataset not found")
	ErrEditionNotFound          = errors.New("edition not found")
	ErrVersionNotFound          = errors.New("version not found")
	ErrObservationsNotFound     = errors.New("no observations found")
	ErrTooManyWildcards         = errors.New("only one wildcard (*) is allowed as a value in selected query parameters")
	ErrMissingVersionDimensions = errors.New("missing list of dimensions from version doc")
	ErrIndexOutOfRange          = errors.New("index out of range")
	ErrInternalServer           = errors.New("internal error")
	ErrResourceState            = errors.New("incorrect resource state")
)

// ObservationQueryError is an error structure to handle observation query errors
type ObservationQueryError struct {
	message string
}

// Error returns the error message
func (e ObservationQueryError) Error() string {
	return e.message
}

// ErrorIncorrectQueryParameters returns an error for incorrect selection of query paramters
func ErrorIncorrectQueryParameters(params []string) error {
	return ObservationQueryError{
		message: fmt.Sprintf("incorrect selection of query parameters: %v, these dimensions do not exist for this version of the dataset", params),
	}
}

// ErrorMissingQueryParameters returns an error for missing parameters
func ErrorMissingQueryParameters(params []string) error {
	return ObservationQueryError{
		message: fmt.Sprintf("missing query parameters for the following dimensions: %v", params),
	}
}

// ErrorMultivaluedQueryParameters returns an error for multi-valued query parameters
func ErrorMultivaluedQueryParameters(params []string) error {
	return ObservationQueryError{
		message: fmt.Sprintf("multi-valued query parameters for the following dimensions: %v", params),
	}
}
