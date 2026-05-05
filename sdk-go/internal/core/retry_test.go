package core_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/3-Common/sdk/sdk-go/internal/core"
)

func TestIsIdempotent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		method string
		hasKey bool
		want   bool
	}{
		{http.MethodGet, false, true},
		{http.MethodPatch, false, true},
		{http.MethodPut, false, true},
		{http.MethodPost, false, false},
		{http.MethodDelete, false, false},
		{http.MethodPost, true, true},
		{http.MethodDelete, true, true},
	}

	for _, tc := range cases {
		t.Run(tc.method, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, core.IsIdempotent(tc.method, tc.hasKey))
		})
	}
}

func TestIsRetryableStatus(t *testing.T) {
	t.Parallel()

	for _, s := range []int{408, 425, 429, 500, 502, 503, 504} {
		assert.True(t, core.IsRetryableStatus(s), "status %d should be retryable", s)
	}
	for _, s := range []int{200, 301, 400, 401, 404, 422, 501} {
		assert.False(t, core.IsRetryableStatus(s), "status %d should not be retryable", s)
	}
}

func TestComputeBackoff_RetryAfterTakesPrecedence(t *testing.T) {
	t.Parallel()

	policy := core.RetryPolicy{
		MaxRetries: 3,
		Initial:    100 * time.Millisecond,
		Max:        2 * time.Second,
		Jitter:     false,
	}

	got := core.ComputeBackoff(0, 500*time.Millisecond, policy)
	assert.Equal(t, 500*time.Millisecond, got)
}

func TestComputeBackoff_RetryAfterCappedAtMax(t *testing.T) {
	t.Parallel()

	policy := core.RetryPolicy{Initial: 100 * time.Millisecond, Max: 1 * time.Second, Jitter: false}
	got := core.ComputeBackoff(0, 10*time.Second, policy)
	assert.Equal(t, 1*time.Second, got)
}

func TestComputeBackoff_ExponentialNoJitter(t *testing.T) {
	t.Parallel()

	policy := core.RetryPolicy{Initial: 100 * time.Millisecond, Max: 2 * time.Second, Jitter: false}
	assert.Equal(t, 100*time.Millisecond, core.ComputeBackoff(0, 0, policy))
	assert.Equal(t, 200*time.Millisecond, core.ComputeBackoff(1, 0, policy))
	assert.Equal(t, 400*time.Millisecond, core.ComputeBackoff(2, 0, policy))
	assert.Equal(t, 800*time.Millisecond, core.ComputeBackoff(3, 0, policy))
	assert.Equal(t, 1600*time.Millisecond, core.ComputeBackoff(4, 0, policy))
	// 5: 3.2s — capped at Max
	assert.Equal(t, 2*time.Second, core.ComputeBackoff(5, 0, policy))
}

func TestComputeBackoff_NegativeAttemptClamped(t *testing.T) {
	t.Parallel()

	policy := core.RetryPolicy{Initial: 100 * time.Millisecond, Max: 2 * time.Second, Jitter: false}
	assert.Equal(t, 100*time.Millisecond, core.ComputeBackoff(-1, 0, policy))
}

func TestComputeBackoff_JitterStaysWithinBounds(t *testing.T) {
	t.Parallel()

	policy := core.RetryPolicy{Initial: 100 * time.Millisecond, Max: 2 * time.Second, Jitter: true}
	for i := 0; i < 100; i++ {
		got := core.ComputeBackoff(2, 0, policy)
		assert.GreaterOrEqual(t, got, time.Duration(0))
		assert.Less(t, got, 400*time.Millisecond)
	}
}

func TestComputeBackoff_JitterZeroBackoff(t *testing.T) {
	t.Parallel()

	policy := core.RetryPolicy{Initial: 0, Max: 0, Jitter: true}
	assert.Equal(t, time.Duration(0), core.ComputeBackoff(0, 0, policy))
}
