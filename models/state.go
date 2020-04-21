package models

import (
	errs "github.com/ONSdigital/dp-observation-api/apierrors"
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
)

var validVersionStates = map[string]int{
	EditionConfirmedState: 1,
	AssociatedState:       1,
	PublishedState:        1,
}

var validStates = map[string]int{
	CreatedState:          1,
	SubmittedState:        1,
	CompletedState:        1,
	EditionConfirmedState: 1,
	AssociatedState:       1,
	PublishedState:        1,
}

// CheckState checks state against a whitelist of valid states
func CheckState(docType, state string) error {
	var states map[string]int
	switch docType {
	case "version":
		states = validVersionStates
	default:
		states = validStates
	}

	if states[state] == 1 {
		return nil
	}

	return errs.ErrResourceState
}
