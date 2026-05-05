package threecommon_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	threecommon "github.com/3-Common/sdk/sdk-go"
)

func TestAPIError_FormatsWithRequestID(t *testing.T) {
	t.Parallel()

	e := &threecommon.APIError{
		Code:      "not_found",
		Message:   "Event evt_1 not found",
		RequestID: "req-abc-123",
	}
	assert.Equal(t, "[not_found] Event evt_1 not found (request_id=req-abc-123)", e.Error())
}

func TestAPIError_FormatsWithoutRequestID(t *testing.T) {
	t.Parallel()

	e := &threecommon.APIError{Code: "request_failed", Message: "boom"}
	assert.Equal(t, "[request_failed] boom", e.Error())
}

func TestAPIError_UnwrapReturnsCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("dial tcp: timeout")
	e := &threecommon.APIError{Code: "connection_error", Message: "boom", Cause: cause}

	assert.Same(t, cause, errors.Unwrap(e))
	assert.True(t, errors.Is(e, cause))
}

func TestTypedErrors_MatchViaErrorsAs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
		dest any
		want bool
	}{
		{"auth", &threecommon.AuthError{APIError: &threecommon.APIError{Code: "unauthorized"}}, new(*threecommon.AuthError), true},
		{"permission", &threecommon.PermissionError{APIError: &threecommon.APIError{Code: "forbidden"}}, new(*threecommon.PermissionError), true},
		{"not_found", &threecommon.NotFoundError{APIError: &threecommon.APIError{Code: "not_found"}}, new(*threecommon.NotFoundError), true},
		{"validation", &threecommon.ValidationError{APIError: &threecommon.APIError{Code: "validation_failed"}}, new(*threecommon.ValidationError), true},
		{"conflict", &threecommon.ConflictError{APIError: &threecommon.APIError{Code: "conflict"}}, new(*threecommon.ConflictError), true},
		{"server", &threecommon.ServerError{APIError: &threecommon.APIError{Code: "internal_error"}}, new(*threecommon.ServerError), true},
		{"connection", &threecommon.ConnectionError{APIError: &threecommon.APIError{Code: "connection_error"}}, new(*threecommon.ConnectionError), true},
		{"mismatch", &threecommon.NotFoundError{APIError: &threecommon.APIError{Code: "not_found"}}, new(*threecommon.AuthError), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := errors.As(tc.err, tc.dest)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestRateLimitError_CarriesRetryAfter(t *testing.T) {
	t.Parallel()

	rl := &threecommon.RateLimitError{
		APIError:   &threecommon.APIError{Code: "rate_limit_exceeded", Message: "slow down"},
		RetryAfter: 5 * time.Second,
	}

	var target *threecommon.RateLimitError
	require.True(t, errors.As(rl, &target))
	assert.Equal(t, 5*time.Second, target.RetryAfter)
	assert.Equal(t, "rate_limit_exceeded", target.Code) // promoted from embedded *APIError
}

func TestTypedError_PromotesBaseFields(t *testing.T) {
	t.Parallel()

	nf := &threecommon.NotFoundError{APIError: &threecommon.APIError{
		Code:       "not_found",
		Message:    "missing",
		HTTPStatus: 404,
		RequestID:  "req-xyz",
		Details:    map[string]any{"id": "evt_999"},
	}}

	assert.Equal(t, "not_found", nf.Code)
	assert.Equal(t, 404, nf.HTTPStatus)
	assert.Equal(t, "req-xyz", nf.RequestID)
	assert.Equal(t, "evt_999", nf.Details["id"])
	assert.Contains(t, nf.Error(), "request_id=req-xyz")
}

func TestTypedError_FieldsAccessibleViaPromotion(t *testing.T) {
	t.Parallel()

	// Embedded *APIError fields read directly off the typed wrapper — no
	// `.APIError.RequestID` chain needed.
	src := &threecommon.NotFoundError{APIError: &threecommon.APIError{Code: "not_found", RequestID: "req-1"}}
	assert.Equal(t, "not_found", src.Code)
	assert.Equal(t, "req-1", src.RequestID)
}
