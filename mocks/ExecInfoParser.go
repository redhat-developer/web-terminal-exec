// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// ExecInfoParser is an autogenerated mock type for the ExecInfoParser type
type ExecInfoParser struct {
	mock.Mock
}

// ParseShellFromEtcPassWd provides a mock function with given fields: etcPassWdContent, userId
func (_m *ExecInfoParser) ParseShellFromEtcPassWd(etcPassWdContent string, userId string) (string, error) {
	ret := _m.Called(etcPassWdContent, userId)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(etcPassWdContent, userId)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(etcPassWdContent, userId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ParseUID provides a mock function with given fields: content
func (_m *ExecInfoParser) ParseUID(content string) (string, error) {
	ret := _m.Called(content)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(content)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(content)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
