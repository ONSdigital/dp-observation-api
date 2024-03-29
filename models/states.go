package models

import errs "github.com/ONSdigital/dp-observation-api/apierrors"

type docType int

// List of possible document types
const (
	Version docType = iota
	Dataset
)

// A list of reusable states across application
const (
	CreatedState          = "created"
	SubmittedState        = "submitted"
	CompletedState        = "completed"
	EditionConfirmedState = "edition-confirmed"
	AssociatedState       = "associated"
	PublishedState        = "published"
	DetachedState         = "detached"
	FailedState           = "failed"
)

var validDatasetStates = map[string]int{
	CreatedState:          1,
	SubmittedState:        1,
	CompletedState:        1,
	EditionConfirmedState: 1,
	AssociatedState:       1,
	PublishedState:        1,
	FailedState:           1,
}

var validVersionStates = map[string]int{
	EditionConfirmedState: 1,
	AssociatedState:       1,
	PublishedState:        1,
}

// CheckState checks state against a whitelist of valid states
func CheckState(doc docType, state string) error {
	var states map[string]int
	switch doc {
	case Version:
		states = validVersionStates
	case Dataset:
		states = validDatasetStates
	default:
		return errs.ErrInvalidDocType
	}

	if states[state] == 1 {
		return nil
	}

	return errs.ErrResourceState
}
