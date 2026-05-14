package core_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/3-Common/sdk/sdk-go/internal/core"
)

func TestBuildURL_TrimsTrailingSlashAndAddsLeadingSlash(t *testing.T) {
	t.Parallel()

	got := core.BuildURL("https://api.3common.com//", "/v1", "events", nil)
	assert.Equal(t, "https://api.3common.com/v1/events", got)
}

func TestBuildURL_StableQueryOrdering(t *testing.T) {
	t.Parallel()

	q := map[string]string{"status": "open", "page": "0", "pageSize": "50"}
	got := core.BuildURL("https://api.3common.com", "/v1", "/events", q)
	assert.Equal(t, "https://api.3common.com/v1/events?page=0&pageSize=50&status=open", got)
}

func TestBuildURL_OmitsEmptyValues(t *testing.T) {
	t.Parallel()

	q := map[string]string{"status": "", "page": "0"}
	got := core.BuildURL("https://api.3common.com", "/v1", "/events", q)
	assert.Equal(t, "https://api.3common.com/v1/events?page=0", got)
}

func TestBuildURL_NoQueryWhenEmpty(t *testing.T) {
	t.Parallel()

	got := core.BuildURL("https://api.3common.com", "/v1", "/events", nil)
	assert.Equal(t, "https://api.3common.com/v1/events", got)

	got = core.BuildURL("https://api.3common.com", "/v1", "/events", map[string]string{"a": ""})
	assert.Equal(t, "https://api.3common.com/v1/events", got)
}

func TestBuildURL_EncodesSpecialCharacters(t *testing.T) {
	t.Parallel()

	q := map[string]string{"search": "a b&c"}
	got := core.BuildURL("https://api.3common.com", "/v1", "/events", q)
	assert.Equal(t, "https://api.3common.com/v1/events?search=a+b%26c", got)
}
