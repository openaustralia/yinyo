// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import commands "github.com/openaustralia/yinyo/pkg/commands"
import io "io"
import mock "github.com/stretchr/testify/mock"
import protocol "github.com/openaustralia/yinyo/pkg/protocol"

// App is an autogenerated mock type for the App type
type App struct {
	mock.Mock
}

// CreateEvent provides a mock function with given fields: runID, event
func (_m *App) CreateEvent(runID string, event protocol.Event) error {
	ret := _m.Called(runID, event)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, protocol.Event) error); ok {
		r0 = rf(runID, event)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateRun provides a mock function with given fields: params
func (_m *App) CreateRun(params map[string]string) (protocol.Run, error) {
	ret := _m.Called(params)

	var r0 protocol.Run
	if rf, ok := ret.Get(0).(func(map[string]string) protocol.Run); ok {
		r0 = rf(params)
	} else {
		r0 = ret.Get(0).(protocol.Run)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(map[string]string) error); ok {
		r1 = rf(params)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteRun provides a mock function with given fields: runID
func (_m *App) DeleteRun(runID string) error {
	ret := _m.Called(runID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(runID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetApp provides a mock function with given fields: runID
func (_m *App) GetApp(runID string) (io.Reader, error) {
	ret := _m.Called(runID)

	var r0 io.Reader
	if rf, ok := ret.Get(0).(func(string) io.Reader); ok {
		r0 = rf(runID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Reader)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(runID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCache provides a mock function with given fields: runID
func (_m *App) GetCache(runID string) (io.Reader, error) {
	ret := _m.Called(runID)

	var r0 io.Reader
	if rf, ok := ret.Get(0).(func(string) io.Reader); ok {
		r0 = rf(runID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Reader)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(runID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetEvents provides a mock function with given fields: runID, lastID
func (_m *App) GetEvents(runID string, lastID string) commands.EventIterator {
	ret := _m.Called(runID, lastID)

	var r0 commands.EventIterator
	if rf, ok := ret.Get(0).(func(string, string) commands.EventIterator); ok {
		r0 = rf(runID, lastID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(commands.EventIterator)
		}
	}

	return r0
}

// GetExitData provides a mock function with given fields: runID
func (_m *App) GetExitData(runID string) (protocol.ExitData, error) {
	ret := _m.Called(runID)

	var r0 protocol.ExitData
	if rf, ok := ret.Get(0).(func(string) protocol.ExitData); ok {
		r0 = rf(runID)
	} else {
		r0 = ret.Get(0).(protocol.ExitData)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(runID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOutput provides a mock function with given fields: runID
func (_m *App) GetOutput(runID string) (io.Reader, error) {
	ret := _m.Called(runID)

	var r0 io.Reader
	if rf, ok := ret.Get(0).(func(string) io.Reader); ok {
		r0 = rf(runID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Reader)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(runID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsRunCreated provides a mock function with given fields: runID
func (_m *App) IsRunCreated(runID string) (bool, error) {
	ret := _m.Called(runID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(runID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(runID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PutApp provides a mock function with given fields: runID, reader, objectSize
func (_m *App) PutApp(runID string, reader io.Reader, objectSize int64) error {
	ret := _m.Called(runID, reader, objectSize)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, io.Reader, int64) error); ok {
		r0 = rf(runID, reader, objectSize)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PutCache provides a mock function with given fields: runID, reader, objectSize
func (_m *App) PutCache(runID string, reader io.Reader, objectSize int64) error {
	ret := _m.Called(runID, reader, objectSize)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, io.Reader, int64) error); ok {
		r0 = rf(runID, reader, objectSize)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PutOutput provides a mock function with given fields: runID, reader, objectSize
func (_m *App) PutOutput(runID string, reader io.Reader, objectSize int64) error {
	ret := _m.Called(runID, reader, objectSize)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, io.Reader, int64) error); ok {
		r0 = rf(runID, reader, objectSize)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RecordTraffic provides a mock function with given fields: runID, external, in, out
func (_m *App) RecordTraffic(runID string, external bool, in int64, out int64) error {
	ret := _m.Called(runID, external, in, out)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, bool, int64, int64) error); ok {
		r0 = rf(runID, external, in, out)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StartRun provides a mock function with given fields: runID, output, env, callbackURL, maxRunTime
func (_m *App) StartRun(runID string, output string, env map[string]string, callbackURL string, maxRunTime int64) error {
	ret := _m.Called(runID, output, env, callbackURL, maxRunTime)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, map[string]string, string, int64) error); ok {
		r0 = rf(runID, output, env, callbackURL, maxRunTime)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
