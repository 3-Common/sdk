package core

import (
	"encoding/json"
	"sync"
	"time"
)

// Telemetry tracks one previous-request snapshot per client and emits the
// Threecommon-Client-Telemetry header value for the next request. Goroutine-
// safe — the snapshot is updated under a mutex because every field is read
// together.
type Telemetry struct {
	mu      sync.Mutex
	enabled bool
	last    *telemetrySnapshot
}

type telemetrySnapshot struct {
	Method   string
	Path     string
	Status   int
	Duration time.Duration
}

// NewTelemetry returns a [*Telemetry] in the given enabled state.
func NewTelemetry(enabled bool) *Telemetry {
	return &Telemetry{enabled: enabled}
}

// Disable turns telemetry off and clears the cached snapshot.
func (t *Telemetry) Disable() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.enabled = false
	t.last = nil
}

// Enabled reports whether the next [Telemetry.HeaderValue] call will emit a header.
func (t *Telemetry) Enabled() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.enabled
}

// Record stores a snapshot of the just-completed request. No-op when disabled.
func (t *Telemetry) Record(method, path string, status int, duration time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.enabled {
		return
	}
	t.last = &telemetrySnapshot{Method: method, Path: path, Status: status, Duration: duration}
}

// HeaderValue returns the JSON value for the Threecommon-Client-Telemetry
// header on the next request. The empty string means "do not send the header".
func (t *Telemetry) HeaderValue(sdkVersion, apiVersion string) string {
	t.mu.Lock()
	last := t.last
	enabled := t.enabled
	t.mu.Unlock()

	if !enabled {
		return ""
	}

	type lastEntry struct {
		M string `json:"m"`
		P string `json:"p"`
		S int    `json:"s"`
		D int64  `json:"d"`
	}
	type payload struct {
		Lang string     `json:"lang"`
		SDK  string     `json:"sdk"`
		API  string     `json:"api"`
		Last *lastEntry `json:"last,omitempty"`
	}

	pl := payload{Lang: "go", SDK: sdkVersion, API: apiVersion}
	if last != nil {
		pl.Last = &lastEntry{
			M: last.Method,
			P: last.Path,
			S: last.Status,
			D: last.Duration.Milliseconds(),
		}
	}
	out, err := json.Marshal(pl)
	if err != nil {
		return ""
	}
	return string(out)
}
