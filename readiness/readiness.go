package readiness

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/InVisionApp/go-health/v2"
)

// Checkable is the interface check implementations must satisfy.
// Structurally identical to go-health's ICheckable — existing go-health
// built-in checkers (MongoDB, Redis, HTTP, etc.) satisfy this without adapters.
type Checkable interface {
	Status() (any, error)
}

// Config holds the configuration for a single health check.
type Config struct {
	Name     string
	Checker  Checkable
	Interval time.Duration
	Fatal    bool
}

// Option configures a Checker.
type Option func(*Checker)

// WithLogger enables structured logging via slog.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Checker) {
		if logger == nil {
			return
		}

		c.logger = logger
		c.h.Logger = &slogAdapter{logger}
	}
}

// WithErrorDetails enables per-check error messages in HTTP responses.
// By default the HTTP handler returns only a generic error to avoid
// leaking internal details (connection strings, IPs, etc.).
func WithErrorDetails() Option {
	return func(c *Checker) {
		c.errorDetails = true
	}
}

// Checker runs health checks and reports readiness.
type Checker struct {
	h      *health.Health
	logger *slog.Logger

	names        map[string]struct{}
	shutdown     atomic.Bool
	started      atomic.Bool
	stopOnce     sync.Once
	stopErr      error
	errorDetails bool
}

// New creates a Checker. Logging is disabled by default.
func New(opts ...Option) *Checker {
	c := &Checker{h: health.New(), names: make(map[string]struct{})}
	c.h.DisableLogging()
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// AddChecks registers health checks. Must be called before Start.
func (c *Checker) AddChecks(configs ...*Config) error {
	if c.started.Load() {
		return ErrAlreadyStarted
	}
	for _, cfg := range configs {
		if cfg == nil {
			return ErrConfigShouldNotBeNil
		}
		if _, exists := c.names[cfg.Name]; exists {
			return fmt.Errorf("%w: %s", ErrDuplicateName, cfg.Name)
		}
		if err := c.h.AddCheck(&health.Config{
			Name:     cfg.Name,
			Checker:  cfg.Checker,
			Interval: cfg.Interval,
			Fatal:    cfg.Fatal,
		}); err != nil {
			return fmt.Errorf("add check %q: %w", cfg.Name, err)
		}
		c.names[cfg.Name] = struct{}{}
	}
	return nil
}

// Start launches background check goroutines.
func (c *Checker) Start() error {
	if !c.started.CompareAndSwap(false, true) {
		return ErrAlreadyStarted
	}

	if err := c.h.Start(); err != nil {
		c.started.Store(false)
		return fmt.Errorf("start health checker: %w", err)
	}

	return nil
}

// Readiness returns nil when the service is ready, or a descriptive error.
func (c *Checker) Readiness() error {
	if c.shutdown.Load() {
		return ErrShutdownSignalReceived
	}

	states, _, err := c.h.State()
	if err != nil {
		return fmt.Errorf("unable to fetch states: %w", err)
	}

	if len(states) == 0 {
		return ErrNoStates
	}

	for _, s := range states {
		if s.Err != "" {
			return fmt.Errorf("check %q failed: %s", s.Name, s.Err)
		}
	}

	return nil
}

// Shutdown atomically marks the service as shutting down and stops
// background goroutines. Idempotent — safe to call multiple times.
func (c *Checker) Shutdown() error {
	c.shutdown.Store(true)

	c.stopOnce.Do(func() {
		c.stopErr = c.h.Stop()
	})

	return c.stopErr
}
