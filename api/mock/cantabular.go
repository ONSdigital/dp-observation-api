// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-observation-api/api"
	"sync"
)

// Ensure, that CantabularClientMock does implement api.CantabularClient.
// If this is not the case, regenerate this file with moq.
var _ api.CantabularClient = &CantabularClientMock{}

// CantabularClientMock is a mock implementation of api.CantabularClient.
//
// 	func TestSomethingThatUsesCantabularClient(t *testing.T) {
//
// 		// make and configure a mocked api.CantabularClient
// 		mockedCantabularClient := &CantabularClientMock{
// 			CheckerFunc: func(contextMoqParam context.Context, checkState *healthcheck.CheckState) error {
// 				panic("mock out the Checker method")
// 			},
// 		}
//
// 		// use mockedCantabularClient in code that requires api.CantabularClient
// 		// and then make assertions.
//
// 	}
type CantabularClientMock struct {
	// CheckerFunc mocks the Checker method.
	CheckerFunc func(contextMoqParam context.Context, checkState *healthcheck.CheckState) error

	// calls tracks calls to the methods.
	calls struct {
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// CheckState is the checkState argument value.
			CheckState *healthcheck.CheckState
		}
	}
	lockChecker sync.RWMutex
}

// Checker calls CheckerFunc.
func (mock *CantabularClientMock) Checker(contextMoqParam context.Context, checkState *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("CantabularClientMock.CheckerFunc: method is nil but CantabularClient.Checker was just called")
	}
	callInfo := struct {
		ContextMoqParam context.Context
		CheckState      *healthcheck.CheckState
	}{
		ContextMoqParam: contextMoqParam,
		CheckState:      checkState,
	}
	mock.lockChecker.Lock()
	mock.calls.Checker = append(mock.calls.Checker, callInfo)
	mock.lockChecker.Unlock()
	return mock.CheckerFunc(contextMoqParam, checkState)
}

// CheckerCalls gets all the calls that were made to Checker.
// Check the length with:
//     len(mockedCantabularClient.CheckerCalls())
func (mock *CantabularClientMock) CheckerCalls() []struct {
	ContextMoqParam context.Context
	CheckState      *healthcheck.CheckState
} {
	var calls []struct {
		ContextMoqParam context.Context
		CheckState      *healthcheck.CheckState
	}
	mock.lockChecker.RLock()
	calls = mock.calls.Checker
	mock.lockChecker.RUnlock()
	return calls
}