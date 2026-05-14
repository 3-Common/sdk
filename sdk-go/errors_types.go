package threecommon

import "time"

// AuthError is returned for 401 Unauthorized — invalid, missing, or expired
// API key.
type AuthError struct{ *APIError }

// PermissionError is returned for 403 Forbidden — the API key lacks the scope
// required by the endpoint.
type PermissionError struct{ *APIError }

// NotFoundError is returned for 404 Not Found.
type NotFoundError struct{ *APIError }

// ValidationError is returned for 400 Bad Request and 422 Unprocessable
// Entity — request validation failed.
type ValidationError struct{ *APIError }

// ConflictError is returned for 409 Conflict — the request conflicts with
// current resource state.
type ConflictError struct{ *APIError }

// RateLimitError is returned for 429 Too Many Requests. RetryAfter carries
// the parsed Retry-After header so callers can implement their own backoff;
// it is zero when the server did not provide one.
type RateLimitError struct {
	*APIError
	RetryAfter time.Duration
}

// ServerError is returned for 5xx — the API returned an unexpected
// server-side failure.
type ServerError struct{ *APIError }

// ConnectionError is returned when the request never completed: DNS failure,
// TCP reset, TLS error, context cancellation, etc. The original cause is
// available via [errors.Unwrap].
type ConnectionError struct{ *APIError }
