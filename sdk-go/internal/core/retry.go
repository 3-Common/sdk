package core

import (
	"math/rand/v2"
	"net/http"
	"time"
)

// RetryPolicy mirrors threecommon.RetryDelay plus a max-attempts cap.
type RetryPolicy struct {
	MaxRetries int
	Initial    time.Duration
	Max        time.Duration
	Jitter     bool
}

// retryableStatuses are the HTTP statuses we re-issue idempotent requests on.
var retryableStatuses = map[int]bool{
	http.StatusRequestTimeout:      true, // 408
	425:                            true, // 425 Too Early
	http.StatusTooManyRequests:     true, // 429
	http.StatusInternalServerError: true, // 500
	http.StatusBadGateway:          true, // 502
	http.StatusServiceUnavailable:  true, // 503
	http.StatusGatewayTimeout:      true, // 504
}

// idempotentMethods are HTTP methods the SDK retries automatically. POST and
// DELETE only retry when the caller passes an Idempotency-Key.
var idempotentMethods = map[string]bool{
	http.MethodGet:   true,
	http.MethodPatch: true,
	http.MethodPut:   true,
}

// IsIdempotent reports whether the SDK may safely retry a request with the
// given method. Caller-supplied idempotency keys upgrade non-idempotent
// methods.
func IsIdempotent(method string, hasIdempotencyKey bool) bool {
	if hasIdempotencyKey {
		return true
	}
	return idempotentMethods[method]
}

// IsRetryableStatus reports whether status is one we should retry on
// alongside method idempotency.
func IsRetryableStatus(status int) bool {
	return retryableStatuses[status]
}

// ComputeBackoff returns the next sleep duration. When retryAfter is non-zero
// (e.g. parsed from a Retry-After header) it takes precedence, capped at
// policy.Max. Otherwise: exponential 2^attempt * Initial, capped at Max,
// with optional full-jitter randomization.
func ComputeBackoff(attempt int, retryAfter time.Duration, policy RetryPolicy) time.Duration {
	if retryAfter > 0 {
		if retryAfter > policy.Max {
			return policy.Max
		}
		return retryAfter
	}
	if attempt < 0 {
		attempt = 0
	}
	exp := policy.Initial * (1 << uint(attempt))  //nolint:gosec // attempt is bounded by MaxRetries
	if exp > policy.Max || exp < policy.Initial { // overflow guard
		exp = policy.Max
	}
	if !policy.Jitter {
		return exp
	}
	if exp <= 0 {
		return 0
	}
	return time.Duration(rand.Int64N(int64(exp))) //nolint:gosec // jitter only
}
