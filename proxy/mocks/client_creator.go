// Code generated by mockery v2.50.0. DO NOT EDIT.

package mocks

import (
	abcicli "github.com/cometbft/cometbft/v2/abci/client"
	mock "github.com/stretchr/testify/mock"
)

// ClientCreator is an autogenerated mock type for the ClientCreator type
type ClientCreator struct {
	mock.Mock
}

// NewABCIConsensusClient provides a mock function with no fields
func (_m *ClientCreator) NewABCIConsensusClient() (abcicli.Client, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for NewABCIConsensusClient")
	}

	var r0 abcicli.Client
	var r1 error
	if rf, ok := ret.Get(0).(func() (abcicli.Client, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() abcicli.Client); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(abcicli.Client)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewABCIMempoolClient provides a mock function with no fields
func (_m *ClientCreator) NewABCIMempoolClient() (abcicli.Client, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for NewABCIMempoolClient")
	}

	var r0 abcicli.Client
	var r1 error
	if rf, ok := ret.Get(0).(func() (abcicli.Client, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() abcicli.Client); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(abcicli.Client)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewABCIQueryClient provides a mock function with no fields
func (_m *ClientCreator) NewABCIQueryClient() (abcicli.Client, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for NewABCIQueryClient")
	}

	var r0 abcicli.Client
	var r1 error
	if rf, ok := ret.Get(0).(func() (abcicli.Client, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() abcicli.Client); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(abcicli.Client)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewABCISnapshotClient provides a mock function with no fields
func (_m *ClientCreator) NewABCISnapshotClient() (abcicli.Client, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for NewABCISnapshotClient")
	}

	var r0 abcicli.Client
	var r1 error
	if rf, ok := ret.Get(0).(func() (abcicli.Client, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() abcicli.Client); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(abcicli.Client)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewClientCreator creates a new instance of ClientCreator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewClientCreator(t interface {
	mock.TestingT
	Cleanup(func())
}) *ClientCreator {
	mock := &ClientCreator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
