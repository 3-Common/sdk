// Package core implements the HTTP machinery shared by every resource
// package. Decomposed into one concern per file: URL building, header
// building, send, retry policy, response parsing, and telemetry.
// Internal package — users cannot import it.
package core

import (
	"context"
	"errors"
	"net/http"
	"runtime"
	"time"

	threecommon "github.com/3-Common/sdk/sdk-go"
)

// ClientOptions configures a [*Client].
type ClientOptions struct {
	APIKey     string
	BaseURL    string
	APIVersion string
	SDKVersion string
	Timeout    time.Duration
	Retry      RetryPolicy
	HTTPClient *http.Client
	Telemetry  *Telemetry
	Logger     threecommon.Logger
	NowFunc    func() time.Time                                 // injectable for tests
	SleepFunc  func(ctx context.Context, d time.Duration) error // injectable for tests
}

// Client orchestrates URL building → header building → send → parse → error
// mapping → retry. One instance per [github.com/3-Common/sdk/sdk-go/client.API].
type Client struct {
	opts ClientOptions
}

// NewClient constructs a [*Client]. Defaults Sleep and Now when omitted.
func NewClient(opts ClientOptions) *Client {
	if opts.HTTPClient == nil {
		opts.HTTPClient = http.DefaultClient
	}
	if opts.NowFunc == nil {
		opts.NowFunc = time.Now
	}
	if opts.SleepFunc == nil {
		opts.SleepFunc = sleepWithContext
	}
	return &Client{opts: opts}
}

// Request describes one logical SDK call. The httpclient handles URL
// building, retries, and error mapping; the caller supplies path, method,
// query, and body.
type Request struct {
	Method         string
	Path           string
	Query          map[string]string
	Body           any
	Out            any           // pointer to decode 2xx body into
	IdempotencyKey string        // optional
	Timeout        time.Duration // overrides ClientOptions.Timeout when non-zero
	MaxRetries     int           // overrides ClientOptions.Retry.MaxRetries when non-zero (use -1 for "no retries")
}

// Do execute a [Request] honoring the client's retry policy. Returns a typed
// error from the threecommon package on failure. ctx is checked between
// retries; cancellation propagates immediately.
func (c *Client) Do(ctx context.Context, req Request) error {
	url := BuildURL(c.opts.BaseURL, threecommon.APIPath, req.Path, req.Query)

	maxRetries := c.opts.Retry.MaxRetries
	if req.MaxRetries != 0 {
		if req.MaxRetries < 0 {
			maxRetries = 0
		} else {
			maxRetries = req.MaxRetries
		}
	}
	idempotent := IsIdempotent(req.Method, req.IdempotencyKey != "")

	timeout := c.opts.Timeout
	if req.Timeout > 0 {
		timeout = req.Timeout
	}

	for attempt := 0; ; attempt++ {
		if err := ctx.Err(); err != nil {
			return wrapConnection(err.Error(), err)
		}

		start := c.opts.NowFunc()
		headers := BuildHeaders(HeadersInput{
			APIKey:          c.opts.APIKey,
			APIVersion:      c.opts.APIVersion,
			SDKVersion:      c.opts.SDKVersion,
			UserAgentSuffix: userAgentSuffix(),
			TelemetryHeader: c.opts.Telemetry.HeaderValue(c.opts.SDKVersion, c.opts.APIVersion),
			IdempotencyKey:  req.IdempotencyKey,
			HasBody:         req.Body != nil,
		})

		resp, sendErr := Send(ctx, SendInput{
			HTTPClient: c.opts.HTTPClient,
			URL:        url,
			Method:     req.Method,
			Headers:    headers,
			Body:       req.Body,
			Timeout:    timeout,
		})

		if sendErr != nil {
			if errors.Is(sendErr, context.Canceled) || errors.Is(sendErr, context.DeadlineExceeded) {
				return wrapConnection(sendErr.Error(), sendErr)
			}
			if idempotent && attempt < maxRetries {
				if err := c.opts.SleepFunc(ctx, ComputeBackoff(attempt, 0, c.opts.Retry)); err != nil {
					return wrapConnection(err.Error(), err)
				}
				continue
			}
			return wrapConnection(sendErr.Error(), sendErr)
		}

		duration := c.opts.NowFunc().Sub(start)
		c.opts.Telemetry.Record(req.Method, req.Path, resp.Status, duration)
		if c.opts.Logger != nil {
			c.opts.Logger.Debug(
				"threecommon:request",
				"method", req.Method,
				"path", req.Path,
				"status", resp.Status,
				"duration_ms", duration.Milliseconds(),
				"request_id", resp.RequestID,
				"attempt", attempt,
			)
		}

		if resp.Status >= 200 && resp.Status < 300 {
			return ParseSuccessBody(resp, req.Out)
		}

		retryAfter := ParseRetryAfter(resp.Header.Get("Retry-After"))
		if idempotent && attempt < maxRetries && IsRetryableStatus(resp.Status) {
			if err := c.opts.SleepFunc(ctx, ComputeBackoff(attempt, retryAfter, c.opts.Retry)); err != nil {
				return wrapConnection(err.Error(), err)
			}
			continue
		}

		return mapErrorResponse(resp, retryAfter)
	}
}

