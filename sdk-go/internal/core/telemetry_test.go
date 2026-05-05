package core_test

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/3-Common/sdk/sdk-go/internal/core"
)

func TestTelemetry_DisabledReturnsEmptyHeader(t *testing.T) {
	t.Parallel()

	tel := core.NewTelemetry(false)
	assert.False(t, tel.Enabled())
	assert.Empty(t, tel.HeaderValue("0.1.0", "2026-04-29"))
}

func TestTelemetry_EnabledNoLastEmitsBaselinePayload(t *testing.T) {
	t.Parallel()

	tel := core.NewTelemetry(true)
	got := tel.HeaderValue("0.1.0", "2026-04-29")

	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(got), &payload))
	assert.Equal(t, "go", payload["lang"])
	assert.Equal(t, "0.1.0", payload["sdk"])
	assert.Equal(t, "2026-04-29", payload["api"])
	assert.NotContains(t, payload, "last")
}

func TestTelemetry_RecordPopulatesLast(t *testing.T) {
	t.Parallel()

	tel := core.NewTelemetry(true)
	tel.Record("GET", "/events", 200, 123*time.Millisecond)

	got := tel.HeaderValue("0.1.0", "2026-04-29")
	assert.Contains(t, got, `"m":"GET"`)
	assert.Contains(t, got, `"p":"/events"`)
	assert.Contains(t, got, `"s":200`)
	assert.Contains(t, got, `"d":123`)
}

func TestTelemetry_DisableClearsState(t *testing.T) {
	t.Parallel()

	tel := core.NewTelemetry(true)
	tel.Record("GET", "/events", 200, time.Second)
	tel.Disable()

	assert.False(t, tel.Enabled())
	assert.Empty(t, tel.HeaderValue("0.1.0", "2026-04-29"))
}

func TestTelemetry_RecordWhenDisabledIsNoOp(t *testing.T) {
	t.Parallel()

	tel := core.NewTelemetry(false)
	tel.Record("GET", "/events", 200, time.Second) // shouldn't panic or store

	tel2 := core.NewTelemetry(true)
	tel2.Disable()
	tel2.Record("GET", "/events", 200, time.Second)
	assert.Empty(t, tel2.HeaderValue("0.1.0", "2026-04-29"))
}

func TestTelemetry_ConcurrentSafe(t *testing.T) {
	t.Parallel()

	tel := core.NewTelemetry(true)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			tel.Record("GET", "/events", 200, time.Millisecond)
		}()
		go func() {
			defer wg.Done()
			_ = tel.HeaderValue("0.1.0", "2026-04-29")
		}()
	}
	wg.Wait()
}
