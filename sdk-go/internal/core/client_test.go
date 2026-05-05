package core_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/internal/core"
)

func newClient(srv *httptest.Server, opts ...func(*core.ClientOptions)) *core.Client {
	o := core.ClientOptions{
		APIKey:     "3co_test",
		BaseURL:    srv.URL,
		APIVersion: threecommon.APIVersion,
		SDKVersion: threecommon.Version,
		Timeout:    2 * time.Second,
		Retry: core.RetryPolicy{
			MaxRetries: 0,
			Initial:    1 * time.Millisecond,
			Max:        5 * time.Millisecond,
			Jitter:     false,
		},
		HTTPClient: srv.Client(),
		Telemetry:  core.NewTelemetry(true),
		SleepFunc: func(_ context.Context, _ time.Duration) error {
			return nil
		},
	}
	for _, fn := range opts {
		fn(&o)
	}
	return core.NewClient(o)
}

func TestClient_DecodesSuccessResponse(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"evt_1"}],"hasMore":false}`))
	}))
	defer srv.Close()

	c := newClient(srv)

	var out struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		HasMore bool `json:"hasMore"`
	}
	err := c.Do(context.Background(), core.Request{Method: "GET", Path: "/events", Out: &out})
	require.NoError(t, err)
	require.Len(t, out.Data, 1)
	assert.Equal(t, "evt_1", out.Data[0].ID)
	assert.False(t, out.HasMore)
}

func TestClient_AppendsAPIPathAndQuery(t *testing.T) {
	t.Parallel()

	var seenPath string
	var seenQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		seenQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newClient(srv)
	err := c.Do(context.Background(), core.Request{
		Method: "GET",
		Path:   "/events",
		Query:  map[string]string{"page": "0", "status": "open"},
	})
	require.NoError(t, err)
	assert.Equal(t, "/v1/events", seenPath)
	assert.Equal(t, "page=0&status=open", seenQuery)
}

func TestClient_MapsTypedErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		status int
		body   string
		assert func(t *testing.T, err error)
	}{
		{401, `{"error":{"code":"unauthorized","message":"bad key"}}`, func(t *testing.T, err error) {
			t.Helper()
			var e *threecommon.AuthError
			require.True(t, errors.As(err, &e))
			assert.Equal(t, "unauthorized", e.Code)
		}},
		{403, `{"error":{"code":"forbidden","message":"no scope"}}`, func(t *testing.T, err error) {
			t.Helper()
			var e *threecommon.PermissionError
			require.True(t, errors.As(err, &e))
		}},
		{404, `{"error":{"code":"not_found","message":"missing"}}`, func(t *testing.T, err error) {
			t.Helper()
			var e *threecommon.NotFoundError
			require.True(t, errors.As(err, &e))
			assert.Equal(t, 404, e.HTTPStatus)
		}},
		{409, `{"error":{"code":"conflict","message":"clash"}}`, func(t *testing.T, err error) {
			t.Helper()
			var e *threecommon.ConflictError
			require.True(t, errors.As(err, &e))
		}},
		{422, `{"error":{"code":"validation_failed","message":"bad"}}`, func(t *testing.T, err error) {
			t.Helper()
			var e *threecommon.ValidationError
			require.True(t, errors.As(err, &e))
		}},
		{500, `{"error":{"code":"internal_error","message":"boom"}}`, func(t *testing.T, err error) {
			t.Helper()
			var e *threecommon.ServerError
			require.True(t, errors.As(err, &e))
			assert.Equal(t, 500, e.HTTPStatus)
		}},
		{418, ``, func(t *testing.T, err error) { // unknown 4xx
			t.Helper()
			var e *threecommon.ValidationError
			require.True(t, errors.As(err, &e))
		}},
	}

	for _, tc := range cases {
		t.Run(strconv.Itoa(tc.status), func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			}))
			defer srv.Close()

			c := newClient(srv, func(o *core.ClientOptions) {
				o.Retry.MaxRetries = 0 // disable retries for 5xx
			})
			err := c.Do(context.Background(), core.Request{Method: "GET", Path: "/events"})
			tc.assert(t, err)
		})
	}
}

func TestClient_RateLimit_CarriesRetryAfter(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "7")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = io.WriteString(w, `{"error":{"code":"rate_limit_exceeded","message":"slow"}}`)
	}))
	defer srv.Close()

	c := newClient(srv)
	err := c.Do(context.Background(), core.Request{Method: "GET", Path: "/events"})

	var rl *threecommon.RateLimitError
	require.True(t, errors.As(err, &rl))
	assert.Equal(t, 7*time.Second, rl.RetryAfter)
}

func TestClient_RetriesIdempotentOn500(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(w, `{"error":{"code":"internal_error","message":"first"}}`)
			return
		}
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	c := newClient(srv, func(o *core.ClientOptions) {
		o.Retry.MaxRetries = 1
	})

	err := c.Do(context.Background(), core.Request{Method: "GET", Path: "/events"})
	require.NoError(t, err)
	assert.Equal(t, int32(2), calls.Load())
}

func TestClient_DoesNotRetryNonIdempotent(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":{"code":"internal_error","message":"boom"}}`)
	}))
	defer srv.Close()

	c := newClient(srv, func(o *core.ClientOptions) {
		o.Retry.MaxRetries = 3
	})

	err := c.Do(context.Background(), core.Request{Method: "POST", Path: "/events"})
	var server *threecommon.ServerError
	require.True(t, errors.As(err, &server))
	assert.Equal(t, int32(1), calls.Load())
}

