// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	model "github.com/eclipse/che-machine-exec/api/model"
	websocket "github.com/gorilla/websocket"
	mock "github.com/stretchr/testify/mock"
)

// ExecManager is an autogenerated mock type for the ExecManager type
type ExecManager struct {
	mock.Mock
}

// Attach provides a mock function with given fields: id, conn
func (_m *ExecManager) Attach(id int, conn *websocket.Conn) error {
	ret := _m.Called(id, conn)

	var r0 error
	if rf, ok := ret.Get(0).(func(int, *websocket.Conn) error); ok {
		r0 = rf(id, conn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Check provides a mock function with given fields: id
func (_m *ExecManager) Check(id int) (int, error) {
	ret := _m.Called(id)

	var r0 int
	if rf, ok := ret.Get(0).(func(int) int); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Create provides a mock function with given fields: machineExec
func (_m *ExecManager) Create(machineExec *model.MachineExec) (int, error) {
	ret := _m.Called(machineExec)

	var r0 int
	if rf, ok := ret.Get(0).(func(*model.MachineExec) int); ok {
		r0 = rf(machineExec)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.MachineExec) error); ok {
		r1 = rf(machineExec)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateKubeConfig provides a mock function with given fields: kubeConfigParams, containerInfo
func (_m *ExecManager) CreateKubeConfig(kubeConfigParams *model.KubeConfigParams, containerInfo *model.ContainerInfo) error {
	ret := _m.Called(kubeConfigParams, containerInfo)

	var r0 error
	if rf, ok := ret.Get(0).(func(*model.KubeConfigParams, *model.ContainerInfo) error); ok {
		r0 = rf(kubeConfigParams, containerInfo)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Remove provides a mock function with given fields: execId
func (_m *ExecManager) Remove(execId int) {
	_m.Called(execId)
}

// Resize provides a mock function with given fields: id, cols, rows
func (_m *ExecManager) Resize(id int, cols uint, rows uint) error {
	ret := _m.Called(id, cols, rows)

	var r0 error
	if rf, ok := ret.Get(0).(func(int, uint, uint) error); ok {
		r0 = rf(id, cols, rows)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Resolve provides a mock function with given fields: container, token
func (_m *ExecManager) Resolve(container string, token string) (*model.ResolvedExec, error) {
	ret := _m.Called(container, token)

	var r0 *model.ResolvedExec
	if rf, ok := ret.Get(0).(func(string, string) *model.ResolvedExec); ok {
		r0 = rf(container, token)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ResolvedExec)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(container, token)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
