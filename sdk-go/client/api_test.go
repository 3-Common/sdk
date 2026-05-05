package client_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/events"
)

func TestNew_RequiresAPIKey(t *testing.T) {
	t.Setenv("THREECOMMON_API_KEY", "")
	_, err := client.New(threecommon.Config{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
}

func TestNew_PopulatesEvents(t *testing.T) {
	t.Parallel()

	api, err := client.New(threecommon.Config{APIKey: "k"})
	require.NoError(t, err)
	assert.NotNil(t, api.Events)
}

func TestAPI_EndToEndAgainstHTTPTestServer(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/events", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":[{"id":"evt_a"}],"hasMore":false}`))
	}))
	defer srv.Close()

	api, err := client.New(threecommon.Config{
		APIKey:     "3co_test",
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		MaxRetries: threecommon.Int(0),
	})
	require.NoError(t, err)

	result, err := api.Events.List(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, result.Data, 1)
	assert.Equal(t, "evt_a", result.Data[0].ID)
}

func TestDisableTelemetry_StopsHeader(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	var sawHeaderBeforeDisable, sawHeaderAfterDisable atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen := r.Header.Get("Threecommon-Client-Telemetry")
		switch calls.Add(1) {
		case 1:
			sawHeaderBeforeDisable.Store(seen != "")
		case 2:
			sawHeaderAfterDisable.Store(seen != "")
		}
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	api, err := client.New(threecommon.Config{
		APIKey:     "3co_test",
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		MaxRetries: threecommon.Int(0),
	})
	require.NoError(t, err)

	_, err = api.Events.List(context.Background(), nil)
	require.NoError(t, err)

	api.DisableTelemetry()

	_, err = api.Events.List(context.Background(), nil)
	require.NoError(t, err)

	assert.True(t, sawHeaderBeforeDisable.Load(), "expected telemetry header before Disable")
	assert.False(t, sawHeaderAfterDisable.Load(), "expected no telemetry header after Disable")
}

func TestAPI_BackendSharedAcrossResources(t *testing.T) {
	t.Parallel()

	// Smoke-test that disabling telemetry on the API affects whatever resource subsequently issues a request.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	api, err := client.New(threecommon.Config{
		APIKey:     "3co_test",
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		MaxRetries: threecommon.Int(0),
	})
	require.NoError(t, err)

	// `events.Client` accessed via api.Events should be the shared instance.
	assert.IsType(t, &events.Client{}, api.Events)

	_, err = api.Events.List(context.Background(), nil)
	require.NoError(t, err)
}
