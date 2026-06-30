package subscriptions_test

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
	"github.com/3-Common/sdk/sdk-go/resources/subscriptions"
)

// newTestClient returns a subscriptions.Client whose backend points at the
// supplied httptest server.
func newTestClient(t *testing.T, srv *httptest.Server) *subscriptions.Client {
	t.Helper()
	cl, err := subscriptions.New(threecommon.Config{
		APIKey:     "3co_test",
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		MaxRetries: threecommon.Int(0),
	})
	require.NoError(t, err)
	return cl
}

func TestList_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/subscriptions", r.URL.Path)
		assert.Equal(t, "active", r.URL.Query().Get("status"))
		_, _ = w.Write([]byte(`{"data":[{"id":"sub_a","status":"active"}],"hasMore":false}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.List(context.Background(), &subscriptions.ListParams{Status: subscriptions.StatusActive})
	require.NoError(t, err)
	require.Len(t, got.Data, 1)
	assert.Equal(t, "sub_a", got.Data[0].ID)
	assert.Equal(t, subscriptions.StatusActive, got.Data[0].Status)
}

func TestList_AllParamsEncoded(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.Equal(t, "0", q.Get("page"))
		assert.Equal(t, "5", q.Get("pageSize"))
		assert.Equal(t, "active", q.Get("status"))
		assert.Equal(t, "cnt_42", q.Get("contactId"))
		assert.Equal(t, "price_7", q.Get("priceId"))
		assert.Equal(t, "id,status", q.Get("fields"))
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	page := 0
	pageSize := 5
	_, err := cl.List(context.Background(), &subscriptions.ListParams{
		Page:      &page,
		PageSize:  &pageSize,
		Status:    subscriptions.StatusActive,
		ContactID: "cnt_42",
		PriceID:   "price_7",
		Fields:    "id,status",
	})
	require.NoError(t, err)
}

func TestList_EmptyParamsReturnNil(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.RawQuery, "empty ListParams must produce no query string")
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.List(context.Background(), &subscriptions.ListParams{})
	require.NoError(t, err)
}

func TestRetrieve_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/subscriptions/sub_123", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"id":"sub_123","status":"active"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Retrieve(context.Background(), "sub_123", nil)
	require.NoError(t, err)
	assert.Equal(t, "sub_123", got.ID)
}

func TestRetrieve_AppliesFieldsParam(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "id,status", r.URL.Query().Get("fields"))
		_, _ = w.Write([]byte(`{"data":{"id":"sub_1"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Retrieve(context.Background(), "sub_1", &subscriptions.RetrieveParams{Fields: "id,status"})
	require.NoError(t, err)
}

func TestRetrieve_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := subscriptions.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Retrieve(context.Background(), "", nil)
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
	_, err := cl.Retrieve(context.Background(), "sub_missing", nil)
	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
}

func TestCreate_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/subscriptions", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "cnt_42", got["contactId"])
		assert.Equal(t, "price_7", got["priceId"])

		_, _ = w.Write([]byte(`{"data":{"id":"sub_new","status":"trialing"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Create(context.Background(), &subscriptions.CreateParams{
		ContactID: "cnt_42",
		PriceID:   "price_7",
		Quantity:  threecommon.Int64(1),
		TrialDays: threecommon.Int(14),
	})
	require.NoError(t, err)
	assert.Equal(t, "sub_new", got.ID)
}

func TestCreate_RequiresParams(t *testing.T) {
	t.Parallel()
	cl, _ := subscriptions.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Create(context.Background(), nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestUpdate_SendsBodyAndReturnsProration(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "price_up", got["priceId"])

		_, _ = w.Write([]byte(`{
			"data":{"id":"sub_1","priceId":"price_up","quantity":2,"status":"active"},
			"invoice":{"id":"inv_p","status":"open","total":1234,"currency":"USD"},
			"proration":{"netAmountMinor":1234,"daysRemaining":7,"daysInCycle":31}
		}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Update(context.Background(), "sub_1", &subscriptions.UpdateParams{
		PriceID:  "price_up",
		Quantity: threecommon.Int64(2),
	})
	require.NoError(t, err)
	assert.Equal(t, "price_up", got.Subscription.PriceID)
	require.NotNil(t, got.Invoice)
	assert.Equal(t, "inv_p", got.Invoice.ID)
	assert.Equal(t, int64(1234), got.Proration.NetAmountMinor)
	assert.Equal(t, int64(7), got.Proration.DaysRemaining)
}

func TestUpdate_HandlesMissingInvoice(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"data":{"id":"sub_1","status":"active"},
			"proration":{"netAmountMinor":0,"daysRemaining":7,"daysInCycle":31}
		}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Update(context.Background(), "sub_1", &subscriptions.UpdateParams{Quantity: threecommon.Int64(1)})
	require.NoError(t, err)
	assert.Nil(t, got.Invoice)
	assert.Equal(t, int64(31), got.Proration.DaysInCycle)
}

func TestUpdate_ValidatesID(t *testing.T) {
	t.Parallel()
	cl, _ := subscriptions.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Update(context.Background(), "", &subscriptions.UpdateParams{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestUpdate_ValidatesParams(t *testing.T) {
	t.Parallel()
	cl, _ := subscriptions.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Update(context.Background(), "sub_1", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestRetrieveManageURL_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/subscriptions/sub_123/manage-url", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"url":"https://billing.3common.com/p/session/sub_123_a1b2c3"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.RetrieveManageURL(context.Background(), "sub_123")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "https://billing.3common.com/p/session/sub_123_a1b2c3", got.URL)
}

func TestRetrieveManageURL_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := subscriptions.New(threecommon.Config{APIKey: "k"})
	_, err := cl.RetrieveManageURL(context.Background(), "")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestRetrieveManageURL_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"missing"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.RetrieveManageURL(context.Background(), "sub_missing")
	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
}

func TestActivate_Posts(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/subscriptions/sub_1/activate", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"id":"sub_1","status":"active"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Activate(context.Background(), "sub_1")
	require.NoError(t, err)
	assert.Equal(t, subscriptions.StatusActive, got.Status)
}

func TestActivate_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := subscriptions.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Activate(context.Background(), "")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
}

func TestCancel_WithReason(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/subscriptions/sub_1/cancel", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "churn", got["reason"])
		_, _ = w.Write([]byte(`{"data":{"id":"sub_1","status":"active","cancelAtPeriodEnd":true}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Cancel(context.Background(), "sub_1", &subscriptions.CancelParams{Reason: "churn"})
	require.NoError(t, err)
	require.NotNil(t, got.CancelAtPeriodEnd)
	assert.True(t, *got.CancelAtPeriodEnd)
}

func TestCancel_NilParams(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.JSONEq(t, `{}`, string(body))
		_, _ = w.Write([]byte(`{"data":{"id":"sub_1","cancelAtPeriodEnd":true}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Cancel(context.Background(), "sub_1", nil)
	require.NoError(t, err)
}

func TestCancelImmediately_WithReason(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/subscriptions/sub_1/cancel-immediately", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "fraud", got["reason"])
		_, _ = w.Write([]byte(`{"data":{"id":"sub_1","status":"canceled","endedAt":"2026-05-25T00:00:00Z"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.CancelImmediately(context.Background(), "sub_1", &subscriptions.CancelImmediatelyParams{Reason: "fraud"})
	require.NoError(t, err)
	assert.Equal(t, subscriptions.StatusCanceled, got.Status)
	assert.Equal(t, "2026-05-25T00:00:00Z", got.EndedAt)
}

func TestMarkUnpaid_Posts(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/subscriptions/sub_1/mark-unpaid", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"id":"sub_1","status":"unpaid"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.MarkUnpaid(context.Background(), "sub_1")
	require.NoError(t, err)
	assert.Equal(t, subscriptions.StatusUnpaid, got.Status)
}

func TestCompNextCycle_Posts(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/subscriptions/sub_1/comp-next-cycle", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		assert.Empty(t, body, "comp-next-cycle takes no request body")
		_, _ = w.Write([]byte(`{"data":{"id":"sub_1","status":"active"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.CompNextCycle(context.Background(), "sub_1")
	require.NoError(t, err)
	assert.Equal(t, "sub_1", got.ID)
	assert.Equal(t, subscriptions.StatusActive, got.Status)
}

func TestCompNextCycle_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := subscriptions.New(threecommon.Config{APIKey: "k"})
	_, err := cl.CompNextCycle(context.Background(), "")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestCompNextCycle_409Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = io.WriteString(w, `{"error":{"code":"subscription_not_compable","message":"canceled and cannot have its next cycle comped"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.CompNextCycle(context.Background(), "sub_canceled")
	var c *threecommon.ConflictError
	require.True(t, errors.As(err, &c))
	assert.Equal(t, "subscription_not_compable", c.Code)
}

func TestUncompNextCycle_Posts(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/subscriptions/sub_1/uncomp-next-cycle", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		assert.Empty(t, body, "uncomp-next-cycle takes no request body")
		_, _ = w.Write([]byte(`{"data":{"id":"sub_1","status":"active"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.UncompNextCycle(context.Background(), "sub_1")
	require.NoError(t, err)
	assert.Equal(t, "sub_1", got.ID)
	assert.Equal(t, subscriptions.StatusActive, got.Status)
}

func TestUncompNextCycle_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := subscriptions.New(threecommon.Config{APIKey: "k"})
	_, err := cl.UncompNextCycle(context.Background(), "")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestUncompNextCycle_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"missing"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.UncompNextCycle(context.Background(), "sub_missing")
	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
}

func TestBill_ReturnsInvoice(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/subscriptions/sub_1/bill", r.URL.Path)
		_, _ = w.Write([]byte(`{
			"data":{"id":"sub_1","status":"active"},
			"invoice":{"id":"inv_9","status":"draft","total":50000,"currency":"USD"}
		}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Bill(context.Background(), "sub_1")
	require.NoError(t, err)
	assert.Equal(t, "sub_1", got.Subscription.ID)
	assert.Equal(t, "inv_9", got.Invoice.ID)
	assert.Equal(t, int64(50000), got.Invoice.Total)
}

func TestBill_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := subscriptions.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Bill(context.Background(), "")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
}

func TestRenew_WithInvoice(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"data":{"id":"sub_1","status":"active"},
			"invoice":{"id":"inv_r","status":"open","total":50000,"currency":"USD"}
		}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Renew(context.Background(), "sub_1")
	require.NoError(t, err)
	require.NotNil(t, got.Invoice)
	assert.Equal(t, "inv_r", got.Invoice.ID)
}

func TestRenew_WithoutInvoice(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"id":"sub_1","status":"canceled"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Renew(context.Background(), "sub_1")
	require.NoError(t, err)
	assert.Nil(t, got.Invoice)
	assert.Equal(t, subscriptions.StatusCanceled, got.Subscription.Status)
}

func TestPreviewUpcomingInvoice_Populated(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/subscriptions/sub_1/upcoming", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"invoice":{
			"customerId":"cnt_42","subscriptionId":"sub_1","currency":"USD",
			"lineItems":[{"description":"Pro","quantity":1,"unitAmount":5000}],
			"subtotal":5000,"total":5000,
			"periodStart":"2026-06-01T00:00:00Z","periodEnd":"2026-07-01T00:00:00Z"
		}}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.PreviewUpcomingInvoice(context.Background(), "sub_1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, int64(5000), got.Total)
	assert.Equal(t, "USD", got.Currency)
	require.Len(t, got.LineItems, 1)
	assert.Equal(t, int64(5000), got.LineItems[0].UnitAmount)
}

func TestPreviewUpcomingInvoice_Null(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"invoice":null}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.PreviewUpcomingInvoice(context.Background(), "sub_1")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPreviewUpcomingInvoice_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := subscriptions.New(threecommon.Config{APIKey: "k"})
	_, err := cl.PreviewUpcomingInvoice(context.Background(), "")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
}

func TestListAutoPaginate_WalksEveryPage(t *testing.T) {
	t.Parallel()

	pages := []string{
		`{"data":[{"id":"sub_1"},{"id":"sub_2"}],"hasMore":true}`,
		`{"data":[{"id":"sub_3"}],"hasMore":false}`,
	}

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx, _ := strconv.Atoi(r.URL.Query().Get("page"))
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
	assert.Equal(t, []string{"sub_1", "sub_2", "sub_3"}, ids)
	assert.Equal(t, int32(2), calls.Load())
}

func TestListAutoPaginate_HonorsExplicitStartPage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "5", r.URL.Query().Get("page"))
		_, _ = io.WriteString(w, `{"data":[{"id":"sub_5_a"}],"hasMore":false}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	startPage := 5
	iter := cl.ListAutoPaginate(context.Background(), &subscriptions.ListParams{Page: &startPage})

	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	require.NoError(t, iter.Err())
	assert.Equal(t, []string{"sub_5_a"}, ids)
}

func TestListAutoPaginate_ContextCancellationStopsIteration(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"data":[{"id":"sub_1"}],"hasMore":true}`)
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

func TestBill_409Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = io.WriteString(w, `{"error":{"code":"subscription_already_billed","message":"already billed"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Bill(context.Background(), "sub_1")
	var c *threecommon.ConflictError
	require.True(t, errors.As(err, &c))
}

func TestCreate_500SurfacesAsServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":{"code":"internal_error","message":"boom"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Create(context.Background(), &subscriptions.CreateParams{PriceID: "price_1"})
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
	_, err := cl.Update(context.Background(), "sub_1", &subscriptions.UpdateParams{Quantity: threecommon.Int64(1)})
	var server *threecommon.ServerError
	require.True(t, errors.As(err, &server))
}

func TestActivate_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"missing"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Activate(context.Background(), "sub_missing")
	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
}

func TestCancel_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := subscriptions.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Cancel(context.Background(), "", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
}

func TestCancel_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"missing"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Cancel(context.Background(), "sub_missing", &subscriptions.CancelParams{Reason: "x"})
	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
}

func TestCancelImmediately_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := subscriptions.New(threecommon.Config{APIKey: "k"})
	_, err := cl.CancelImmediately(context.Background(), "", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
}

func TestCancelImmediately_NilParams(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.JSONEq(t, `{}`, string(body))
		_, _ = w.Write([]byte(`{"data":{"id":"sub_1","status":"canceled"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.CancelImmediately(context.Background(), "sub_1", nil)
	require.NoError(t, err)
}

func TestCancelImmediately_500Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":{"code":"internal_error","message":"boom"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.CancelImmediately(context.Background(), "sub_1", nil)
	var server *threecommon.ServerError
	require.True(t, errors.As(err, &server))
}

func TestMarkUnpaid_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := subscriptions.New(threecommon.Config{APIKey: "k"})
	_, err := cl.MarkUnpaid(context.Background(), "")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
}

func TestMarkUnpaid_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"missing"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.MarkUnpaid(context.Background(), "sub_missing")
	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
}

func TestRenew_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := subscriptions.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Renew(context.Background(), "")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
}

func TestRenew_500Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":{"code":"internal_error","message":"boom"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Renew(context.Background(), "sub_1")
	var server *threecommon.ServerError
	require.True(t, errors.As(err, &server))
}

func TestPreviewUpcomingInvoice_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"missing"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.PreviewUpcomingInvoice(context.Background(), "sub_missing")
	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
}

func TestNew_RequiresAPIKey(t *testing.T) {
	t.Setenv("THREECOMMON_API_KEY", "")
	_, err := subscriptions.New(threecommon.Config{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_api_key", v.Code)
}

func TestFromBackend_InternalConstructorUsable(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	backend, err := core.NewFromConfig(threecommon.Config{
		APIKey:     "3co_test",
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		MaxRetries: threecommon.Int(0),
	})
	require.NoError(t, err)

	cl := subscriptions.FromBackend(backend)
	require.NotNil(t, cl)
	_, err = cl.List(context.Background(), nil)
	require.NoError(t, err)
}
