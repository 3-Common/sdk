package features_test

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
	"github.com/3-Common/sdk/sdk-go/resources/features"
)

const sampleFeature = `{
	"id": "feat_123",
	"hostId": "host_1",
	"key": "api_calls",
	"name": "API calls",
	"description": "Monthly API call quota",
	"type": "quantity",
	"active": true,
	"createdAt": "2026-05-01T00:00:00.000Z",
	"updatedAt": "2026-05-01T00:00:00.000Z"
}`

const sampleResolved = `{
	"feature": ` + sampleFeature + `,
	"value": {"type": "quantity", "quantity": 1000, "balance": 850},
	"contributingSubscriptionIds": ["sub_1"]
}`

func newTestClient(t *testing.T, srv *httptest.Server) *features.Client {
	t.Helper()
	cl, err := features.New(threecommon.Config{
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
	_, err := features.New(threecommon.Config{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_api_key", v.Code)
}

func TestList_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/features", r.URL.Path)
		assert.Equal(t, "quantity", r.URL.Query().Get("type"))
		assert.Equal(t, "true", r.URL.Query().Get("active"))
		_, _ = io.WriteString(w, `{"data":[`+sampleFeature+`],"hasMore":false}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.List(context.Background(), &features.ListParams{
		Type:   features.TypeQuantity,
		Active: threecommon.Bool(true),
	})
	require.NoError(t, err)
	require.Len(t, got.Data, 1)
	assert.Equal(t, "api_calls", got.Data[0].Key)
	assert.Equal(t, features.TypeQuantity, got.Data[0].Type)
	require.NotNil(t, got.Data[0].Active)
	assert.True(t, *got.Data[0].Active)
}

func TestList_NilParamsAccepted(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.RawQuery)
		_, _ = io.WriteString(w, `{"data":[],"hasMore":false}`)
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
		assert.Equal(t, "25", q.Get("pageSize"))
		assert.Equal(t, "enum", q.Get("type"))
		assert.Equal(t, "false", q.Get("active")) // booleans render lowercase
		assert.Equal(t, "id,key", q.Get("fields"))
		_, _ = io.WriteString(w, `{"data":[],"hasMore":false}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	page := 0
	pageSize := 25
	_, err := cl.List(context.Background(), &features.ListParams{
		Page:     &page,
		PageSize: &pageSize,
		Type:     features.TypeEnum,
		Active:   threecommon.Bool(false),
		Fields:   "id,key",
	})
	require.NoError(t, err)
}

func TestResolve_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/features/resolve", r.URL.Path)
		assert.Equal(t, "cnt_7", r.URL.Query().Get("contactId"))
		assert.Equal(t, "api_calls", r.URL.Query().Get("featureKey"))
		_, _ = io.WriteString(w, `{"data":`+sampleResolved+`}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Resolve(context.Background(), &features.ResolveParams{
		ContactID:  "cnt_7",
		FeatureKey: "api_calls",
	})
	require.NoError(t, err)
	assert.Equal(t, "api_calls", got.Feature.Key)
	assert.Equal(t, []string{"sub_1"}, got.ContributingSubscriptionIDs)
	assert.Equal(t, features.TypeQuantity, got.Value.Type)
	require.NotNil(t, got.Value.Quantity)
	assert.Equal(t, int64(1000), *got.Value.Quantity)
	require.NotNil(t, got.Value.Balance)
	assert.Equal(t, int64(850), *got.Value.Balance)
}

func TestResolve_RequiresParams(t *testing.T) {
	t.Parallel()

	cl, _ := features.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Resolve(context.Background(), nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestResolve_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"unknown feature"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Resolve(context.Background(), &features.ResolveParams{ContactID: "cnt_7", FeatureKey: "nope"})
	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
}

func TestRetrieve_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/features/feat_123", r.URL.Path)
		_, _ = io.WriteString(w, `{"data":`+sampleFeature+`}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Retrieve(context.Background(), "feat_123", nil)
	require.NoError(t, err)
	assert.Equal(t, "api_calls", got.Key)
}

func TestRetrieve_AppliesFieldsParam(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "id,key", r.URL.Query().Get("fields"))
		_, _ = io.WriteString(w, `{"data":{"id":"feat_1","key":"api_calls"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Retrieve(context.Background(), "feat_1", &features.RetrieveParams{Fields: "id,key"})
	require.NoError(t, err)
}

func TestRetrieve_RequiresID(t *testing.T) {
	t.Parallel()

	cl, _ := features.New(threecommon.Config{APIKey: "k"})
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
	_, err := cl.Retrieve(context.Background(), "feat_missing", nil)

	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
	assert.Equal(t, "req-404", nf.RequestID)
}

func TestCreate_SendsBodyAndDecodes(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/features", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got features.CreateParams
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "api_calls", got.Key)
		assert.Equal(t, "API calls", got.Name)
		assert.Equal(t, features.TypeQuantity, got.Type)

		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"data":`+sampleFeature+`}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Create(context.Background(), &features.CreateParams{
		Key:  "api_calls",
		Name: "API calls",
		Type: features.TypeQuantity,
	})
	require.NoError(t, err)
	assert.Equal(t, "feat_123", got.ID)
}

func TestCreate_RequiresParams(t *testing.T) {
	t.Parallel()

	cl, _ := features.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Create(context.Background(), nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestCreate_409Conflict(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = io.WriteString(w, `{"error":{"code":"conflict","message":"feature key exists"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Create(context.Background(), &features.CreateParams{
		Key:  "api_calls",
		Name: "API calls",
		Type: features.TypeQuantity,
	})
	var conflict *threecommon.ConflictError
	require.True(t, errors.As(err, &conflict))
	assert.Equal(t, "conflict", conflict.Code)
}

func TestUpdate_SendsBodyAndDecodes(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/v1/features/feat_123", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "API requests", got["name"])

		_, _ = io.WriteString(w, `{"data":{"id":"feat_123","name":"API requests"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Update(context.Background(), "feat_123", &features.UpdateParams{
		Name: threecommon.String("API requests"),
	})
	require.NoError(t, err)
	assert.Equal(t, "API requests", got.Name)
}

func TestUpdate_RequiresID(t *testing.T) {
	t.Parallel()

	cl, _ := features.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Update(context.Background(), "", &features.UpdateParams{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestUpdate_RequiresParams(t *testing.T) {
	t.Parallel()

	cl, _ := features.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Update(context.Background(), "feat_1", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestArchive_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/features/feat_123/archive", r.URL.Path)
		_, _ = io.WriteString(w, `{"data":{"id":"feat_123","active":false}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Archive(context.Background(), "feat_123")
	require.NoError(t, err)
	require.NotNil(t, got.Active)
	assert.False(t, *got.Active)
}

func TestUnarchive_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/features/feat_123/unarchive", r.URL.Path)
		_, _ = io.WriteString(w, `{"data":{"id":"feat_123","active":true}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Unarchive(context.Background(), "feat_123")
	require.NoError(t, err)
	require.NotNil(t, got.Active)
	assert.True(t, *got.Active)
}

func TestArchive_RequiresID(t *testing.T) {
	t.Parallel()

	cl, _ := features.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Archive(context.Background(), "")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestListAutoPaginate_WalksEveryPage(t *testing.T) {
	t.Parallel()

	pages := []string{
		`{"data":[{"id":"feat_1"},{"id":"feat_2"}],"hasMore":true}`,
		`{"data":[{"id":"feat_3"}],"hasMore":false}`,
	}

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx, _ := strconv.Atoi(r.URL.Query().Get("page"))
		calls.Add(1)
		require.Less(t, idx, len(pages))
		assert.Equal(t, "true", r.URL.Query().Get("active"))
		_, _ = io.WriteString(w, pages[idx])
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	iter := cl.ListAutoPaginate(context.Background(), &features.ListParams{Active: threecommon.Bool(true)})

	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	require.NoError(t, iter.Err())
	assert.Equal(t, []string{"feat_1", "feat_2", "feat_3"}, ids)
	assert.Equal(t, int32(2), calls.Load())
}
