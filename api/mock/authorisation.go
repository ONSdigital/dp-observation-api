// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"github.com/ONSdigital/dp-authorisation/auth"
	"github.com/ONSdigital/dp-observation-api/api"
	"net/http"
	"sync"
)

var (
	lockIAuthHandlerMockRequire sync.RWMutex
)

// Ensure, that IAuthHandlerMock does implement api.IAuthHandler.
// If this is not the case, regenerate this file with moq.
var _ api.IAuthHandler = &IAuthHandlerMock{}

// IAuthHandlerMock is a mock implementation of api.IAuthHandler.
//
//     func TestSomethingThatUsesIAuthHandler(t *testing.T) {
//
//         // make and configure a mocked api.IAuthHandler
//         mockedIAuthHandler := &IAuthHandlerMock{
//             RequireFunc: func(required auth.Permissions, handler http.HandlerFunc) http.HandlerFunc {
// 	               panic("mock out the Require method")
//             },
//         }
//
//         // use mockedIAuthHandler in code that requires api.IAuthHandler
//         // and then make assertions.
//
//     }
type IAuthHandlerMock struct {
	// RequireFunc mocks the Require method.
	RequireFunc func(required auth.Permissions, handler http.HandlerFunc) http.HandlerFunc

	// calls tracks calls to the methods.
	calls struct {
		// Require holds details about calls to the Require method.
		Require []struct {
			// Required is the required argument value.
			Required auth.Permissions
			// Handler is the handler argument value.
			Handler http.HandlerFunc
		}
	}
}

// Require calls RequireFunc.
func (mock *IAuthHandlerMock) Require(required auth.Permissions, handler http.HandlerFunc) http.HandlerFunc {
	if mock.RequireFunc == nil {
		panic("IAuthHandlerMock.RequireFunc: method is nil but IAuthHandler.Require was just called")
	}
	callInfo := struct {
		Required auth.Permissions
		Handler  http.HandlerFunc
	}{
		Required: required,
		Handler:  handler,
	}
	lockIAuthHandlerMockRequire.Lock()
	mock.calls.Require = append(mock.calls.Require, callInfo)
	lockIAuthHandlerMockRequire.Unlock()
	return mock.RequireFunc(required, handler)
}

// RequireCalls gets all the calls that were made to Require.
// Check the length with:
//     len(mockedIAuthHandler.RequireCalls())
func (mock *IAuthHandlerMock) RequireCalls() []struct {
	Required auth.Permissions
	Handler  http.HandlerFunc
} {
	var calls []struct {
		Required auth.Permissions
		Handler  http.HandlerFunc
	}
	lockIAuthHandlerMockRequire.RLock()
	calls = mock.calls.Require
	lockIAuthHandlerMockRequire.RUnlock()
	return calls
}
