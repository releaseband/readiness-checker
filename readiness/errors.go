package readiness

import "errors"

var (
	ErrConfigShouldNotBeNil   = errors.New("config should not be nil")
	ErrDuplicateName          = errors.New("duplicate check name")
	ErrAlreadyStarted         = errors.New("checker already started: AddChecks must be called before Start")
	ErrNoStates               = errors.New("no states: checks have not run yet")
	ErrShutdownSignalReceived = errors.New("shutdown signal received")
)
