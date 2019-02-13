// Code generated by MockGen. DO NOT EDIT.
// Source: net/http (interfaces: Handler)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	http "net/http"
	reflect "reflect"
)

// MockHTTPHandler is a mock of Handler interface
type MockHTTPHandler struct {
	ctrl     *gomock.Controller
	recorder *MockHTTPHandlerMockRecorder
}

// MockHTTPHandlerMockRecorder is the mock recorder for MockHTTPHandler
type MockHTTPHandlerMockRecorder struct {
	mock *MockHTTPHandler
}

// NewMockHTTPHandler creates a new mock instance
func NewMockHTTPHandler(ctrl *gomock.Controller) *MockHTTPHandler {
	mock := &MockHTTPHandler{ctrl: ctrl}
	mock.recorder = &MockHTTPHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHTTPHandler) EXPECT() *MockHTTPHandlerMockRecorder {
	return m.recorder
}

// ServeHTTP mocks base method
func (m *MockHTTPHandler) ServeHTTP(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ServeHTTP", arg0, arg1)
}

// ServeHTTP indicates an expected call of ServeHTTP
func (mr *MockHTTPHandlerMockRecorder) ServeHTTP(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ServeHTTP", reflect.TypeOf((*MockHTTPHandler)(nil).ServeHTTP), arg0, arg1)
}
