package entitlements_test

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
	"github.com/3-Common/sdk/sdk-go/resources/entitlements"
)

const sampleEntitlement = `{
	"id": "ent_123",
	"hostId": "host_1",
	"contactId": "cnt_7",
	"featureKey": "api_calls",
	"balance": 100,
	"grants": [
		{"id": "grant_1", "source": "manual", "amount": 100, "remaining": 100, "addedAt": "2026-05-01T18:00:00.000Z"}
	],
	"totalGranted": 100,
	"totalConsumed": 0,
	"metadata": {},
	"createdAt": "2026-05-01T18:00:00.000Z",
	"updatedAt": "2026-05-01T18:00:00.000Z"
}`

// newTestClient returns an entitlements.Client whose backend points at the
// supplied httptest server.
func newTestClient(t *testing.T, srv *httptest.Server) *entitlements.Client {
	t.Helper()
	cl, err := entitlements.New(threecommon.Config{
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
	_, err := entitlements.New(threecommon.Config{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_api_key", v.Code)
}

func TestList_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/entitlements", r.URL.Path)
		assert.Equal(t, "api_calls", r.URL.Query().Get("featureKey"))
		assert.Equal(t, "1", r.URL.Query().Get("minBalance"))
		assert.Equal(t, "Bearer 3co_test", r.Header.Get("Authorization"))
		_, _ = io.WriteString(w, `{"data":[`+sampleEntitlement+`],"hasMore":false}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.List(context.Background(), &entitlements.ListParams{
		FeatureKey: "api_calls",
		MinBalance: threecommon.Int64(1),
	})
	require.NoError(t, err)
	require.Len(t, got.Data, 1)
	assert.Equal(t, "ent_123", got.Data[0].ID)
	require.NotNil(t, got.Data[0].Balance)
	assert.Equal(t, int64(100), *got.Data[0].Balance)
	require.Len(t, got.Data[0].Grants, 1)
	assert.Equal(t, entitlements.GrantSourceManual, got.Data[0].Grants[0].Source)
	assert.False(t, got.HasMore)
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
		assert.Equal(t, "cnt_7", q.Get("contactId"))
		assert.Equal(t, "api_calls", q.Get("featureKey"))
		assert.Equal(t, "0", q.Get("minBalance"))
		assert.Equal(t, "id,balance", q.Get("fields"))
		_, _ = io.WriteString(w, `{"data":[],"hasMore":false}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	page := 0
	pageSize := 25
	_, err := cl.List(context.Background(), &entitlements.ListParams{
		Page:       &page,
		PageSize:   &pageSize,
		ContactID:  "cnt_7",
		FeatureKey: "api_calls",
		MinBalance: threecommon.Int64(0),
		Fields:     "id,balance",
	})
	require.NoError(t, err)
}

func TestRetrieve_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/entitlements/ent_123", r.URL.Path)
		_, _ = io.WriteString(w, `{"data":`+sampleEntitlement+`}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Retrieve(context.Background(), "ent_123", nil)
	require.NoError(t, err)
	assert.Equal(t, "ent_123", got.ID)
	assert.Equal(t, "api_calls", got.FeatureKey)
}

func TestRetrieve_AppliesFieldsParam(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "id,balance", r.URL.Query().Get("fields"))
		_, _ = io.WriteString(w, `{"data":{"id":"ent_1","balance":5}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Retrieve(context.Background(), "ent_1", &entitlements.RetrieveParams{Fields: "id,balance"})
	require.NoError(t, err)
	require.NotNil(t, got.Balance)
	assert.Equal(t, int64(5), *got.Balance)
}

func TestRetrieve_RequiresID(t *testing.T) {
	t.Parallel()

	cl, _ := entitlements.New(threecommon.Config{APIKey: "k"})
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
	_, err := cl.Retrieve(context.Background(), "ent_missing", nil)

	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
	assert.Equal(t, "not_found", nf.Code)
	assert.Equal(t, "req-404", nf.RequestID)
}

func TestLookup_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/entitlements/lookup", r.URL.Path)
		assert.Equal(t, "cnt_7", r.URL.Query().Get("contactId"))
		assert.Equal(t, "api_calls", r.URL.Query().Get("featureKey"))
		_, _ = io.WriteString(w, `{"data":`+sampleEntitlement+`}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Lookup(context.Background(), &entitlements.LookupParams{
		ContactID:  "cnt_7",
		FeatureKey: "api_calls",
		Fields:     "id,balance",
	})
	require.NoError(t, err)
	assert.Equal(t, "ent_123", got.ID)
}

func TestLookup_RequiresParams(t *testing.T) {
	t.Parallel()

	cl, _ := entitlements.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Lookup(context.Background(), nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_params", v.Code)
}

func TestLookup_RequiresContactID(t *testing.T) {
	t.Parallel()

	cl, _ := entitlements.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Lookup(context.Background(), &entitlements.LookupParams{
		ContactID:  "",
		FeatureKey: "api_calls",
	})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_contact_id", v.Code)
}

func TestLookup_RequiresFeatureKey(t *testing.T) {
	t.Parallel()

	cl, _ := entitlements.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Lookup(context.Background(), &entitlements.LookupParams{
		ContactID:  "cnt_7",
		FeatureKey: "",
	})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_feature_key", v.Code)
}

func TestLookup_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"no record"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Lookup(context.Background(), &entitlements.LookupParams{
		ContactID:  "cnt_7",
		FeatureKey: "unknown",
	})
	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
}

func TestGrant_SendsBodyAndDecodes(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/entitlements/grants", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got entitlements.GrantParams
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "cnt_7", got.ContactID)
		assert.Equal(t, "api_calls", got.FeatureKey)
		assert.Equal(t, int64(50), got.Amount)
		assert.Equal(t, "grant_2", got.GrantID)

		_, _ = io.WriteString(w, `{"data":{"id":"ent_123","balance":150}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Grant(context.Background(), &entitlements.GrantParams{
		ContactID:  "cnt_7",
		FeatureKey: "api_calls",
		Amount:     50,
		GrantID:    "grant_2",
		Metadata:   map[string]string{"reason": "comp"},
	})
	require.NoError(t, err)
	require.NotNil(t, got.Balance)
	assert.Equal(t, int64(150), *got.Balance)
}

func TestGrant_RequiresParams(t *testing.T) {
	t.Parallel()

	cl, _ := entitlements.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Grant(context.Background(), nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestConsume_SendsBodyAndDecodes(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/entitlements/consume", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got entitlements.ConsumeParams
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "cnt_7", got.ContactID)
		assert.Equal(t, int64(1), got.Amount)
		assert.Equal(t, "POST /generate", got.Reason)

		_, _ = io.WriteString(w, `{"data":{"id":"ent_123","balance":99}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Consume(context.Background(), &entitlements.ConsumeParams{
		ContactID:  "cnt_7",
		FeatureKey: "api_calls",
		Amount:     1,
		Reason:     "POST /generate",
	})
	require.NoError(t, err)
	require.NotNil(t, got.Balance)
	assert.Equal(t, int64(99), *got.Balance)
}

func TestConsume_409Conflict(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = io.WriteString(w, `{"error":{"code":"conflict","message":"insufficient balance"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Consume(context.Background(), &entitlements.ConsumeParams{
		ContactID:  "cnt_7",
		FeatureKey: "api_calls",
		Amount:     9999,
	})
	var conflict *threecommon.ConflictError
	require.True(t, errors.As(err, &conflict))
	assert.Equal(t, "conflict", conflict.Code)
}

func TestConsume_RequiresParams(t *testing.T) {
	t.Parallel()

	cl, _ := entitlements.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Consume(context.Background(), nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestListAutoPaginate_WalksEveryPage(t *testing.T) {
	t.Parallel()

	pages := []string{
		`{"data":[{"id":"ent_a"},{"id":"ent_b"}],"hasMore":true}`,
		`{"data":[{"id":"ent_c"}],"hasMore":false}`,
	}

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx, _ := strconv.Atoi(r.URL.Query().Get("page"))
		calls.Add(1)
		require.Less(t, idx, len(pages))
		assert.Equal(t, "api_calls", r.URL.Query().Get("featureKey"))
		_, _ = io.WriteString(w, pages[idx])
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	iter := cl.ListAutoPaginate(context.Background(), &entitlements.ListParams{FeatureKey: "api_calls"})

	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	require.NoError(t, iter.Err())
	assert.Equal(t, []string{"ent_a", "ent_b", "ent_c"}, ids)
	assert.Equal(t, int32(2), calls.Load())
}

func TestListAutoPaginate_HonorsExplicitStartPage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "5", r.URL.Query().Get("page"))
		_, _ = io.WriteString(w, `{"data":[{"id":"ent_z"}],"hasMore":false}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	page := 5
	iter := cl.ListAutoPaginate(context.Background(), &entitlements.ListParams{Page: &page})

	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	require.NoError(t, iter.Err())
	assert.Equal(t, []string{"ent_z"}, ids)
}
