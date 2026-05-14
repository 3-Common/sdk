package core_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/3-Common/sdk/sdk-go/internal/core"
)

func TestSend_ReturnsBufferedResponse(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer k", r.Header.Get("Authorization"))
		w.Header().Set("X-Request-Id", "req-1")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	headers := http.Header{}
	headers.Set("Authorization", "Bearer k")

	resp, err := core.Send(context.Background(), core.SendInput{
		HTTPClient: srv.Client(),
		URL:        srv.URL,
		Method:     "GET",
		Headers:    headers,
	})
	require.NoError(t, err)
	assert.Equal(t, 200, resp.Status)
	assert.Equal(t, `{"data":[]}`, resp.BodyText)
	assert.Equal(t, "req-1", resp.RequestID)
}

func TestSend_MarshalsJSONBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, `{"name":"x"}`, string(body))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	_, err := core.Send(context.Background(), core.SendInput{
		HTTPClient: srv.Client(),
		URL:        srv.URL,
		Method:     "PATCH",
		Headers:    http.Header{},
		Body:       map[string]string{"name": "x"},
	})
	require.NoError(t, err)
}

func TestSend_RespectsContextCancellation(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := core.Send(ctx, core.SendInput{
		HTTPClient: srv.Client(),
		URL:        srv.URL,
		Method:     "GET",
		Headers:    http.Header{},
	})
	assert.ErrorIs(t, err, context.Canceled)
}

func TestSend_RespectsTimeout(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer srv.Close()

	_, err := core.Send(context.Background(), core.SendInput{
		HTTPClient: srv.Client(),
		URL:        srv.URL,
		Method:     "GET",
		Headers:    http.Header{},
		Timeout:    20 * time.Millisecond,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded), "want DeadlineExceeded, got %v", err)
}

func TestSend_MalformedURLReturnsError(t *testing.T) {
	t.Parallel()

	_, err := core.Send(context.Background(), core.SendInput{
		HTTPClient: http.DefaultClient,
		URL:        "://bad",
		Method:     "GET",
		Headers:    http.Header{},
	})
	assert.Error(t, err)
}

func TestSend_NonMarshalableBodyReturnsError(t *testing.T) {
	t.Parallel()

	// channels can't be JSON-marshaled
	_, err := core.Send(context.Background(), core.SendInput{
		HTTPClient: http.DefaultClient,
		URL:        "https://example.com",
		Method:     "POST",
		Headers:    http.Header{},
		Body:       make(chan int),
	})
	assert.Error(t, err)
}
