// Code generated by MockGen. DO NOT EDIT.
// Source: packagemanager.go

// Package fake is a generated GoMock package.
package fake

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockPackageManager is a mock of PackageManager interface
type MockPackageManager struct {
	ctrl     *gomock.Controller
	recorder *MockPackageManagerMockRecorder
}

// MockPackageManagerMockRecorder is the mock recorder for MockPackageManager
type MockPackageManagerMockRecorder struct {
	mock *MockPackageManager
}

// NewMockPackageManager creates a new mock instance
func NewMockPackageManager(ctrl *gomock.Controller) *MockPackageManager {
	mock := &MockPackageManager{ctrl: ctrl}
	mock.recorder = &MockPackageManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPackageManager) EXPECT() *MockPackageManagerMockRecorder {
	return m.recorder
}

// String mocks base method
func (m *MockPackageManager) String() string {
	ret := m.ctrl.Call(m, "String")
	ret0, _ := ret[0].(string)
	return ret0
}

// String indicates an expected call of String
func (mr *MockPackageManagerMockRecorder) String() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "String", reflect.TypeOf((*MockPackageManager)(nil).String))
}

// Install mocks base method
func (m *MockPackageManager) Install(pkg string) error {
	ret := m.ctrl.Call(m, "Install", pkg)
	ret0, _ := ret[0].(error)
	return ret0
}

// Install indicates an expected call of Install
func (mr *MockPackageManagerMockRecorder) Install(pkg interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Install", reflect.TypeOf((*MockPackageManager)(nil).Install), pkg)
}

// Update mocks base method
func (m *MockPackageManager) Update(pkg string) error {
	ret := m.ctrl.Call(m, "Update", pkg)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update
func (mr *MockPackageManagerMockRecorder) Update(pkg interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockPackageManager)(nil).Update), pkg)
}

// Uninstall mocks base method
func (m *MockPackageManager) Uninstall(pkg string) error {
	ret := m.ctrl.Call(m, "Uninstall", pkg)
	ret0, _ := ret[0].(error)
	return ret0
}

// Uninstall indicates an expected call of Uninstall
func (mr *MockPackageManagerMockRecorder) Uninstall(pkg interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Uninstall", reflect.TypeOf((*MockPackageManager)(nil).Uninstall), pkg)
}