// mapErrorResponse converts a non-2xx [*Response] into the appropriate typed
// threecommon error.
func mapErrorResponse(resp *Response, retryAfter time.Duration) error {
	code, message, details := ParseErrorBody(resp.BodyText)
	if code == "" {
		code = defaultCodeForStatus(resp.Status)
	}
	if message == "" {
		message = defaultMessageForStatus(resp.Status)
	}

	base := &threecommon.APIError{
		Code:        code,
		Message:     message,
		HTTPStatus:  resp.Status,
		RequestID:   resp.RequestID,
		Details:     details,
		RawResponse: resp.BodyText,
	}

	switch {
	case resp.Status == http.StatusTooManyRequests:
		return &threecommon.RateLimitError{APIError: base, RetryAfter: retryAfter}
	case resp.Status == http.StatusUnauthorized:
		return &threecommon.AuthError{APIError: base}
	case resp.Status == http.StatusForbidden:
		return &threecommon.PermissionError{APIError: base}
	case resp.Status == http.StatusNotFound:
		return &threecommon.NotFoundError{APIError: base}
	case resp.Status == http.StatusConflict:
		return &threecommon.ConflictError{APIError: base}
	case resp.Status == http.StatusBadRequest, resp.Status == http.StatusUnprocessableEntity:
		return &threecommon.ValidationError{APIError: base}
	case resp.Status >= 500:
		return &threecommon.ServerError{APIError: base}
	}
	return &threecommon.ValidationError{APIError: base}
}

func defaultCodeForStatus(s int) string {
	switch s {
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "conflict"
	case http.StatusTooManyRequests:
		return "rate_limit_exceeded"
	}
	if s >= 500 {
		return "internal_error"
	}
	return "request_failed"
}

func defaultMessageForStatus(s int) string {
	return "Request failed with status " + http.StatusText(s)
}

// wrapConnection builds a *threecommon.ConnectionError from a transport-level
// failure. The result is always a non-nil error of concrete type
// *threecommon.ConnectionError.
func wrapConnection(message string, cause error) error {
	return &threecommon.ConnectionError{APIError: &threecommon.APIError{
		Code:    "connection_error",
		Message: message,
		Cause:   cause,
	}}
}

// userAgentSuffix returns the runtime + OS portion of the User-Agent.
func userAgentSuffix() string {
	return "Go/" + runtime.Version() + "; " + runtime.GOOS + "-" + runtime.GOARCH
}

// sleepWithContext sleeps for d, returning early if ctx is cancelled.
func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
