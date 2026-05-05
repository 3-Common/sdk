package core

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Internal test (package httpclient, not httpclient_test) so we can exercise
// the unexported sleepWithContext directly.

func TestSleepWithContext_ZeroDurationIsNoOp(t *testing.T) {
	t.Parallel()

	start := time.Now()
	require.NoError(t, sleepWithContext(context.Background(), 0))
	assert.Less(t, time.Since(start), 5*time.Millisecond)
}

func TestSleepWithContext_NegativeDurationIsNoOp(t *testing.T) {
	t.Parallel()

	require.NoError(t, sleepWithContext(context.Background(), -time.Second))
}

func TestSleepWithContext_NormalSleep(t *testing.T) {
	t.Parallel()

	start := time.Now()
	require.NoError(t, sleepWithContext(context.Background(), 30*time.Millisecond))
	assert.GreaterOrEqual(t, time.Since(start), 20*time.Millisecond)
}

func TestSleepWithContext_RespectsCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := sleepWithContext(ctx, 5*time.Second)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestUserAgentSuffix_NotEmpty(t *testing.T) {
	t.Parallel()

	got := userAgentSuffix()
	assert.NotEmpty(t, got)
	assert.Contains(t, got, "Go/")
}

func TestNewClient_AppliesDefaults(t *testing.T) {
	t.Parallel()

	c := NewClient(ClientOptions{
		APIKey:    "k",
		Telemetry: NewTelemetry(false),
	})
	assert.NotNil(t, c)
}

func TestDefaultCodeForStatus(t *testing.T) {
	t.Parallel()

	cases := map[int]string{
		401: "unauthorized",
		403: "forbidden",
		404: "not_found",
		409: "conflict",
		429: "rate_limit_exceeded",
		500: "internal_error",
		502: "internal_error",
		418: "request_failed",
	}
	for status, want := range cases {
		assert.Equal(t, want, defaultCodeForStatus(status), "status %d", status)
	}
}
