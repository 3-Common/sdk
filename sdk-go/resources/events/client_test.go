package events_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/filters"
	"github.com/3-Common/sdk/sdk-go/internal/core"
	"github.com/3-Common/sdk/sdk-go/resources/events"
)

// newTestClient returns an events.Client whose backend points at the supplied
// httptest server.
func newTestClient(t *testing.T, srv *httptest.Server) *events.Client {
	t.Helper()
	cl, err := events.New(threecommon.Config{
		APIKey:     "3co_test",
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		MaxRetries: threecommon.Int(0),
	})
	require.NoError(t, err)
	return cl
}

func TestNew_RequiresAPIKey(t *testing.T) {
	t.Setenv("THREECOMMON_API_KEY", "")
	_, err := events.New(threecommon.Config{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_api_key", v.Code)
}

func TestList_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/events", r.URL.Path)
		assert.Equal(t, "open", r.URL.Query().Get("status"))
		assert.Equal(t, "10", r.URL.Query().Get("pageSize"))
		assert.Equal(t, "Bearer 3co_test", r.Header.Get("Authorization"))
		w.Header().Set("X-Request-Id", "req-list-1")
		_, _ = w.Write([]byte(`{"data":[{"id":"evt_a","name":"A","status":"open"}],"hasMore":false}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	pageSize := 10
	got, err := cl.List(context.Background(), &events.ListParams{
		Status:   events.StatusOpen,
		PageSize: &pageSize,
	})
	require.NoError(t, err)
	require.Len(t, got.Data, 1)
	assert.Equal(t, "evt_a", got.Data[0].ID)
	assert.Equal(t, events.StatusOpen, got.Data[0].Status)
	assert.False(t, got.HasMore)
}

func TestList_NilParamsAccepted(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.RawQuery)
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.List(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, got.Data)
}

func TestList_AllListParamsEncoded(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.Equal(t, "0", q.Get("page"))
		assert.Equal(t, "5", q.Get("pageSize"))
		assert.Equal(t, "open", q.Get("status"))
		assert.Equal(t, "music", q.Get("search"))
		assert.Equal(t, "2026-12-31", q.Get("startBefore"))
		assert.Equal(t, "2026-01-01", q.Get("startAfter"))
		assert.Equal(t, "name", q.Get("sortField"))
		assert.Equal(t, "asc", q.Get("sortDirection"))
		assert.Equal(t, "id,name", q.Get("fields"))
		assert.NotEmpty(t, q.Get("filters"))
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	page := 0
	pageSize := 5
	f := filters.And(filters.Field("status").IsEqualTo("open"))
	_, err := cl.List(context.Background(), &events.ListParams{
		Page:          &page,
		PageSize:      &pageSize,
		Status:        events.StatusOpen,
		Search:        "music",
		StartBefore:   "2026-12-31",
		StartAfter:    "2026-01-01",
		SortField:     "name",
		SortDirection: "asc",
		Fields:        "id,name",
		Filters:       f.MustSerialize(),
	})
	require.NoError(t, err)
}

func TestRetrieve_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/events/evt_123", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"id":"evt_123","name":"Demo"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Retrieve(context.Background(), "evt_123", nil)
	require.NoError(t, err)
	assert.Equal(t, "evt_123", got.ID)
	assert.Equal(t, "Demo", got.Name)
}

func TestRetrieve_AppliesFieldsParam(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "id,name", r.URL.Query().Get("fields"))
		_, _ = w.Write([]byte(`{"data":{"id":"evt_1"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Retrieve(context.Background(), "evt_1", &events.RetrieveParams{Fields: "id,name"})
	require.NoError(t, err)
}

func TestRetrieve_RequiresID(t *testing.T) {
	t.Parallel()

	cl, _ := events.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Retrieve(context.Background(), "", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestRetrieve_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Request-Id", "req-404")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"missing"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Retrieve(context.Background(), "evt_missing", nil)

	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
	assert.Equal(t, "not_found", nf.Code)
	assert.Equal(t, "req-404", nf.RequestID)
}

func TestUpdate_SendsBodyAndDecodes(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/v1/events/evt_1", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]string
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "Renamed", got["name"])

		_, _ = w.Write([]byte(`{"data":{"id":"evt_1","name":"Renamed"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Update(context.Background(), "evt_1", &events.UpdateParams{Name: threecommon.String("Renamed")})
	require.NoError(t, err)
	assert.Equal(t, "Renamed", got.Name)
}

func TestUpdate_ValidatesID(t *testing.T) {
	t.Parallel()

	cl, _ := events.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Update(context.Background(), "", &events.UpdateParams{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestUpdate_ValidatesParams(t *testing.T) {
	t.Parallel()

	cl, _ := events.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Update(context.Background(), "evt_1", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestListAutoPaginate_WalksEveryPage(t *testing.T) {
	t.Parallel()

	pages := []string{
		`{"data":[{"id":"evt_1"},{"id":"evt_2"}],"hasMore":true}`,
		`{"data":[{"id":"evt_3"}],"hasMore":false}`,
	}

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx, _ := strconv.Atoi(r.URL.Query().Get("page"))
		calls.Add(1)
		require.Less(t, idx, len(pages))
		assert.Equal(t, "open", r.URL.Query().Get("status"))
		_, _ = io.WriteString(w, pages[idx])
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	iter := cl.ListAutoPaginate(context.Background(), &events.ListParams{Status: events.StatusOpen})

	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	require.NoError(t, iter.Err())
	assert.Equal(t, []string{"evt_1", "evt_2", "evt_3"}, ids)
	assert.Equal(t, int32(2), calls.Load())
}

func TestListAutoPaginate_HonorsExplicitStartPage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "5", r.URL.Query().Get("page"))
		_, _ = io.WriteString(w, `{"data":[{"id":"evt_5_a"}],"hasMore":false}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	startPage := 5
	iter := cl.ListAutoPaginate(context.Background(), &events.ListParams{Page: &startPage})

	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	require.NoError(t, iter.Err())
	assert.Equal(t, []string{"evt_5_a"}, ids)
}

func TestListAutoPaginate_SurfacesPageError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":{"code":"internal_error","message":"boom"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	iter := cl.ListAutoPaginate(context.Background(), nil)

	for iter.Next() { /* no values yielded before error */
	}
	require.Error(t, iter.Err())
	var server *threecommon.ServerError
	require.True(t, errors.As(iter.Err(), &server))
}

func TestFilterWith_AppliesSerializedFilter(t *testing.T) {
	t.Parallel()

	f := filters.And(filters.Field("status").IsEqualTo("open"))
	params := (&events.ListParams{}).FilterWith(f)
	assert.NotEmpty(t, params.Filters)
	assert.Contains(t, params.Filters, `"field":"status"`)
}

func TestFilterWith_NilFilterIsNoOp(t *testing.T) {
	t.Parallel()

	params := (&events.ListParams{Filters: "x"}).FilterWith(nil)
	assert.Equal(t, "x", params.Filters)
}

func TestFromBackend_InternalConstructorUsable(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	// Round-trip: build via core.NewFromConfig directly, then wrap
	// with FromBackend. This is exactly the path the aggregator uses.
	backend, err := core.NewFromConfig(threecommon.Config{
		APIKey:     "3co_test",
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		MaxRetries: threecommon.Int(0),
	})
	require.NoError(t, err)

	cl := events.FromBackend(backend)
	require.NotNil(t, cl)
	_, err = cl.List(context.Background(), nil)
	require.NoError(t, err)
}

func TestList_500SurfacesAsServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":{"code":"internal_error","message":"boom"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.List(context.Background(), nil)

	var server *threecommon.ServerError
	require.True(t, errors.As(err, &server))
}

func TestUpdate_500SurfacesAsServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":{"code":"internal_error","message":"boom"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Update(context.Background(), "evt_1", &events.UpdateParams{Name: threecommon.String("X")})

	var server *threecommon.ServerError
	require.True(t, errors.As(err, &server))
}

func TestListAutoPaginate_ContextCancellationStopsIteration(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"data":[{"id":"evt_1"}],"hasMore":true}`)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cl := newTestClient(t, srv)
	iter := cl.ListAutoPaginate(ctx, nil)
	for iter.Next() { /* no values yielded */
	}
	require.Error(t, iter.Err())
}

func TestEncodeListParams_EmptyParamsReturnNil(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.RawQuery, "empty ListParams must produce no query string")
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.List(context.Background(), &events.ListParams{}) // every field zero-value
	require.NoError(t, err)
}
