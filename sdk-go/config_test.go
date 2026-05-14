package threecommon_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	threecommon "github.com/3-Common/sdk/sdk-go"
)

func TestVersionConstants_AreNonEmpty(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, threecommon.Version)
	assert.NotEmpty(t, threecommon.APIVersion)
	assert.Equal(t, "/v1", threecommon.APIPath)
}

func TestDefaultRetryDelay_HasReasonableValues(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 500*time.Millisecond, threecommon.DefaultRetryDelay.Initial)
	assert.Equal(t, 8*time.Second, threecommon.DefaultRetryDelay.Max)
	assert.True(t, threecommon.DefaultRetryDelay.Jitter)
}

func TestConfig_ZeroValueIsUsable(t *testing.T) {
	t.Parallel()

	var cfg threecommon.Config
	assert.Empty(t, cfg.APIKey)
	assert.Empty(t, cfg.BaseURL)
	assert.Zero(t, cfg.Timeout)
	assert.Zero(t, cfg.MaxRetries)
	assert.Nil(t, cfg.HTTPClient)
	assert.Nil(t, cfg.Logger)
	assert.Nil(t, cfg.Telemetry)
}
