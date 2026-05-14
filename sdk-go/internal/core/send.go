package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// SendInput captures everything Send needs. Pre-built so Send stays a thin
// wrapper around the standard library — no header building, no URL building,
// no retry logic.
type SendInput struct {
	HTTPClient *http.Client
	URL        string
	Method     string
	Headers    http.Header
	Body       any           // marshaled to JSON when non-nil
	Timeout    time.Duration // 0 disables the per-request timeout
}

// Send issues a single HTTP request and returns a fully-buffered [Response].
// The supplied ctx and Input.Timeout combine: whichever fires first cancels
// the request. Send does not retry; that is the caller's responsibility.
func Send(ctx context.Context, in SendInput) (*Response, error) {
	if in.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, in.Timeout)
		defer cancel()
	}

	var bodyReader *bytes.Reader
	if in.Body != nil {
		buf, err := json.Marshal(in.Body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(buf)
	}

	var req *http.Request
	var err error
	if bodyReader == nil {
		req, err = http.NewRequestWithContext(ctx, in.Method, in.URL, http.NoBody)
	} else {
		req, err = http.NewRequestWithContext(ctx, in.Method, in.URL, bodyReader)
	}
	if err != nil {
		return nil, err
	}
	req.Header = in.Headers

	resp, err := in.HTTPClient.Do(req) //nolint:bodyclose // ReadResponse closes the body; the linter doesn't see through the call
	if err != nil {
		// context errors propagate as-is so callers can branch on them.
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, err
	}
	return ReadResponse(resp)
}
