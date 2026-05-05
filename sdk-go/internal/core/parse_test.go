package core_test

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/3-Common/sdk/sdk-go/internal/core"
)

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func (errReader) Close() error             { return nil }

func TestReadResponse_BuffersBody(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
		Header:     http.Header{"X-Request-Id": []string{"req-1"}},
	}
	got, err := core.ReadResponse(resp)
	require.NoError(t, err)
	assert.Equal(t, 200, got.Status)
	assert.Equal(t, `{"ok":true}`, got.BodyText)
	assert.Equal(t, "req-1", got.RequestID)
}

func TestReadResponse_BodyReadFailureSurfaces(t *testing.T) {
	t.Parallel()

	resp := &http.Response{StatusCode: 500, Body: errReader{}, Header: http.Header{}}
	_, err := core.ReadResponse(resp)
	assert.Error(t, err)
}

func TestParseSuccessBody_DecodesJSON(t *testing.T) {
	t.Parallel()

	r := &core.Response{BodyText: `{"id":"evt_1"}`}
	var out struct {
		ID string `json:"id"`
	}
	require.NoError(t, core.ParseSuccessBody(r, &out))
	assert.Equal(t, "evt_1", out.ID)
}

func TestParseSuccessBody_NilOutOrEmptyBody(t *testing.T) {
	t.Parallel()

	r := &core.Response{BodyText: ""}
	require.NoError(t, core.ParseSuccessBody(r, &struct{}{}))

	r2 := &core.Response{BodyText: `{"id":"1"}`}
	require.NoError(t, core.ParseSuccessBody(r2, nil))
}

func TestParseSuccessBody_MalformedJSONReturnsError(t *testing.T) {
	t.Parallel()

	r := &core.Response{BodyText: `{not-json`}
	var out struct{}
	assert.Error(t, core.ParseSuccessBody(r, &out))
}

func TestParseErrorBody(t *testing.T) {
	t.Parallel()

	body := `{"error":{"code":"not_found","message":"missing","details":{"id":"evt_1"}}}`
	code, msg, details := core.ParseErrorBody(body)
	assert.Equal(t, "not_found", code)
	assert.Equal(t, "missing", msg)
	assert.Equal(t, "evt_1", details["id"])
}

func TestParseErrorBody_EmptyOrMalformedReturnsZeroValues(t *testing.T) {
	t.Parallel()

	code, msg, details := core.ParseErrorBody("")
	assert.Empty(t, code)
	assert.Empty(t, msg)
	assert.Nil(t, details)

	code, msg, details = core.ParseErrorBody(`{not-json`)
	assert.Empty(t, code)
	assert.Empty(t, msg)
	assert.Nil(t, details)
}

func TestParseRetryAfter_DeltaSeconds(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 5*time.Second, core.ParseRetryAfter("5"))
	assert.Equal(t, 1500*time.Millisecond, core.ParseRetryAfter("1.5"))
}

func TestParseRetryAfter_NegativeAndEmpty(t *testing.T) {
	t.Parallel()

	assert.Equal(t, time.Duration(0), core.ParseRetryAfter(""))
	assert.Equal(t, time.Duration(0), core.ParseRetryAfter("-3"))
	assert.Equal(t, time.Duration(0), core.ParseRetryAfter("not a number"))
}

func TestParseRetryAfter_HTTPDate(t *testing.T) {
	t.Parallel()

	future := time.Now().Add(10 * time.Second).UTC().Format(http.TimeFormat)
	got := core.ParseRetryAfter(future)
	assert.GreaterOrEqual(t, got, 5*time.Second)
	assert.LessOrEqual(t, got, 11*time.Second)

	past := time.Now().Add(-1 * time.Hour).UTC().Format(http.TimeFormat)
	assert.Equal(t, time.Duration(0), core.ParseRetryAfter(past))
}
