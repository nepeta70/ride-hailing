package mocks

import (
	"sync"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type MockLogger struct {
	mu      sync.Mutex
	Entries []string
}

func (m *MockLogger) Debug(msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Entries = append(m.Entries, "DEBUG:"+msg)
}

func (m *MockLogger) Info(msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Entries = append(m.Entries, "INFO:"+msg)
}

func (m *MockLogger) Warn(msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Entries = append(m.Entries, "WARN:"+msg)
}

func (m *MockLogger) Error(msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Entries = append(m.Entries, "ERROR:"+msg)
}

var _ ports.Logger = (*MockLogger)(nil)
