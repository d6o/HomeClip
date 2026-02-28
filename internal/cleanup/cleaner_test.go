package cleanup

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type mockCleanable struct {
	mu       sync.Mutex
	calls    int
	maxAge   time.Duration
	returnErr error
}

func (m *mockCleanable) Cleanup(_ context.Context, maxAge time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	m.maxAge = maxAge
	return m.returnErr
}

func (m *mockCleanable) getCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

func TestNewCleaner(t *testing.T) {
	m1 := &mockCleanable{}
	m2 := &mockCleanable{}

	c := NewCleaner(5*time.Minute, 24*time.Hour, m1, m2)

	if c.interval != 5*time.Minute {
		t.Errorf("expected interval 5m, got %v", c.interval)
	}
	if c.maxAge != 24*time.Hour {
		t.Errorf("expected maxAge 24h, got %v", c.maxAge)
	}
	if len(c.targets) != 2 {
		t.Errorf("expected 2 targets, got %d", len(c.targets))
	}
}

func TestCleaner_RunCycle(t *testing.T) {
	m1 := &mockCleanable{}
	m2 := &mockCleanable{}

	c := NewCleaner(time.Minute, 2*time.Hour, m1, m2)
	c.runCycle(context.Background())

	if m1.getCalls() != 1 {
		t.Errorf("expected 1 call to m1, got %d", m1.getCalls())
	}
	if m2.getCalls() != 1 {
		t.Errorf("expected 1 call to m2, got %d", m2.getCalls())
	}
	if m1.maxAge != 2*time.Hour {
		t.Errorf("expected maxAge 2h, got %v", m1.maxAge)
	}
}

func TestCleaner_RunCycleWithError(t *testing.T) {
	m := &mockCleanable{returnErr: errors.New("cleanup error")}

	c := NewCleaner(time.Minute, time.Hour, m)
	c.runCycle(context.Background())

	if m.getCalls() != 1 {
		t.Errorf("expected 1 call, got %d", m.getCalls())
	}
}

func TestCleaner_RunContextCancel(t *testing.T) {
	m := &mockCleanable{}
	c := NewCleaner(time.Hour, time.Hour, m)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := c.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestCleaner_RunTicksAndCleans(t *testing.T) {
	m := &mockCleanable{}
	c := NewCleaner(10*time.Millisecond, time.Hour, m)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = c.Run(ctx)

	if m.getCalls() < 1 {
		t.Errorf("expected at least 1 cleanup call, got %d", m.getCalls())
	}
}
