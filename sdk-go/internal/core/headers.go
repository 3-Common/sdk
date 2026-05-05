package core

import (
	"net/http"
)

// HeadersInput captures everything BuildHeaders needs to populate a request's
// header map. Pre-resolved by the caller so this stays a pure function.
type HeadersInput struct {
	APIKey          string
	APIVersion      string
	SDKVersion      string
	UserAgentSuffix string
	TelemetryHeader string // "" omits the header
	IdempotencyKey  string // "" omits the header
}

// BuildHeaders returns a fresh [http.Header] populated with every header the
// SDK sends on every request.
func BuildHeaders(in HeadersInput) http.Header {
	h := http.Header{}
	h.Set("Authorization", "Bearer "+in.APIKey)
	h.Set("Threecommon-Version", in.APIVersion)
	h.Set("User-Agent", "ThreeCommonGo/"+in.SDKVersion+" ("+in.UserAgentSuffix+")")
	h.Set("Accept", "application/json")
	h.Set("Content-Type", "application/json")

	if in.TelemetryHeader != "" {
		h.Set("Threecommon-Client-Telemetry", in.TelemetryHeader)
	}
	if in.IdempotencyKey != "" {
		h.Set("Idempotency-Key", in.IdempotencyKey)
	}
	return h
}
