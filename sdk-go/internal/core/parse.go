package core

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Response is a fully-buffered, post-read normalization of [*http.Response].
// Headers are kept as the canonical [http.Header] map; the body is read once
// into a string and never read again from the underlying response. Callers
// must not pass the wrapped *http.Response back to the network.
type Response struct {
	Status    int
	Header    http.Header
	BodyText  string
	RequestID string
}

// ReadResponse drains response body and returns a [Response]. The original
// [*http.Response] body is closed before this returns.
func ReadResponse(resp *http.Response) (*Response, error) {
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Response{
		Status:    resp.StatusCode,
		Header:    resp.Header,
		BodyText:  string(body),
		RequestID: resp.Header.Get("X-Request-Id"),
	}, nil
}

// ParseSuccessBody decodes a 2xx body into out. Empty or non-JSON bodies are
// silently ignored — out keeps its zero value. Returns a JSON error only when
// the body looks like JSON but is malformed.
func ParseSuccessBody(r *Response, out any) error {
	if out == nil || r.BodyText == "" {
		return nil
	}
	return json.Unmarshal([]byte(r.BodyText), out)
}

// ParseErrorBody best-effort-parses the API's standard error envelope.
// Returns ("", "", nil) when the body is empty or not the expected shape;
// callers should fall back to status-based defaults.
type errorEnvelope struct {
	Error struct {
		Code    string         `json:"code"`
		Message string         `json:"message"`
		Details map[string]any `json:"details"`
	} `json:"error"`
}

// ParseErrorBody returns the parsed code, message, and details from the
// standard {"error": {...}} response shape. Returns zero values when the body
// can't be parsed.
func ParseErrorBody(bodyText string) (code, message string, details map[string]any) {
	if bodyText == "" {
		return "", "", nil
	}
	var env errorEnvelope
	if err := json.Unmarshal([]byte(bodyText), &env); err != nil {
		return "", "", nil
	}
	return env.Error.Code, env.Error.Message, env.Error.Details
}

// ParseRetryAfter parses a Retry-After header value. Accepts either
// delta-seconds or an HTTP-date. Returns 0 for missing, malformed, or
// already-elapsed values.
func ParseRetryAfter(header string) time.Duration {
	if header == "" {
		return 0
	}
	if secs, err := strconv.ParseFloat(header, 64); err == nil && secs >= 0 {
		return time.Duration(secs * float64(time.Second))
	}
	if t, err := http.ParseTime(header); err == nil {
		delta := time.Until(t)
		if delta < 0 {
			return 0
		}
		return delta
	}
	return 0
}
