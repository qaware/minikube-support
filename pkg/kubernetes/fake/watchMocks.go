// Code generated by MockGen. DO NOT EDIT.
// Source: watch.go

// Package fake is a generated GoMock package.
package fake

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
)

// MockWatchHandler is a mock of WatchHandler interface.
type MockWatchHandler struct {
	ctrl     *gomock.Controller
	recorder *MockWatchHandlerMockRecorder
}

// MockWatchHandlerMockRecorder is the mock recorder for MockWatchHandler.
type MockWatchHandlerMockRecorder struct {
	mock *MockWatchHandler
}

// NewMockWatchHandler creates a new mock instance.
func NewMockWatchHandler(ctrl *gomock.Controller) *MockWatchHandler {
	mock := &MockWatchHandler{ctrl: ctrl}
	mock.recorder = &MockWatchHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWatchHandler) EXPECT() *MockWatchHandlerMockRecorder {
	return m.recorder
}

// AddedEvent mocks base method.
func (m *MockWatchHandler) AddedEvent(obj runtime.Object) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddedEvent", obj)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddedEvent indicates an expected call of AddedEvent.
func (mr *MockWatchHandlerMockRecorder) AddedEvent(obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddedEvent", reflect.TypeOf((*MockWatchHandler)(nil).AddedEvent), obj)
}

// DeletedEvent mocks base method.
func (m *MockWatchHandler) DeletedEvent(obj runtime.Object) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletedEvent", obj)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeletedEvent indicates an expected call of DeletedEvent.
func (mr *MockWatchHandlerMockRecorder) DeletedEvent(obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletedEvent", reflect.TypeOf((*MockWatchHandler)(nil).DeletedEvent), obj)
}

// PostEvent mocks base method.
func (m *MockWatchHandler) PostEvent() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PostEvent")
	ret0, _ := ret[0].(error)
	return ret0
}

// PostEvent indicates an expected call of PostEvent.
func (mr *MockWatchHandlerMockRecorder) PostEvent() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostEvent", reflect.TypeOf((*MockWatchHandler)(nil).PostEvent))
}

// PreWatch mocks base method.
func (m *MockWatchHandler) PreWatch(options v1.ListOptions) (watch.Interface, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PreWatch", options)
	ret0, _ := ret[0].(watch.Interface)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PreWatch indicates an expected call of PreWatch.
func (mr *MockWatchHandlerMockRecorder) PreWatch(options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PreWatch", reflect.TypeOf((*MockWatchHandler)(nil).PreWatch), options)
}

// UpdatedEvent mocks base method.
func (m *MockWatchHandler) UpdatedEvent(obj runtime.Object) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatedEvent", obj)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdatedEvent indicates an expected call of UpdatedEvent.
func (mr *MockWatchHandlerMockRecorder) UpdatedEvent(obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatedEvent", reflect.TypeOf((*MockWatchHandler)(nil).UpdatedEvent), obj)
}
