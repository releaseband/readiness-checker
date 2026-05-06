package readiness_test

import (
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/releaseband/readiness-checker/v2/readiness"
)

// mockCheck implements readiness.Checkable for tests.
type mockCheck struct {
	err error
}

func (m *mockCheck) Status() (any, error) {
	return nil, m.err
}

func cfg(name string, check *mockCheck) *readiness.Config {
	return &readiness.Config{
		Name:     name,
		Checker:  check,
		Interval: 10 * time.Millisecond,
	}
}

// waitForFirstRun polls until checks have run at least once (ErrNoStates goes away).
func waitForFirstRun(t *testing.T, c *readiness.Checker) {
	t.Helper()
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if err := c.Readiness(); !errors.Is(err, readiness.ErrNoStates) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out waiting for first check run")
}

func TestNew_returnsNonNil(t *testing.T) {
	if readiness.New() == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNew_withLogger(t *testing.T) {
	c := readiness.New(
		readiness.WithLogger(
			readiness.NewSlogAdapter(
				slog.Default())),
	)
	if c == nil {
		t.Fatal("New(WithLogger) returned nil")
	}
}

func TestNew_withErrorDetails(t *testing.T) {
	c := readiness.New(readiness.WithErrorDetails())
	if c == nil {
		t.Fatal("New(WithErrorDetails) returned nil")
	}
}

func TestAddChecks_nilConfig(t *testing.T) {
	c := readiness.New()
	err := c.AddChecks(nil)
	if !errors.Is(err, readiness.ErrConfigShouldNotBeNil) {
		t.Fatalf("expected ErrConfigShouldNotBeNil, got %v", err)
	}
}

func TestAddChecks_duplicateName(t *testing.T) {
	c := readiness.New()
	mc := &mockCheck{}
	if err := c.AddChecks(cfg("db", mc)); err != nil {
		t.Fatalf("first AddChecks failed: %v", err)
	}
	err := c.AddChecks(cfg("db", mc))
	if !errors.Is(err, readiness.ErrDuplicateName) {
		t.Fatalf("expected ErrDuplicateName, got %v", err)
	}
}

func TestAddChecks_afterStart(t *testing.T) {
	c := readiness.New()
	mc := &mockCheck{}
	c.AddChecks(cfg("db", mc))
	c.Start()
	defer c.Shutdown()

	err := c.AddChecks(cfg("cache", mc))
	if !errors.Is(err, readiness.ErrAlreadyStarted) {
		t.Fatalf("expected ErrAlreadyStarted, got %v", err)
	}
}

func TestReadiness_noStates(t *testing.T) {
	c := readiness.New()
	mc := &mockCheck{}
	c.AddChecks(cfg("db", mc))
	c.Start()
	defer c.Shutdown()

	// Immediately after Start, checks have not run yet.
	err := c.Readiness()
	if !errors.Is(err, readiness.ErrNoStates) {
		t.Fatalf("expected ErrNoStates, got %v", err)
	}
}

func TestReadiness_healthy(t *testing.T) {
	c := readiness.New()
	mc := &mockCheck{err: nil}
	c.AddChecks(cfg("db", mc))
	c.Start()
	defer c.Shutdown()

	waitForFirstRun(t, c)

	if err := c.Readiness(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestReadiness_checkFails(t *testing.T) {
	c := readiness.New()
	mc := &mockCheck{err: errors.New("connection refused")}
	c.AddChecks(cfg("mongo", mc))
	c.Start()
	defer c.Shutdown()

	waitForFirstRun(t, c)

	err := c.Readiness()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Error must include the check name and the original error text.
	errStr := err.Error()
	if !strings.Contains(errStr, "mongo") {
		t.Errorf("error %q does not contain check name 'mongo'", errStr)
	}
	if !strings.Contains(errStr, "connection refused") {
		t.Errorf("error %q does not contain original error 'connection refused'", errStr)
	}
}

func TestReadiness_afterShutdown(t *testing.T) {
	c := readiness.New()
	mc := &mockCheck{}
	c.AddChecks(cfg("db", mc))
	c.Start()
	c.Shutdown()

	err := c.Readiness()
	if !errors.Is(err, readiness.ErrShutdownSignalReceived) {
		t.Fatalf("expected ErrShutdownSignalReceived, got %v", err)
	}
}

func TestShutdown_idempotent(t *testing.T) {
	c := readiness.New()
	mc := &mockCheck{}
	c.AddChecks(cfg("db", mc))
	c.Start()

	if err := c.Shutdown(); err != nil {
		t.Fatalf("first Shutdown() failed: %v", err)
	}
	// Second call must not panic or return an error.
	if err := c.Shutdown(); err != nil {
		t.Fatalf("second Shutdown() failed: %v", err)
	}
}
