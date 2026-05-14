package threecommon

import "fmt"

// APIError is the base type carried by every error the SDK returns from a
// request. The HTTP-status-specific subtypes ([NotFoundError], [AuthError],
// etc.) embed [*APIError] so its fields are accessible directly via field
// promotion. Branch on the subtype with [errors.As]:
//
//	var notFound *threecommon.NotFoundError
//	if errors.As(err, &notFound) {
//		log.Println("missing:", notFound.RequestID)
//	}
//
// Fields are populated best-effort: HTTPStatus is zero for connection errors,
// RequestID is empty when the server didn't return one, RawResponse is empty
// for non-text responses.
type APIError struct {
	// Code is a stable string matching the API's error.code field, e.g.
	// "not_found" or "rate_limit_exceeded". For SDK-originated errors it
	// describes the local condition (e.g. "missing_api_key").
	Code string

	// Message is human-readable. Safe to surface to end users.
	Message string

	// HTTPStatus is the response status, or 0 if the error originated before any response was received.
	HTTPStatus int

	// RequestID is the value of the X-Request-ID response header, when present. Useful for support correlation.
	RequestID string

	// Details is the parsed API error.details payload, when present.
	Details map[string]any

	// RawResponse is the raw response body, retained for debugging.
	RawResponse string

	// Cause is the underlying error for transport-level failures. Read it via [errors.Unwrap] or directly.
	Cause error
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("[%s] %s (request_id=%s)", e.Code, e.Message, e.RequestID)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap exposes [APIError.Cause] for [errors.Is], [errors.As], and the %w
// verb.
func (e *APIError) Unwrap() error { return e.Cause }
