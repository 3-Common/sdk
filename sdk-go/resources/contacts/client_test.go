package contacts_test

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
	"github.com/3-Common/sdk/sdk-go/internal/core"
	"github.com/3-Common/sdk/sdk-go/resources/contacts"
)

// newTestClient returns a contacts.Client whose backend points at the
// supplied httptest server.
func newTestClient(t *testing.T, srv *httptest.Server) *contacts.Client {
	t.Helper()
	cl, err := contacts.New(threecommon.Config{
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
	_, err := contacts.New(threecommon.Config{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_api_key", v.Code)
}

func TestFromBackend(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false,"pageNumber":0,"pageSize":20}`))
	}))
	defer srv.Close()

	backend, err := core.NewFromConfig(threecommon.Config{
		APIKey:     "3co_test",
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		MaxRetries: threecommon.Int(0),
	})
	require.NoError(t, err)

	cl := contacts.FromBackend(backend)
	require.NotNil(t, cl)
	_, err = cl.List(context.Background(), nil)
	require.NoError(t, err)
}

func TestList_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/contacts", r.URL.Path)
		assert.Equal(t, "opted-in", r.URL.Query().Get("filter"))
		_, _ = w.Write([]byte(`{
			"data":[{"id":"cnt_a","firstName":"A","lastName":"B","fullName":"A B","email":"a@example.com","vendorId":"hst_1","orderSum":0,"grossSum":0,"status":"opted-in","eventsAttended_IDS":[],"itemsPurchased_IDS":[],"productsPurchased_IDS":[]}],
			"hasMore":false,"pageNumber":0,"pageSize":20
		}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.List(context.Background(), &contacts.ListParams{Filter: contacts.QuickFilterOptedIn})
	require.NoError(t, err)
	require.Len(t, got.Data, 1)
	assert.Equal(t, "cnt_a", got.Data[0].ID)
	assert.Equal(t, contacts.StatusOptedIn, got.Data[0].Status)
	assert.False(t, got.HasMore)
}

func TestList_AllParamsEncoded(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.Equal(t, "0", q.Get("pageNumber"))
		assert.Equal(t, "100", q.Get("pageSize"))
		assert.Equal(t, "grossSum", q.Get("sortField"))
		assert.Equal(t, "desc", q.Get("sortDirection"))
		assert.Equal(t, "opted-in", q.Get("filter"))
		assert.Equal(t, "{}", q.Get("filters"))
		assert.Equal(t, "garcia", q.Get("search"))
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false,"pageNumber":0,"pageSize":100}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	page := 0
	size := 100
	_, err := cl.List(context.Background(), &contacts.ListParams{
		PageNumber:    &page,
		PageSize:      &size,
		SortField:     "grossSum",
		SortDirection: "desc",
		Filter:        contacts.QuickFilterOptedIn,
		Filters:       "{}",
		Search:        "garcia",
	})
	require.NoError(t, err)
}

func TestList_NilParams(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.RawQuery)
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false,"pageNumber":0,"pageSize":20}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.List(context.Background(), nil)
	require.NoError(t, err)
}

func TestList_EmptyParamsReturnNil(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.RawQuery, "empty ListParams must produce no query string")
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false,"pageNumber":0,"pageSize":20}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.List(context.Background(), &contacts.ListParams{})
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

func TestCount_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/contacts/count", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"count":4823}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	count, err := cl.Count(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(4823), count)
}

func TestCount_ServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":{"code":"internal_error","message":"boom"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Count(context.Background())
	var server *threecommon.ServerError
	require.True(t, errors.As(err, &server))
}

func TestRetrieve_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/contacts/cnt_123", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"id":"cnt_123","firstName":"Alex","lastName":"Garcia","fullName":"Alex Garcia","email":"alex@example.com","vendorId":"hst_1","orderSum":0,"grossSum":0,"status":"opted-in","eventsAttended_IDS":[],"itemsPurchased_IDS":[],"productsPurchased_IDS":[]}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Retrieve(context.Background(), "cnt_123")
	require.NoError(t, err)
	assert.Equal(t, "cnt_123", got.ID)
	assert.Equal(t, "alex@example.com", got.Email)
}

