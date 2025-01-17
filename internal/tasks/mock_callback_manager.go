// Code generated by MockGen. DO NOT EDIT.
// Source: callback_manager.go
//
// Generated by this command:
//
//	mockgen -source=callback_manager.go -destination=../../internal/tasks/mock_callback_manager.go -package=tasks
//

// Package tasks is a generated GoMock package.
package tasks

import (
	reflect "reflect"

	model "github.com/flightctl/flightctl/internal/store/model"
	uuid "github.com/google/uuid"
	gomock "go.uber.org/mock/gomock"
)

// MockCallbackManager is a mock of CallbackManager interface.
type MockCallbackManager struct {
	ctrl     *gomock.Controller
	recorder *MockCallbackManagerMockRecorder
}

// MockCallbackManagerMockRecorder is the mock recorder for MockCallbackManager.
type MockCallbackManagerMockRecorder struct {
	mock *MockCallbackManager
}

// NewMockCallbackManager creates a new mock instance.
func NewMockCallbackManager(ctrl *gomock.Controller) *MockCallbackManager {
	mock := &MockCallbackManager{ctrl: ctrl}
	mock.recorder = &MockCallbackManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCallbackManager) EXPECT() *MockCallbackManagerMockRecorder {
	return m.recorder
}

// AllDevicesDeletedCallback mocks base method.
func (m *MockCallbackManager) AllDevicesDeletedCallback(orgId uuid.UUID) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AllDevicesDeletedCallback", orgId)
}

// AllDevicesDeletedCallback indicates an expected call of AllDevicesDeletedCallback.
func (mr *MockCallbackManagerMockRecorder) AllDevicesDeletedCallback(orgId any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllDevicesDeletedCallback", reflect.TypeOf((*MockCallbackManager)(nil).AllDevicesDeletedCallback), orgId)
}

// AllFleetsDeletedCallback mocks base method.
func (m *MockCallbackManager) AllFleetsDeletedCallback(orgId uuid.UUID) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AllFleetsDeletedCallback", orgId)
}

// AllFleetsDeletedCallback indicates an expected call of AllFleetsDeletedCallback.
func (mr *MockCallbackManagerMockRecorder) AllFleetsDeletedCallback(orgId any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllFleetsDeletedCallback", reflect.TypeOf((*MockCallbackManager)(nil).AllFleetsDeletedCallback), orgId)
}

// AllRepositoriesDeletedCallback mocks base method.
func (m *MockCallbackManager) AllRepositoriesDeletedCallback(orgId uuid.UUID) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AllRepositoriesDeletedCallback", orgId)
}

// AllRepositoriesDeletedCallback indicates an expected call of AllRepositoriesDeletedCallback.
func (mr *MockCallbackManagerMockRecorder) AllRepositoriesDeletedCallback(orgId any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllRepositoriesDeletedCallback", reflect.TypeOf((*MockCallbackManager)(nil).AllRepositoriesDeletedCallback), orgId)
}

// DeviceSourceUpdated mocks base method.
func (m *MockCallbackManager) DeviceSourceUpdated(orgId uuid.UUID, name string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DeviceSourceUpdated", orgId, name)
}

// DeviceSourceUpdated indicates an expected call of DeviceSourceUpdated.
func (mr *MockCallbackManagerMockRecorder) DeviceSourceUpdated(orgId, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeviceSourceUpdated", reflect.TypeOf((*MockCallbackManager)(nil).DeviceSourceUpdated), orgId, name)
}

// DeviceUpdatedCallback mocks base method.
func (m *MockCallbackManager) DeviceUpdatedCallback(before, after *model.Device) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DeviceUpdatedCallback", before, after)
}

// DeviceUpdatedCallback indicates an expected call of DeviceUpdatedCallback.
func (mr *MockCallbackManagerMockRecorder) DeviceUpdatedCallback(before, after any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeviceUpdatedCallback", reflect.TypeOf((*MockCallbackManager)(nil).DeviceUpdatedCallback), before, after)
}

// FleetSourceUpdated mocks base method.
func (m *MockCallbackManager) FleetSourceUpdated(orgId uuid.UUID, name string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "FleetSourceUpdated", orgId, name)
}

// FleetSourceUpdated indicates an expected call of FleetSourceUpdated.
func (mr *MockCallbackManagerMockRecorder) FleetSourceUpdated(orgId, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FleetSourceUpdated", reflect.TypeOf((*MockCallbackManager)(nil).FleetSourceUpdated), orgId, name)
}

// FleetUpdatedCallback mocks base method.
func (m *MockCallbackManager) FleetUpdatedCallback(before, after *model.Fleet) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "FleetUpdatedCallback", before, after)
}

// FleetUpdatedCallback indicates an expected call of FleetUpdatedCallback.
func (mr *MockCallbackManagerMockRecorder) FleetUpdatedCallback(before, after any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FleetUpdatedCallback", reflect.TypeOf((*MockCallbackManager)(nil).FleetUpdatedCallback), before, after)
}

// RepositoryUpdatedCallback mocks base method.
func (m *MockCallbackManager) RepositoryUpdatedCallback(repository *model.Repository) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RepositoryUpdatedCallback", repository)
}

// RepositoryUpdatedCallback indicates an expected call of RepositoryUpdatedCallback.
func (mr *MockCallbackManagerMockRecorder) RepositoryUpdatedCallback(repository any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RepositoryUpdatedCallback", reflect.TypeOf((*MockCallbackManager)(nil).RepositoryUpdatedCallback), repository)
}

// TemplateVersionCreatedCallback mocks base method.
func (m *MockCallbackManager) TemplateVersionCreatedCallback(templateVersion *model.TemplateVersion) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "TemplateVersionCreatedCallback", templateVersion)
}

// TemplateVersionCreatedCallback indicates an expected call of TemplateVersionCreatedCallback.
func (mr *MockCallbackManagerMockRecorder) TemplateVersionCreatedCallback(templateVersion any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TemplateVersionCreatedCallback", reflect.TypeOf((*MockCallbackManager)(nil).TemplateVersionCreatedCallback), templateVersion)
}
