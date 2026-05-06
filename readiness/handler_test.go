package readiness_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/releaseband/readiness-checker/v2/readiness"
)

func do(handler http.Handler) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/readiness", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func decodeJSON(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var v map[string]any
	if err := json.NewDecoder(w.Body).Decode(&v); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	return v
}

func TestHTTPHandler_healthy(t *testing.T) {
	c := readiness.New()
	mc := &mockCheck{err: nil}
	c.AddChecks(cfg("db", mc))
	c.Start()
	defer c.Shutdown()
	waitForFirstRun(t, c)

	w := do(c.HTTPHandler())

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Errorf("expected empty body, got %q", w.Body.String())
	}
}

func TestHTTPHandler_checkFails_defaultHidesDetails(t *testing.T) {
	c := readiness.New()
	mc := &mockCheck{err: errors.New("connection refused")}
	c.AddChecks(cfg("mongo", mc))
	c.Start()
	defer c.Shutdown()
	waitForFirstRun(t, c)

	w := do(c.HTTPHandler())

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
	body := decodeJSON(t, w)
	if body["error"] != "health check failed" {
		t.Errorf("unexpected error field: %v", body["error"])
	}
	if _, ok := body["failed_checks"]; ok {
		t.Error("failed_checks must not appear in default mode")
	}
}

func TestHTTPHandler_checkFails_withErrorDetails(t *testing.T) {
	c := readiness.New(readiness.WithErrorDetails())
	mc := &mockCheck{err: errors.New("connection refused")}
	c.AddChecks(cfg("mongo", mc))
	c.Start()
	defer c.Shutdown()
	waitForFirstRun(t, c)

	w := do(c.HTTPHandler())

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
	body := decodeJSON(t, w)
	fc, ok := body["failed_checks"]
	if !ok {
		t.Fatal("expected failed_checks in response")
	}
	checks := fc.([]any)
	if len(checks) == 0 {
		t.Fatal("expected at least one entry in failed_checks")
	}
	entry := checks[0].(map[string]any)
	if entry["name"] != "mongo" {
		t.Errorf("unexpected check name: %v", entry["name"])
	}
	if entry["error"] != "connection refused" {
		t.Errorf("unexpected check error: %v", entry["error"])
	}
}

func TestHTTPHandler_noStates(t *testing.T) {
	c := readiness.New()
	mc := &mockCheck{}
	c.AddChecks(cfg("db", mc))
	// Call handler before Start() — state map is always empty before Start,
	// so Readiness() always returns ErrNoStates. This avoids a race with
	// go-health's immediate first-run goroutine.
	w := do(c.HTTPHandler())

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
	body := decodeJSON(t, w)
	if body["error"] != readiness.ErrNoStates.Error() {
		t.Errorf("unexpected error: %v", body["error"])
	}
}

func TestHTTPHandler_shutdown(t *testing.T) {
	c := readiness.New()
	mc := &mockCheck{}
	c.AddChecks(cfg("db", mc))
	c.Start()
	c.Shutdown()

	w := do(c.HTTPHandler())

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
	body := decodeJSON(t, w)
	if body["error"] != readiness.ErrShutdownSignalReceived.Error() {
		t.Errorf("unexpected error: %v", body["error"])
	}
}
