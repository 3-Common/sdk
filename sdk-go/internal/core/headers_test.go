package core_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/3-Common/sdk/sdk-go/internal/core"
)

func TestBuildHeaders_RequiredFieldsPopulated(t *testing.T) {
	t.Parallel()

	h := core.BuildHeaders(core.HeadersInput{
		APIKey:          "3co_test",
		APIVersion:      "2026-04-29",
		SDKVersion:      "0.1.0",
		UserAgentSuffix: "Go/go1.22; darwin-arm64",
		HasBody:         true,
	})

	assert.Equal(t, "Bearer 3co_test", h.Get("Authorization"))
	assert.Equal(t, "2026-04-29", h.Get("Threecommon-Version"))
	assert.Equal(t, "ThreeCommonGo/0.1.0 (Go/go1.22; darwin-arm64)", h.Get("User-Agent"))
	assert.Equal(t, "application/json", h.Get("Accept"))
	assert.Equal(t, "application/json", h.Get("Content-Type"))
}

func TestBuildHeaders_ContentTypeOmittedWhenBodyless(t *testing.T) {
	t.Parallel()

	h := core.BuildHeaders(core.HeadersInput{APIKey: "k", HasBody: false})
	assert.Empty(t, h.Get("Content-Type"))
}

func TestBuildHeaders_ContentTypeSetWhenBodyPresent(t *testing.T) {
	t.Parallel()

	h := core.BuildHeaders(core.HeadersInput{APIKey: "k", HasBody: true})
	assert.Equal(t, "application/json", h.Get("Content-Type"))
}

func TestBuildHeaders_OptionalFieldsOmittedWhenEmpty(t *testing.T) {
	t.Parallel()

	h := core.BuildHeaders(core.HeadersInput{APIKey: "k"})
	assert.Empty(t, h.Get("Threecommon-Client-Telemetry"))
	assert.Empty(t, h.Get("Idempotency-Key"))
}

func TestBuildHeaders_OptionalFieldsSetWhenProvided(t *testing.T) {
	t.Parallel()

	h := core.BuildHeaders(core.HeadersInput{
		APIKey:          "k",
		TelemetryHeader: `{"lang":"go"}`,
		IdempotencyKey:  "key-1",
	})
	assert.Equal(t, `{"lang":"go"}`, h.Get("Threecommon-Client-Telemetry"))
	assert.Equal(t, "key-1", h.Get("Idempotency-Key"))
}
