package readiness_checker

import (
	"github.com/InVisionApp/go-health/v2"
	"time"
)

func NewHealthConfig(
	name string,
	interval time.Duration,
	checker health.ICheckable,
) *health.Config {
	return &health.Config{
		Name:       name,
		Checker:    checker,
		Interval:   interval,
		Fatal:      true,
	}
}
