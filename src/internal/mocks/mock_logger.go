package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Error(err error, message string, args ...string) {
	m.Called(err, message, args)
}

func (m *MockLogger) Warning(message string, args ...string) {
	m.Called(message, args)
}

func (m *MockLogger) Information(message string, args ...string) {
	m.Called(message, args)
}

func (m *MockLogger) Debug(message string, args ...string) {
	m.Called(message, args)
}

func (m *MockLogger) Trace(message string, args ...string) {
	m.Called(message, args)
}