func TestRetrieve_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := contacts.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Retrieve(context.Background(), "")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestRetrieve_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"missing"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Retrieve(context.Background(), "cnt_missing")
	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
}

func TestCreate_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/contacts", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "alex@example.com", got["email"])
		assert.Equal(t, "Alex", got["firstName"])

		_, _ = w.Write([]byte(`{"data":{"id":"cnt_new","firstName":"Alex","lastName":"","fullName":"Alex","email":"alex@example.com","vendorId":"hst_1","orderSum":0,"grossSum":0,"status":"opted-in","eventsAttended_IDS":[],"itemsPurchased_IDS":[],"productsPurchased_IDS":[]}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Create(context.Background(), &contacts.CreateParams{
		Email:     "alex@example.com",
		FirstName: "Alex",
	})
	require.NoError(t, err)
	assert.Equal(t, "cnt_new", got.ID)
}

func TestCreate_RequiresParams(t *testing.T) {
	t.Parallel()
	cl, _ := contacts.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Create(context.Background(), nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestCreate_409SurfacesAsConflict(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = io.WriteString(w, `{"error":{"code":"conflict","message":"duplicate email"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Create(context.Background(), &contacts.CreateParams{Email: "a@example.com"})
	var conflict *threecommon.ConflictError
	require.True(t, errors.As(err, &conflict))
}

func TestUpdate_SendsBodyAndReturnsOrderDetails(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/v1/contacts/cnt_123", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		contact, _ := got["contact"].(map[string]any)
		assert.Equal(t, "Alex", contact["firstName"])
		assert.Equal(t, "opted-in", contact["status"])

		_, _ = w.Write([]byte(`{"data":{"_id":"cnt_123","email":"alex@example.com","vendorId":"hst_1","firstName":"Alex","lastName":"Garcia","fullName":"Alex Garcia","status":"opted-in","grossSum":0,"orderSum":0,"events_attended":[],"items_purchased":[],"products_purchased":[]}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Update(context.Background(), "cnt_123", &contacts.UpdateParams{
		Contact: contacts.ContactUpdate{
			FirstName: "Alex",
			LastName:  "Garcia",
			Email:     "alex@example.com",
			Status:    contacts.StatusOptedIn,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "cnt_123", got.ID)
}

func TestUpdate_WithMergeResolution(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "cnt_456", got["mergeWith"])
		assert.Equal(t, "safe-merge", got["resolution"])

		_, _ = w.Write([]byte(`{"data":{"_id":"cnt_123","email":"a@example.com","vendorId":"hst_1","firstName":"A","lastName":"G","fullName":"A G","status":"opted-in","grossSum":0,"orderSum":0,"events_attended":[],"items_purchased":[],"products_purchased":[]}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Update(context.Background(), "cnt_123", &contacts.UpdateParams{
		Contact: contacts.ContactUpdate{
			FirstName: "A", LastName: "G", Email: "a@example.com", Status: contacts.StatusOptedIn,
		},
		MergeWith:  "cnt_456",
		Resolution: contacts.MergeResolutionSafe,
	})
	require.NoError(t, err)
}

func TestUpdate_ValidatesID(t *testing.T) {
	t.Parallel()
	cl, _ := contacts.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Update(context.Background(), "", &contacts.UpdateParams{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestUpdate_ValidatesParams(t *testing.T) {
	t.Parallel()
	cl, _ := contacts.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Update(context.Background(), "cnt_1", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestDelete_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v1/contacts/cnt_123", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"id":"cnt_123"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Delete(context.Background(), "cnt_123")
	require.NoError(t, err)
	assert.Equal(t, "cnt_123", got.ID)
}

func TestDelete_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := contacts.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Delete(context.Background(), "")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
}

func TestDelete_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"gone"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Delete(context.Background(), "cnt_missing")
	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
}

func TestBulkUpsert_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/contacts/bulk", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		items, _ := got["contacts"].([]any)
		assert.Len(t, items, 2)

		_, _ = w.Write([]byte(`{"data":{"affected":2}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.BulkUpsert(context.Background(), &contacts.BulkUpsertParams{
		Contacts: []contacts.BulkUpsertItem{
			{Email: "a@example.com", FirstName: "Ada"},
			{Email: "b@example.com", FirstName: "Beatrix"},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), got.Affected)
}

func TestBulkUpsert_RequiresParams(t *testing.T) {
	t.Parallel()
	cl, _ := contacts.New(threecommon.Config{APIKey: "k"})
	_, err := cl.BulkUpsert(context.Background(), nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestListActivity_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/contacts/cnt_123/activity", r.URL.Path)
		assert.Equal(t, "email_sent", r.URL.Query().Get("filter"))
		_, _ = w.Write([]byte(`{
			"data":[{"_id":"act_1","vendor_id":"hst_1","email":"alex@example.com","contact_id":"cnt_123","type":"email_sent","data":{},"createdAt":"2026-05-01T00:00:00.000Z","updatedAt":"2026-05-01T00:00:00.000Z"}],
			"hasMore":false,"pageNumber":0,"pageSize":20
		}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.ListActivity(context.Background(), "cnt_123", &contacts.ActivityListParams{
		Filter: contacts.ActivityEmailSent,
	})
	require.NoError(t, err)
	require.Len(t, got.Data, 1)
	assert.Equal(t, contacts.ActivityEmailSent, got.Data[0].Type)
}

func TestListActivity_AllParamsEncoded(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.Equal(t, "1", q.Get("pageNumber"))
		assert.Equal(t, "5", q.Get("pageSize"))
		assert.Equal(t, "email_sent", q.Get("filter"))
		assert.Equal(t, "oldest", q.Get("sort"))
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false,"pageNumber":1,"pageSize":5}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	page := 1
	size := 5
	_, err := cl.ListActivity(context.Background(), "cnt_123", &contacts.ActivityListParams{
		PageNumber: &page,
		PageSize:   &size,
		Filter:     contacts.ActivityEmailSent,
		Sort:       "oldest",
	})
	require.NoError(t, err)
}

func TestListActivity_NilParams(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.RawQuery)
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false,"pageNumber":0,"pageSize":20}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.ListActivity(context.Background(), "cnt_123", nil)
	require.NoError(t, err)
}

func TestListActivity_EmptyParamsReturnNil(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.RawQuery, "empty ActivityListParams must produce no query string")
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false,"pageNumber":0,"pageSize":20}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.ListActivity(context.Background(), "cnt_123", &contacts.ActivityListParams{})
	require.NoError(t, err)
}

func TestListActivity_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := contacts.New(threecommon.Config{APIKey: "k"})
	_, err := cl.ListActivity(context.Background(), "", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestListActivity_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"gone"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.ListActivity(context.Background(), "cnt_missing", nil)
	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
}

func TestListAutoPaginate_WalksEveryPage(t *testing.T) {
	t.Parallel()

	pages := []string{
		`{"data":[{"id":"cnt_1","firstName":"A","lastName":"B","fullName":"A B","email":"a@example.com","vendorId":"hst_1","orderSum":0,"grossSum":0,"status":"opted-in","eventsAttended_IDS":[],"itemsPurchased_IDS":[],"productsPurchased_IDS":[]},{"id":"cnt_2","firstName":"C","lastName":"D","fullName":"C D","email":"c@example.com","vendorId":"hst_1","orderSum":0,"grossSum":0,"status":"opted-in","eventsAttended_IDS":[],"itemsPurchased_IDS":[],"productsPurchased_IDS":[]}],"hasMore":true,"pageNumber":0,"pageSize":20}`,
		`{"data":[{"id":"cnt_3","firstName":"E","lastName":"F","fullName":"E F","email":"e@example.com","vendorId":"hst_1","orderSum":0,"grossSum":0,"status":"opted-in","eventsAttended_IDS":[],"itemsPurchased_IDS":[],"productsPurchased_IDS":[]}],"hasMore":false,"pageNumber":1,"pageSize":20}`,
	}

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx, _ := strconv.Atoi(r.URL.Query().Get("pageNumber"))
		calls.Add(1)
		require.Less(t, idx, len(pages))
		_, _ = io.WriteString(w, pages[idx])
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	iter := cl.ListAutoPaginate(context.Background(), nil)

	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	require.NoError(t, iter.Err())
	assert.Equal(t, []string{"cnt_1", "cnt_2", "cnt_3"}, ids)
	assert.Equal(t, int32(2), calls.Load())
}

func TestListAutoPaginate_HonorsExplicitStartPage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "5", r.URL.Query().Get("pageNumber"))
		_, _ = io.WriteString(w, `{"data":[{"id":"cnt_p5","firstName":"P","lastName":"5","fullName":"P 5","email":"p5@example.com","vendorId":"hst_1","orderSum":0,"grossSum":0,"status":"opted-in","eventsAttended_IDS":[],"itemsPurchased_IDS":[],"productsPurchased_IDS":[]}],"hasMore":false,"pageNumber":5,"pageSize":20}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	startPage := 5
	iter := cl.ListAutoPaginate(context.Background(), &contacts.ListParams{PageNumber: &startPage})

	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	require.NoError(t, iter.Err())
	assert.Equal(t, []string{"cnt_p5"}, ids)
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

func TestListAutoPaginate_ContextCancellationStopsIteration(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"data":[],"hasMore":true,"pageNumber":0,"pageSize":20}`)
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

func TestListActivityAutoPaginate_WalksEveryPage(t *testing.T) {
	t.Parallel()

	pages := []string{
		`{"data":[{"_id":"act_1","vendor_id":"hst_1","email":"a@example.com","type":"email_sent","data":{},"createdAt":"2026-05-01T00:00:00.000Z","updatedAt":"2026-05-01T00:00:00.000Z"}],"hasMore":true,"pageNumber":0,"pageSize":20}`,
		`{"data":[{"_id":"act_2","vendor_id":"hst_1","email":"a@example.com","type":"ticket_scanned","data":{},"createdAt":"2026-05-02T00:00:00.000Z","updatedAt":"2026-05-02T00:00:00.000Z"}],"hasMore":false,"pageNumber":1,"pageSize":20}`,
	}

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx, _ := strconv.Atoi(r.URL.Query().Get("pageNumber"))
		calls.Add(1)
		require.Less(t, idx, len(pages))
		_, _ = io.WriteString(w, pages[idx])
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	iter := cl.ListActivityAutoPaginate(context.Background(), "cnt_123", nil)

	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	require.NoError(t, iter.Err())
	assert.Equal(t, []string{"act_1", "act_2"}, ids)
}

func TestListActivityAutoPaginate_ValidatesIDOnIter(t *testing.T) {
	t.Parallel()

	cl, _ := contacts.New(threecommon.Config{APIKey: "k"})
	iter := cl.ListActivityAutoPaginate(context.Background(), "", nil)
	for iter.Next() { /* should not yield */
	}
	var v *threecommon.ValidationError
	require.True(t, errors.As(iter.Err(), &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestListActivityAutoPaginate_HonorsExplicitStartPage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "3", r.URL.Query().Get("pageNumber"))
		_, _ = io.WriteString(w, `{"data":[{"_id":"act_p3","vendor_id":"hst_1","email":"a@example.com","type":"email_sent","data":{},"createdAt":"2026-05-01T00:00:00.000Z","updatedAt":"2026-05-01T00:00:00.000Z"}],"hasMore":false,"pageNumber":3,"pageSize":20}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	startPage := 3
	iter := cl.ListActivityAutoPaginate(context.Background(), "cnt_123", &contacts.ActivityListParams{PageNumber: &startPage})
	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	require.NoError(t, iter.Err())
	assert.Equal(t, []string{"act_p3"}, ids)
}

func TestListActivityAutoPaginate_SurfacesPageError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":{"code":"internal_error","message":"boom"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	iter := cl.ListActivityAutoPaginate(context.Background(), "cnt_123", nil)
	for iter.Next() {
	}
	require.Error(t, iter.Err())
}

func TestListActivityAutoPaginate_ContextCancellation(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"data":[],"hasMore":true,"pageNumber":0,"pageSize":20}`)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cl := newTestClient(t, srv)
	iter := cl.ListActivityAutoPaginate(ctx, "cnt_123", nil)
	for iter.Next() {
	}
	require.Error(t, iter.Err())
}