func TestClient_RetriesPostWithIdempotencyKey(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "key-1", r.Header.Get("Idempotency-Key"))
		n := calls.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newClient(srv, func(o *core.ClientOptions) {
		o.Retry.MaxRetries = 1
	})

	err := c.Do(context.Background(), core.Request{
		Method:         "POST",
		Path:           "/events",
		IdempotencyKey: "key-1",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(2), calls.Load())
}

func TestClient_NonRetryableStatusBubblesUp(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":{"code":"validation_failed","message":"bad"}}`)
	}))
	defer srv.Close()

	c := newClient(srv, func(o *core.ClientOptions) {
		o.Retry.MaxRetries = 5
	})
	err := c.Do(context.Background(), core.Request{Method: "GET", Path: "/events"})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, int32(1), calls.Load())
}

func TestClient_ContextCancellationReturnsConnectionError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		time.Sleep(500 * time.Millisecond)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := newClient(srv)
	err := c.Do(ctx, core.Request{Method: "GET", Path: "/events"})
	var ce *threecommon.ConnectionError
	require.True(t, errors.As(err, &ce))
}

func TestClient_TransportFailureBecomesConnectionError(t *testing.T) {
	t.Parallel()

	// Listener-less URL — connection refused.
	o := core.ClientOptions{
		APIKey:     "k",
		BaseURL:    "http://127.0.0.1:1",
		APIVersion: threecommon.APIVersion,
		SDKVersion: threecommon.Version,
		HTTPClient: &http.Client{Timeout: 50 * time.Millisecond},
		Telemetry:  core.NewTelemetry(false),
		Retry:      core.RetryPolicy{MaxRetries: 0},
	}
	c := core.NewClient(o)
	err := c.Do(context.Background(), core.Request{Method: "GET", Path: "/events"})
	var ce *threecommon.ConnectionError
	require.True(t, errors.As(err, &ce))
	assert.NotNil(t, ce.Cause)
}

func TestClient_TransportFailureRetriesIdempotent(t *testing.T) {
	t.Parallel()

	// First call gets a closed listener; second call gets a live one. We
	// simulate this with a counter inside a single handler that the first
	// time hijacks and closes.
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			hj, _ := w.(http.Hijacker)
			conn, _, _ := hj.Hijack()
			_ = conn.Close()
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newClient(srv, func(o *core.ClientOptions) {
		o.Retry.MaxRetries = 1
	})
	err := c.Do(context.Background(), core.Request{Method: "GET", Path: "/events"})
	require.NoError(t, err)
	assert.Equal(t, int32(2), calls.Load())
}

func TestClient_HonorsServerRetryAfter(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	var sleptDurations []time.Duration
	c := newClient(srv, func(o *core.ClientOptions) {
		o.Retry.MaxRetries = 1
		o.Retry.Max = 10 * time.Second // ensure Retry-After of 1s is not capped
		o.SleepFunc = func(_ context.Context, d time.Duration) error {
			sleptDurations = append(sleptDurations, d)
			return nil
		}
	})

	err := c.Do(context.Background(), core.Request{Method: "GET", Path: "/events"})
	require.NoError(t, err)
	require.Len(t, sleptDurations, 1)
	assert.Equal(t, 1*time.Second, sleptDurations[0])
}

func TestClient_PerRequestMaxRetriesOverride(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newClient(srv, func(o *core.ClientOptions) {
		o.Retry.MaxRetries = 5
	})

	err := c.Do(context.Background(), core.Request{
		Method:     "GET",
		Path:       "/events",
		MaxRetries: -1, // disable retries for this call
	})
	require.Error(t, err)
	assert.Equal(t, int32(1), calls.Load())
}

func TestClient_PerRequestTimeoutOverride(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer srv.Close()

	c := newClient(srv, func(o *core.ClientOptions) {
		o.Timeout = 5 * time.Second // global is fine but…
	})

	err := c.Do(context.Background(), core.Request{
		Method:  "GET",
		Path:    "/events",
		Timeout: 20 * time.Millisecond, // …per-call dominates
	})
	var ce *threecommon.ConnectionError
	require.True(t, errors.As(err, &ce))
}

func TestClient_LoggerInvokedOnEveryRequest(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	logger := &recordingLogger{}
	c := newClient(srv, func(o *core.ClientOptions) {
		o.Logger = logger
	})
	require.NoError(t, c.Do(context.Background(), core.Request{Method: "GET", Path: "/events"}))
	assert.Len(t, logger.entries, 1)
}

type recordingLogger struct {
	entries []logEntry
}
type logEntry struct {
	msg string
	kv  []any
}

func (l *recordingLogger) Debug(msg string, kv ...any) {
	l.entries = append(l.entries, logEntry{msg: msg, kv: kv})
}
