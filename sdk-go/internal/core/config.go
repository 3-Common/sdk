package core

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	threecommon "github.com/3-Common/sdk/sdk-go"
)

// EnvVarAPIKey is the environment variable consulted when [threecommon.Config.APIKey]
// is empty.
const EnvVarAPIKey = "THREECOMMON_API_KEY"

// defaults are applied when the corresponding [threecommon.Config] field is
// zero. Kept package-private; callers either pass overrides or accept these.
var defaults = struct {
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
}{
	BaseURL:    "https://api.3common.com",
	Timeout:    30 * time.Second,
	MaxRetries: 3,
}

// NewFromConfig validates a [threecommon.Config], fills in defaults, and
// returns a ready-to-use [*Client]. Resource packages and the [client]
// aggregator both call this so the entire SDK shares one validation path.
func NewFromConfig(cfg threecommon.Config) (*Client, error) {
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv(EnvVarAPIKey)
	}
	if apiKey == "" {
		return nil, &threecommon.ValidationError{APIError: &threecommon.APIError{
			Code:    "missing_api_key",
			Message: "An API key is required. Pass `APIKey` on the threecommon.Config, or set the " + EnvVarAPIKey + " environment variable.",
		}}
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaults.BaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if u, err := url.Parse(baseURL); err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, &threecommon.ValidationError{APIError: &threecommon.APIError{
			Code:    "invalid_base_url",
			Message: "BaseURL must be an absolute http:// or https:// URL; got " + cfg.BaseURL,
		}}
	}

	apiVersion := cfg.APIVersion
	if apiVersion == "" {
		apiVersion = threecommon.APIVersion
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaults.Timeout
	}
	if timeout < 0 {
		return nil, &threecommon.ValidationError{APIError: &threecommon.APIError{
			Code:    "invalid_timeout",
			Message: "Timeout must be non-negative",
		}}
	}

	maxRetries := defaults.MaxRetries
	if cfg.MaxRetries != nil {
		if *cfg.MaxRetries < 0 {
			return nil, &threecommon.ValidationError{APIError: &threecommon.APIError{
				Code:    "invalid_max_retries",
				Message: "MaxRetries must be non-negative",
			}}
		}
		maxRetries = *cfg.MaxRetries
	}

	retryDelay := cfg.RetryDelay
	if retryDelay == (threecommon.RetryDelay{}) {
		retryDelay = threecommon.DefaultRetryDelay
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	telemetryEnabled := true
	if cfg.Telemetry != nil {
		telemetryEnabled = *cfg.Telemetry
	}

	return NewClient(ClientOptions{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		APIVersion: apiVersion,
		SDKVersion: threecommon.Version,
		Timeout:    timeout,
		Retry: RetryPolicy{
			MaxRetries: maxRetries,
			Initial:    retryDelay.Initial,
			Max:        retryDelay.Max,
			Jitter:     retryDelay.Jitter,
		},
		HTTPClient: httpClient,
		Telemetry:  NewTelemetry(telemetryEnabled),
		Logger:     cfg.Logger,
	}), nil
}

// TelemetryFromClient returns the *Telemetry behind a *Client so the
// aggregator package can implement [github.com/3-Common/sdk/sdk-go/client.API.DisableTelemetry].
func TelemetryFromClient(c *Client) *Telemetry {
	return c.opts.Telemetry
}
