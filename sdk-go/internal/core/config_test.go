package core_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/internal/core"
)

func TestNewFromConfig_RequiresAPIKey(t *testing.T) {
	t.Setenv(core.EnvVarAPIKey, "")

	_, err := core.NewFromConfig(threecommon.Config{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_api_key", v.Code)
}

func TestNewFromConfig_AcceptsAPIKeyFromEnv(t *testing.T) {
	t.Setenv(core.EnvVarAPIKey, "3co_from_env")

	c, err := core.NewFromConfig(threecommon.Config{})
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestNewFromConfig_RejectsInvalidBaseURL(t *testing.T) {
	t.Parallel()

	_, err := core.NewFromConfig(threecommon.Config{
		APIKey:  "k",
		BaseURL: "not-a-url",
	})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "invalid_base_url", v.Code)
}

func TestNewFromConfig_RejectsNegativeTimeout(t *testing.T) {
	t.Parallel()

	_, err := core.NewFromConfig(threecommon.Config{
		APIKey:  "k",
		Timeout: -1 * time.Second,
	})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "invalid_timeout", v.Code)
}

func TestNewFromConfig_RejectsNegativeMaxRetries(t *testing.T) {
	t.Parallel()

	_, err := core.NewFromConfig(threecommon.Config{
		APIKey:     "k",
		MaxRetries: threecommon.Int(-3),
	})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "invalid_max_retries", v.Code)
}

func TestNewFromConfig_ExplicitZeroDisablesRetries(t *testing.T) {
	t.Parallel()

	c, err := core.NewFromConfig(threecommon.Config{
		APIKey:     "k",
		MaxRetries: threecommon.Int(0),
	})
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestNewFromConfig_AppliesDefaults(t *testing.T) {
	t.Parallel()

	c, err := core.NewFromConfig(threecommon.Config{APIKey: "k"})
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.True(t, core.TelemetryFromClient(c).Enabled())
}

func TestNewFromConfig_HonorsTelemetryOptOut(t *testing.T) {
	t.Parallel()

	off := false
	c, err := core.NewFromConfig(threecommon.Config{APIKey: "k", Telemetry: &off})
	require.NoError(t, err)
	assert.False(t, core.TelemetryFromClient(c).Enabled())
}

func TestNewFromConfig_HonorsCustomHTTPClient(t *testing.T) {
	t.Parallel()

	cl := &http.Client{Timeout: 1 * time.Second}
	c, err := core.NewFromConfig(threecommon.Config{APIKey: "k", HTTPClient: cl})
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestNewFromConfig_TrimsTrailingSlashOnBaseURL(t *testing.T) {
	t.Parallel()

	c, err := core.NewFromConfig(threecommon.Config{
		APIKey:  "k",
		BaseURL: "https://api.3common.com//",
	})
	require.NoError(t, err)
	require.NotNil(t, c)
}
