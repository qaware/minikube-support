// Code generated by MockGen. DO NOT EDIT.
// Source: packagemanager.go

// Package fake is a generated GoMock package.
package fake

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
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
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "String")
	ret0, _ := ret[0].(string)
	return ret0
}

// String indicates an expected call of String
func (mr *MockPackageManagerMockRecorder) String() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "String", reflect.TypeOf((*MockPackageManager)(nil).String))
}

// Install mocks base method
func (m *MockPackageManager) Install(pkg string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Install", pkg)
	ret0, _ := ret[0].(error)
	return ret0
}

// Install indicates an expected call of Install
func (mr *MockPackageManagerMockRecorder) Install(pkg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Install", reflect.TypeOf((*MockPackageManager)(nil).Install), pkg)
}

// IsInstalled mocks base method
func (m *MockPackageManager) IsInstalled(pkg string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsInstalled", pkg)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsInstalled indicates an expected call of IsInstalled
func (mr *MockPackageManagerMockRecorder) IsInstalled(pkg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsInstalled", reflect.TypeOf((*MockPackageManager)(nil).IsInstalled), pkg)
}

// Update mocks base method
func (m *MockPackageManager) Update(pkg string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", pkg)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update
func (mr *MockPackageManagerMockRecorder) Update(pkg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockPackageManager)(nil).Update), pkg)
}

// Uninstall mocks base method
func (m *MockPackageManager) Uninstall(pkg string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Uninstall", pkg)
	ret0, _ := ret[0].(error)
	return ret0
}

// Uninstall indicates an expected call of Uninstall
func (mr *MockPackageManagerMockRecorder) Uninstall(pkg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Uninstall", reflect.TypeOf((*MockPackageManager)(nil).Uninstall), pkg)
}

// MockosSpecific is a mock of osSpecific interface
type MockosSpecific struct {
	ctrl     *gomock.Controller
	recorder *MockosSpecificMockRecorder
}

// MockosSpecificMockRecorder is the mock recorder for MockosSpecific
type MockosSpecificMockRecorder struct {
	mock *MockosSpecific
}

// NewMockosSpecific creates a new mock instance
func NewMockosSpecific(ctrl *gomock.Controller) *MockosSpecific {
	mock := &MockosSpecific{ctrl: ctrl}
	mock.recorder = &MockosSpecificMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockosSpecific) EXPECT() *MockosSpecificMockRecorder {
	return m.recorder
}

// String mocks base method
func (m *MockosSpecific) String() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "String")
	ret0, _ := ret[0].(string)
	return ret0
}

// String indicates an expected call of String
func (mr *MockosSpecificMockRecorder) String() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "String", reflect.TypeOf((*MockosSpecific)(nil).String))
}

// Install mocks base method
func (m *MockosSpecific) Install(pkg string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Install", pkg)
	ret0, _ := ret[0].(error)
	return ret0
}

// Install indicates an expected call of Install
func (mr *MockosSpecificMockRecorder) Install(pkg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Install", reflect.TypeOf((*MockosSpecific)(nil).Install), pkg)
}

// IsInstalled mocks base method
func (m *MockosSpecific) IsInstalled(pkg string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsInstalled", pkg)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsInstalled indicates an expected call of IsInstalled
func (mr *MockosSpecificMockRecorder) IsInstalled(pkg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsInstalled", reflect.TypeOf((*MockosSpecific)(nil).IsInstalled), pkg)
}

// Update mocks base method
func (m *MockosSpecific) Update(pkg string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", pkg)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update
func (mr *MockosSpecificMockRecorder) Update(pkg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockosSpecific)(nil).Update), pkg)
}

// Uninstall mocks base method
func (m *MockosSpecific) Uninstall(pkg string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Uninstall", pkg)
	ret0, _ := ret[0].(error)
	return ret0
}

// Uninstall indicates an expected call of Uninstall
func (mr *MockosSpecificMockRecorder) Uninstall(pkg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Uninstall", reflect.TypeOf((*MockosSpecific)(nil).Uninstall), pkg)
}

// IsAvailable mocks base method
func (m *MockosSpecific) IsAvailable() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsAvailable")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsAvailable indicates an expected call of IsAvailable
func (mr *MockosSpecificMockRecorder) IsAvailable() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsAvailable", reflect.TypeOf((*MockosSpecific)(nil).IsAvailable))
}
