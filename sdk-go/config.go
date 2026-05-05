package threecommon

import (
	"net/http"
	"time"
)

// Config controls how a [github.com/3-Common/sdk/sdk-go/client.API] talks to
// the 3Common API. Every field is optional except [Config.APIKey]. Zero values
// fall back to sensible defaults.
type Config struct {
	// APIKey authenticates every request. Required. Generate keys in the
	// 3Common organizer dashboard (Settings -> API Keys). May also be supplied
	// via the THREECOMMON_API_KEY environment variable.
	APIKey string

	// BaseURL is the API root. Defaults to https://api.3common.com.
	BaseURL string

	// APIVersion is sent as the Threecommon-Version header. Defaults to
	// [APIVersion]. Override only if you need to opt into a newer server
	// behavior than this SDK was built against.
	APIVersion string

	// Timeout is the per-request deadline applied via context. A zero value
	// means 30s. To disable, pass a value larger than any expected request.
	Timeout time.Duration

	// MaxRetries is the number of retry attempts for idempotent requests on
	// retryable failures (408, 425, 429, 5xx, network errors). nil uses the
	// default of 3; pass [Int](0) to disable retries explicitly.
	MaxRetries *int

	// RetryDelay configures the exponential-backoff schedule. A zero value
	// uses [DefaultRetryDelay].
	RetryDelay RetryDelay

	// HTTPClient overrides the underlying *http.Client. Useful for injecting
	// a custom transport, proxy, or timeout. When nil, a fresh
	// http.DefaultClient is used.
	HTTPClient *http.Client

	// Logger receives debug-level events. When nil, no logging occurs. The
	// SDK never logs the API key or request/response bodies.
	Logger Logger

	// Telemetry, when non-nil, overrides the default opt-in behavior. Pass
	// [Bool](false) to disable.
	Telemetry *bool
}

// RetryDelay configures the exponential-backoff schedule. Backoff doubles each
// attempt, capped at [RetryDelay.Max]. When [RetryDelay.Jitter] is true the
// SDK picks a random value in [0, capped].
type RetryDelay struct {
	Initial time.Duration
	Max     time.Duration
	Jitter  bool
}

// DefaultRetryDelay is the schedule applied when [Config.RetryDelay] is the
// zero value: 500ms initial, 8s cap, jitter on.
var DefaultRetryDelay = RetryDelay{
	Initial: 500 * time.Millisecond,
	Max:     8 * time.Second,
	Jitter:  true,
}

// Logger is an optional sink for SDK-internal debug events. The interface is
// minimal on purpose: any logger that exposes a key/value method (slog,
// logrus, zap, etc.) is trivial to adapt.
type Logger interface {
	Debug(msg string, kv ...any)
}
