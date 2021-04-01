package readiness_checker

import (
	"errors"
	"fmt"
	"github.com/InVisionApp/go-health/v2"
	log "github.com/InVisionApp/go-logger"
)

var (
	ErrConfigShouldNotBeNil   = errors.New("health config should not be nil")
	ErrNoStates               = errors.New("no states")
	ErrGeneralHealthIsFailed  = errors.New("general health is failed")
	ErrObjectStateCheckError  = errors.New("object state check error")
	ErrShutdownSignalReceived = errors.New("shutdown signal received")
)

type Checker struct {
	h        *health.Health
	shutdown bool
}

func NewReadinessChecker() *Checker {
	return &Checker{h: health.New()}
}

func NewReadinessWithLogger(logger log.Logger) *Checker {
	h := health.New()
	h.Logger = logger

	return &Checker{
		h: h,
	}
}

func (c *Checker) DisableLogging() {
	c.h.DisableLogging()
}

func (c *Checker) Start() error {
	return c.h.Start()
}

func (c *Checker) AddCheckers(configs ...*health.Config) error {
	for _, cfg := range configs {
		if cfg == nil {
			return ErrConfigShouldNotBeNil
		}

		if err := c.h.AddCheck(cfg); err != nil {
			return fmt.Errorf("add check failed for %s: %w", cfg.Name, err)
		}
	}

	return nil
}

func (c *Checker) Readiness() error {
	if c.shutdown {
		return ErrShutdownSignalReceived
	}

	states, failed, err := c.h.State()
	if err != nil {
		return fmt.Errorf("unable to fetch states: %w", err)
	}

	if len(states) == 0 {
		return fmt.Errorf("there may be an initial delay: %w", ErrNoStates)
	}

	if failed {
		return fmt.Errorf("there may be an initial delay: %w", ErrGeneralHealthIsFailed)
	}

	for _, s := range states {
		if s.Err != "" {
			return fmt.Errorf("object: %s: %w", s.Name, ErrObjectStateCheckError)
		}
	}

	return nil
}

func (c *Checker) Shutdown() {
	c.shutdown = true
}
