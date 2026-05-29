package invoices_test

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
	"github.com/3-Common/sdk/sdk-go/resources/invoices"
)

// newTestClient returns an invoices.Client whose backend points at the
// supplied httptest server.
func newTestClient(t *testing.T, srv *httptest.Server) *invoices.Client {
	t.Helper()
	cl, err := invoices.New(threecommon.Config{
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
		assert.Equal(t, "/v1/invoices", r.URL.Path)
		assert.Equal(t, "open", r.URL.Query().Get("status"))
		_, _ = w.Write([]byte(`{"data":[{"id":"inv_a","status":"open"}],"hasMore":false}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.List(context.Background(), &invoices.ListParams{Status: invoices.StatusOpen})
	require.NoError(t, err)
	require.Len(t, got.Data, 1)
	assert.Equal(t, "inv_a", got.Data[0].ID)
	assert.Equal(t, invoices.StatusOpen, got.Data[0].Status)
}

func TestList_AllParamsEncoded(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.Equal(t, "0", q.Get("page"))
		assert.Equal(t, "5", q.Get("pageSize"))
		assert.Equal(t, "open", q.Get("status"))
		assert.Equal(t, "cnt_42", q.Get("customerId"))
		assert.Equal(t, "2026-01-01", q.Get("issuedAfter"))
		assert.Equal(t, "2026-12-31", q.Get("issuedBefore"))
		assert.Equal(t, "id,status", q.Get("fields"))
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	page := 0
	pageSize := 5
	_, err := cl.List(context.Background(), &invoices.ListParams{
		Page:         &page,
		PageSize:     &pageSize,
		Status:       invoices.StatusOpen,
		CustomerID:   "cnt_42",
		IssuedAfter:  "2026-01-01",
		IssuedBefore: "2026-12-31",
		Fields:       "id,status",
	})
	require.NoError(t, err)
}

func TestRetrieve_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/invoices/inv_123", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"id":"inv_123","status":"draft"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Retrieve(context.Background(), "inv_123", nil)
	require.NoError(t, err)
	assert.Equal(t, "inv_123", got.ID)
}

func TestRetrieve_AppliesFieldsParam(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "id,status", r.URL.Query().Get("fields"))
		_, _ = w.Write([]byte(`{"data":{"id":"inv_1"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Retrieve(context.Background(), "inv_1", &invoices.RetrieveParams{Fields: "id,status"})
	require.NoError(t, err)
}

func TestRetrieve_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := invoices.New(threecommon.Config{APIKey: "k"})
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
	_, err := cl.Retrieve(context.Background(), "inv_missing", nil)
	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
}

func TestCreate_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/invoices", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "cnt_42", got["customerId"])
		assert.Equal(t, "USD", got["currency"])

		_, _ = w.Write([]byte(`{"data":{"id":"inv_new","status":"draft"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Create(context.Background(), &invoices.CreateParams{
		CustomerID: "cnt_42",
		Currency:   invoices.CurrencyUSD,
		LineItems: []invoices.LineItem{
			{Description: "Consulting", Quantity: 1, UnitAmount: 50_000},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "inv_new", got.ID)
}

func TestCreate_RequiresParams(t *testing.T) {
	t.Parallel()
	cl, _ := invoices.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Create(context.Background(), nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestUpdate_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "Net 30", got["notes"])

		_, _ = w.Write([]byte(`{"data":{"id":"inv_1","notes":"Net 30"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Update(context.Background(), "inv_1", &invoices.UpdateParams{Notes: "Net 30"})
	require.NoError(t, err)
	assert.Equal(t, "Net 30", got.Notes)
}

func TestUpdate_ValidatesID(t *testing.T) {
	t.Parallel()
	cl, _ := invoices.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Update(context.Background(), "", &invoices.UpdateParams{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestUpdate_ValidatesParams(t *testing.T) {
	t.Parallel()
	cl, _ := invoices.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Update(context.Background(), "inv_1", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestFinalize_Posts(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/invoices/inv_1/finalize", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"id":"inv_1","status":"open","number":"INV-0001"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Finalize(context.Background(), "inv_1")
	require.NoError(t, err)
	assert.Equal(t, invoices.StatusOpen, got.Status)
	require.NotNil(t, got.Number)
	assert.Equal(t, "INV-0001", *got.Number)
}

func TestFinalize_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := invoices.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Finalize(context.Background(), "")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
}

func TestVoid_WithReason(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/invoices/inv_1/void", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "Sent in error", got["reason"])

		_, _ = w.Write([]byte(`{"data":{"id":"inv_1","status":"void"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Void(context.Background(), "inv_1", &invoices.VoidParams{Reason: "Sent in error"})
	require.NoError(t, err)
	assert.Equal(t, invoices.StatusVoid, got.Status)
}

func TestVoid_NilParams(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.JSONEq(t, `{}`, string(body))
		_, _ = w.Write([]byte(`{"data":{"id":"inv_1","status":"void"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Void(context.Background(), "inv_1", nil)
	require.NoError(t, err)
}

func TestRecordPayment_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/invoices/inv_1/payments", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var got struct {
			Payment        int64  `json:"payment"`
			IdempotencyKey string `json:"idempotencyKey"`
		}
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, int64(50_000), got.Payment)
		assert.Equal(t, "pmt-1", got.IdempotencyKey)

		_, _ = w.Write([]byte(`{"data":{"id":"inv_1","status":"paid","amountDue":0}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.RecordPayment(context.Background(), "inv_1", &invoices.PaymentParams{
		Payment:        50_000,
		IdempotencyKey: "pmt-1",
	})
	require.NoError(t, err)
	assert.Equal(t, invoices.StatusPaid, got.Status)
}

func TestRecordPayment_RequiresParams(t *testing.T) {
	t.Parallel()
	cl, _ := invoices.New(threecommon.Config{APIKey: "k"})
	_, err := cl.RecordPayment(context.Background(), "inv_1", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestListAutoPaginate_WalksEveryPage(t *testing.T) {
	t.Parallel()

	pages := []string{
		`{"data":[{"id":"inv_1"},{"id":"inv_2"}],"hasMore":true}`,
		`{"data":[{"id":"inv_3"}],"hasMore":false}`,
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
	assert.Equal(t, []string{"inv_1", "inv_2", "inv_3"}, ids)
	assert.Equal(t, int32(2), calls.Load())
}

func TestNew_RequiresAPIKey(t *testing.T) {
	t.Setenv("THREECOMMON_API_KEY", "")
	_, err := invoices.New(threecommon.Config{})
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

	cl := invoices.FromBackend(backend)
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

func TestCreate_500SurfacesAsServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":{"code":"internal_error","message":"boom"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Create(context.Background(), &invoices.CreateParams{
		CustomerID: "cnt_42",
		Currency:   invoices.CurrencyUSD,
		LineItems:  []invoices.LineItem{{Description: "x", Quantity: 1, UnitAmount: 1}},
	})

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
	_, err := cl.Update(context.Background(), "inv_1", &invoices.UpdateParams{Notes: "x"})

	var server *threecommon.ServerError
	require.True(t, errors.As(err, &server))
}

func TestFinalize_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"missing"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Finalize(context.Background(), "inv_missing")

	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
}

func TestVoid_NilBodySendsEmptyJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.JSONEq(t, `{}`, string(body))
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":{"code":"internal_error","message":"boom"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Void(context.Background(), "inv_1", nil)

	var server *threecommon.ServerError
	require.True(t, errors.As(err, &server))
}

func TestRecordPayment_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"missing"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.RecordPayment(context.Background(), "inv_missing", &invoices.PaymentParams{Payment: 1})

	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
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

func TestListAutoPaginate_HonorsExplicitStartPage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "5", r.URL.Query().Get("page"))
		_, _ = io.WriteString(w, `{"data":[{"id":"inv_5_a"}],"hasMore":false}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	startPage := 5
	iter := cl.ListAutoPaginate(context.Background(), &invoices.ListParams{Page: &startPage})

	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	require.NoError(t, iter.Err())
	assert.Equal(t, []string{"inv_5_a"}, ids)
}

func TestListAutoPaginate_ContextCancellationStopsIteration(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"data":[{"id":"inv_1"}],"hasMore":true}`)
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

func TestList_EmptyParamsReturnNil(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.RawQuery, "empty ListParams must produce no query string")
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.List(context.Background(), &invoices.ListParams{})
	require.NoError(t, err)
}

func TestList_ForwardsSubscriptionID(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "sub_99", r.URL.Query().Get("subscriptionId"))
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.List(context.Background(), &invoices.ListParams{SubscriptionID: "sub_99"})
	require.NoError(t, err)
}

func TestAutoCharge_Paid(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/invoices/inv_1/auto_charge", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"id":"inv_1","status":"paid","amountDue":0},"outcome":"paid"}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.AutoCharge(context.Background(), "inv_1")
	require.NoError(t, err)
	assert.Equal(t, invoices.AutoChargeOutcomePaid, got.Outcome)
	assert.Equal(t, invoices.StatusPaid, got.Invoice.Status)
	assert.Empty(t, got.FailureCode)
}

func TestAutoCharge_Declined(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"id":"inv_1","status":"payment_failed"},"outcome":"failed","failureCode":"card_declined"}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.AutoCharge(context.Background(), "inv_1")
	require.NoError(t, err)
	assert.Equal(t, invoices.AutoChargeOutcomeFailed, got.Outcome)
	assert.Equal(t, invoices.StatusPaymentFailed, got.Invoice.Status)
	assert.Equal(t, "card_declined", got.FailureCode)
}

func TestAutoCharge_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := invoices.New(threecommon.Config{APIKey: "k"})
	_, err := cl.AutoCharge(context.Background(), "")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestRefundPayment_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/invoices/inv_1/payments/pay_9/refunds", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, float64(25_000), got["amount"])
		assert.Equal(t, "requested_by_customer", got["reason"])

		_, _ = w.Write([]byte(`{"data":{"id":"inv_1","status":"paid"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.RefundPayment(context.Background(), "inv_1", "pay_9", &invoices.RefundParams{
		Amount:         25_000,
		Reason:         "requested_by_customer",
		IdempotencyKey: "rfnd-1",
	})
	require.NoError(t, err)
	assert.Equal(t, invoices.StatusPaid, got.Status)
}

func TestRefundPayment_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := invoices.New(threecommon.Config{APIKey: "k"})
	_, err := cl.RefundPayment(context.Background(), "", "pay_9", &invoices.RefundParams{Amount: 1})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestRefundPayment_RequiresPaymentID(t *testing.T) {
	t.Parallel()
	cl, _ := invoices.New(threecommon.Config{APIKey: "k"})
	_, err := cl.RefundPayment(context.Background(), "inv_1", "", &invoices.RefundParams{Amount: 1})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_payment_id", v.Code)
}

func TestRefundPayment_RequiresParams(t *testing.T) {
	t.Parallel()
	cl, _ := invoices.New(threecommon.Config{APIKey: "k"})
	_, err := cl.RefundPayment(context.Background(), "inv_1", "pay_9", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestDeleteDraft_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v1/invoices/inv_1", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"id":"inv_1"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.DeleteDraft(context.Background(), "inv_1")
	require.NoError(t, err)
	assert.Equal(t, "inv_1", got.ID)
}

func TestDeleteDraft_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := invoices.New(threecommon.Config{APIKey: "k"})
	_, err := cl.DeleteDraft(context.Background(), "")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestAutoCharge_502SurfacesAsServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, `{"error":{"code":"processor_error","message":"upstream"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.AutoCharge(context.Background(), "inv_1")

	var server *threecommon.ServerError
	require.True(t, errors.As(err, &server))
}

func TestRefundPayment_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"missing"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.RefundPayment(context.Background(), "inv_missing", "pay_9", &invoices.RefundParams{Amount: 1})

	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
}

func TestDeleteDraft_409Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = io.WriteString(w, `{"error":{"code":"invoice_not_draft","message":"finalized"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.DeleteDraft(context.Background(), "inv_open")

	var conflict *threecommon.ConflictError
	require.True(t, errors.As(err, &conflict))
}
