// Honestly, got ChatGPT to give me this. Looks reasonable.
package slogtest

import (
	"context"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// LogEntry represents a single captured log record.
type LogEntry struct {
	Level   slog.Level
	Message string
	Attrs   map[string]any
}

// MockHandler implements slog.Handler and stores logs in memory.
type MockHandler struct {
	mu    sync.Mutex
	logs  []LogEntry
	level slog.Level
}

// NewMockHandler creates a new MockHandler.
func NewMockHandler(level slog.Level) *MockHandler {
	return &MockHandler{level: level}
}

// Enabled implements slog.Handler.Enabled.
func (h *MockHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle implements slog.Handler.Handle.
func (h *MockHandler) Handle(_ context.Context, record slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	attrs := map[string]any{}
	record.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	h.logs = append(h.logs, LogEntry{
		Level:   record.Level,
		Message: record.Message,
		Attrs:   attrs,
	})
	return nil
}

// WithAttrs implements slog.Handler.WithAttrs.
func (h *MockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// You could propagate handler-level attrs, but tests rarely need this.
	return h
}

// WithGroup implements slog.Handler.WithGroup.
func (h *MockHandler) WithGroup(_ string) slog.Handler {
	return h
}

// Logs returns a snapshot of all log entries.
func (h *MockHandler) Logs() []LogEntry {
	h.mu.Lock()
	defer h.mu.Unlock()
	cpy := make([]LogEntry, len(h.logs))
	copy(cpy, h.logs)
	return cpy
}

func (h *MockHandler) MustHaveSeen(t *testing.T, msg string, level slog.Level, attrs map[string]any) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, l := range h.logs {
		if l.Message != msg || l.Level != level {
			continue
		}

		if attrs != nil {
			assert.Equal(t, attrs, l.Attrs, "log [%s]'%s' did not have correct attributes", level.String(), msg)
		}
		return
	}

	t.Logf("failed to find expected log [%s]'%s'", level.String(), msg)
	t.Fail()
}

// Clear removes all stored log entries (useful between tests).
func (h *MockHandler) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.logs = nil
}

func NewLogger() (*slog.Logger, *MockHandler) {
	h := NewMockHandler(slog.LevelDebug)
	return slog.New(h), h
}
