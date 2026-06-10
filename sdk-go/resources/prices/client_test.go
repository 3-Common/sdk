package prices_test

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
	"github.com/3-Common/sdk/sdk-go/resources/prices"
)

const samplePrice = `{
	"id": "price_123",
	"hostId": "host_1",
	"productId": "prod_7",
	"type": "recurring",
	"currency": "USD",
	"unitAmount": 1500,
	"recurring": {"interval": "month", "intervalCount": 1},
	"features": [
		{"featureKey": "api_calls", "type": "quantity", "quantity": 1000, "rolloverEnabled": false}
	],
	"nickname": "Pro monthly",
	"active": true,
	"createdAt": "2026-05-01T00:00:00.000Z",
	"updatedAt": "2026-05-01T00:00:00.000Z"
}`

func newTestClient(t *testing.T, srv *httptest.Server) *prices.Client {
	t.Helper()
	cl, err := prices.New(threecommon.Config{
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
	_, err := prices.New(threecommon.Config{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_api_key", v.Code)
}

func TestList_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/prices", r.URL.Path)
		assert.Equal(t, "prod_7", r.URL.Query().Get("productId"))
		assert.Equal(t, "true", r.URL.Query().Get("active"))
		_, _ = io.WriteString(w, `{"data":[`+samplePrice+`],"hasMore":false}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.List(context.Background(), &prices.ListParams{
		ProductID: "prod_7",
		Active:    threecommon.Bool(true),
	})
	require.NoError(t, err)
	require.Len(t, got.Data, 1)
	assert.Equal(t, "price_123", got.Data[0].ID)
	require.NotNil(t, got.Data[0].Recurring)
	assert.Equal(t, prices.IntervalMonth, got.Data[0].Recurring.Interval)
	require.Len(t, got.Data[0].Features, 1)
	assert.Equal(t, prices.FeatureTypeQuantity, got.Data[0].Features[0].Type)
	require.NotNil(t, got.Data[0].Features[0].Quantity)
	assert.Equal(t, int64(1000), *got.Data[0].Features[0].Quantity)
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
		assert.Equal(t, "prod_7", q.Get("productId"))
		assert.Equal(t, "recurring", q.Get("type"))
		assert.Equal(t, "false", q.Get("active")) // booleans render lowercase
		assert.Equal(t, "id,unitAmount", q.Get("fields"))
		_, _ = io.WriteString(w, `{"data":[],"hasMore":false}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	page := 0
	pageSize := 25
	_, err := cl.List(context.Background(), &prices.ListParams{
		Page:      &page,
		PageSize:  &pageSize,
		ProductID: "prod_7",
		Type:      prices.TypeRecurring,
		Active:    threecommon.Bool(false),
		Fields:    "id,unitAmount",
	})
	require.NoError(t, err)
}

func TestRetrieve_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/prices/price_123", r.URL.Path)
		_, _ = io.WriteString(w, `{"data":`+samplePrice+`}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Retrieve(context.Background(), "price_123", nil)
	require.NoError(t, err)
	assert.Equal(t, "price_123", got.ID)
	require.NotNil(t, got.UnitAmount)
	assert.Equal(t, int64(1500), *got.UnitAmount)
}

func TestRetrieve_AppliesFieldsParam(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "id,unitAmount", r.URL.Query().Get("fields"))
		_, _ = io.WriteString(w, `{"data":{"id":"price_1","unitAmount":5}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Retrieve(context.Background(), "price_1", &prices.RetrieveParams{Fields: "id,unitAmount"})
	require.NoError(t, err)
}

func TestRetrieve_RequiresID(t *testing.T) {
	t.Parallel()

	cl, _ := prices.New(threecommon.Config{APIKey: "k"})
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
	_, err := cl.Retrieve(context.Background(), "price_missing", nil)

	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
	assert.Equal(t, "req-404", nf.RequestID)
}

func TestCreate_SendsFeatureBodyAndDecodes(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/prices", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		require.NoError(t, json.Unmarshal(body, &m))
		assert.Equal(t, "prod_7", m["productId"])
		assert.InEpsilon(t, float64(1500), m["unitAmount"], 0.0001)

		feats, ok := m["features"].([]any)
		require.True(t, ok)
		require.Len(t, feats, 1)
		f0, ok := feats[0].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "api_calls", f0["featureKey"])
		assert.Equal(t, "quantity", f0["type"])
		assert.InEpsilon(t, float64(1000), f0["quantity"], 0.0001)
		assert.Equal(t, false, f0["rolloverEnabled"])
		// Fields from other variants must not leak onto a quantity feature.
		assert.NotContains(t, f0, "enabled")
		assert.NotContains(t, f0, "enumValue")
		assert.NotContains(t, f0, "durationDays")

		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"data":`+samplePrice+`}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Create(context.Background(), &prices.CreateParams{
		ProductID:  "prod_7",
		Type:       prices.TypeRecurring,
		Currency:   prices.CurrencyUSD,
		UnitAmount: 1500,
		Recurring:  &prices.Recurring{Interval: prices.IntervalMonth, IntervalCount: 1},
		Features: []prices.Feature{
			{
				FeatureKey:      "api_calls",
				Type:            prices.FeatureTypeQuantity,
				Quantity:        threecommon.Int64(1000),
				RolloverEnabled: threecommon.Bool(false),
			},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "price_123", got.ID)
}

func TestCreate_UnlimitedQuantitySerializesNull(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		require.NoError(t, json.Unmarshal(body, &m))
		feats := m["features"].([]any)
		f0 := feats[0].(map[string]any)
		// quantity must be present and explicitly null (= unlimited), not omitted.
		q, ok := f0["quantity"]
		assert.True(t, ok, "quantity key must be present")
		assert.Nil(t, q, "quantity must serialize as null")

		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"data":`+samplePrice+`}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Create(context.Background(), &prices.CreateParams{
		ProductID:  "prod_7",
		Type:       prices.TypeRecurring,
		Currency:   prices.CurrencyUSD,
		UnitAmount: 0,
		Recurring:  &prices.Recurring{Interval: prices.IntervalMonth, IntervalCount: 1},
		Features: []prices.Feature{
			{
				FeatureKey:      "seats",
				Type:            prices.FeatureTypeQuantity,
				Quantity:        nil, // unlimited
				RolloverEnabled: threecommon.Bool(true),
			},
		},
	})
	require.NoError(t, err)
}

func TestCreate_RequiresParams(t *testing.T) {
	t.Parallel()

	cl, _ := prices.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Create(context.Background(), nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestCreate_400Validation(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":{"code":"validation_error","message":"recurring required"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Create(context.Background(), &prices.CreateParams{
		ProductID:  "prod_7",
		Type:       prices.TypeRecurring,
		Currency:   prices.CurrencyUSD,
		UnitAmount: 1500,
	})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "validation_error", v.Code)
}

func TestUpdate_SendsBodyAndDecodes(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/v1/prices/price_123", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		require.NoError(t, json.Unmarshal(body, &m))
		assert.InEpsilon(t, float64(1200), m["unitAmount"], 0.0001)

		_, _ = io.WriteString(w, `{"data":{"id":"price_123","unitAmount":1200}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Update(context.Background(), "price_123", &prices.UpdateParams{
		UnitAmount: threecommon.Int64(1200),
		Nickname:   threecommon.String("Pro monthly (promo)"),
	})
	require.NoError(t, err)
	require.NotNil(t, got.UnitAmount)
	assert.Equal(t, int64(1200), *got.UnitAmount)
}

func TestUpdate_RequiresID(t *testing.T) {
	t.Parallel()

	cl, _ := prices.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Update(context.Background(), "", &prices.UpdateParams{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestUpdate_RequiresParams(t *testing.T) {
	t.Parallel()

	cl, _ := prices.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Update(context.Background(), "price_1", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestArchive_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/prices/price_123/archive", r.URL.Path)
		_, _ = io.WriteString(w, `{"data":{"id":"price_123","active":false}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Archive(context.Background(), "price_123")
	require.NoError(t, err)
	require.NotNil(t, got.Active)
	assert.False(t, *got.Active)
}

func TestUnarchive_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/prices/price_123/unarchive", r.URL.Path)
		_, _ = io.WriteString(w, `{"data":{"id":"price_123","active":true}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Unarchive(context.Background(), "price_123")
	require.NoError(t, err)
	require.NotNil(t, got.Active)
	assert.True(t, *got.Active)
}

func TestArchive_RequiresID(t *testing.T) {
	t.Parallel()

	cl, _ := prices.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Archive(context.Background(), "")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestFeature_MarshalJSON(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		feature prices.Feature
		want    map[string]any
	}{
		{
			name:    "boolean",
			feature: prices.Feature{FeatureKey: "beta", Type: prices.FeatureTypeBoolean, Enabled: threecommon.Bool(true)},
			want:    map[string]any{"featureKey": "beta", "type": "boolean", "enabled": true},
		},
		{
			name: "quantity unlimited",
			feature: prices.Feature{
				FeatureKey: "seats", Type: prices.FeatureTypeQuantity, RolloverEnabled: threecommon.Bool(false),
			},
			want: map[string]any{
				"featureKey": "seats", "type": "quantity", "quantity": nil, "rolloverEnabled": false,
			},
		},
		{
			name:    "enum",
			feature: prices.Feature{FeatureKey: "tier", Type: prices.FeatureTypeEnum, EnumValue: "gold"},
			want:    map[string]any{"featureKey": "tier", "type": "enum", "enumValue": "gold"},
		},
		{
			name:    "duration",
			feature: prices.Feature{FeatureKey: "trial", Type: prices.FeatureTypeDuration, DurationDays: threecommon.Int64(30)},
			want:    map[string]any{"featureKey": "trial", "type": "duration", "durationDays": float64(30)},
		},
		{
			name:    "unknown type falls back to struct",
			feature: prices.Feature{FeatureKey: "x", Type: prices.FeatureType("future")},
			want:    map[string]any{"featureKey": "x", "type": "future"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			raw, err := json.Marshal(tc.feature)
			require.NoError(t, err)
			var got map[string]any
			require.NoError(t, json.Unmarshal(raw, &got))
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestListAutoPaginate_WalksEveryPage(t *testing.T) {
	t.Parallel()

	pages := []string{
		`{"data":[{"id":"price_1"},{"id":"price_2"}],"hasMore":true}`,
		`{"data":[{"id":"price_3"}],"hasMore":false}`,
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
	iter := cl.ListAutoPaginate(context.Background(), &prices.ListParams{Active: threecommon.Bool(true)})

	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	require.NoError(t, iter.Err())
	assert.Equal(t, []string{"price_1", "price_2", "price_3"}, ids)
	assert.Equal(t, int32(2), calls.Load())
}
