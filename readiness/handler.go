package readiness

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type errorResponse struct {
	Error        string       `json:"error"`
	FailedChecks []checkError `json:"failed_checks,omitempty"`
}

type checkError struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

// HTTPHandler returns an http.Handler for the /readiness endpoint.
// Responds 200 when healthy, 503 with JSON body when not ready.
func (c *Checker) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := c.Readiness()
		if err == nil {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)

		encodeErr := json.NewEncoder(w).Encode(c.buildErrorResponse(err))
		if encodeErr != nil && c.logger != nil {
			c.logger.Error("failed to encode",
				slog.Any("error", encodeErr))
		}
	})
}

func (c *Checker) buildErrorResponse(readinessErr error) errorResponse {
	if errors.Is(readinessErr, ErrShutdownSignalReceived) {
		return errorResponse{Error: ErrShutdownSignalReceived.Error()}
	}
	if errors.Is(readinessErr, ErrNoStates) {
		return errorResponse{Error: ErrNoStates.Error()}
	}

	resp := errorResponse{Error: "health check failed"}
	if !c.errorDetails {
		return resp
	}

	// Second State() call is acceptable: readiness probes are low-frequency
	// and we need fresh state to build the detailed response.
	states, _, err := c.h.State()
	if err != nil {
		resp.FailedChecks = append(resp.FailedChecks, checkError{
			Name:  "state",
			Error: err.Error(),
		})

		return resp
	}

	for _, s := range states {
		if s.Err != "" {
			resp.FailedChecks = append(resp.FailedChecks, checkError{
				Name:  s.Name,
				Error: s.Err,
			})
		}
	}
	return resp
}
