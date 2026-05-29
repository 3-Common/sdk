package invoices

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/internal/core"
	"github.com/3-Common/sdk/sdk-go/pagination"
)

// Client is the invoices resource client. Construct one via [New] for
// standalone use, or use [github.com/3-Common/sdk/sdk-go/client.API.Invoices]
// when you need multiple resources sharing a single backend.
type Client struct {
	backend *core.Client
}

// New constructs a [*Client] from a [threecommon.Config].
func New(cfg threecommon.Config) (*Client, error) {
	backend, err := core.NewFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{backend: backend}, nil
}

// FromBackend wraps an existing backend. Used by the
// [github.com/3-Common/sdk/sdk-go/client] aggregator package.
func FromBackend(b *core.Client) *Client {
	return &Client{backend: b}
}

// List fetches a single page of invoices. To iterate every invoice matching a
// filter, use [Client.ListAutoPaginate].
func (c *Client) List(ctx context.Context, params *ListParams) (*ListResponse, error) {
	var out ListResponse
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/invoices",
		Query:  encodeListParams(params),
		Out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Retrieve fetches a single invoice by ID.
func (c *Client) Retrieve(ctx context.Context, id string, params *RetrieveParams) (*Invoice, error) {
	if err := requireID("Retrieve", id); err != nil {
		return nil, err
	}

	var query map[string]string
	if params != nil && params.Fields != "" {
		query = map[string]string{"fields": params.Fields}
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/invoices/" + url.PathEscape(id),
		Query:  query,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Create makes a new draft invoice. Totals are computed server-side from
// line items.
func (c *Client) Create(ctx context.Context, params *CreateParams) (*Invoice, error) {
	if params == nil {
		return nil, missingBody("Create")
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/invoices",
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Update revises a draft invoice. Only legal while in draft.
func (c *Client) Update(ctx context.Context, id string, params *UpdateParams) (*Invoice, error) {
	if err := requireID("Update", id); err != nil {
		return nil, err
	}
	if params == nil {
		return nil, missingBody("Update")
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPatch,
		Path:   "/invoices/" + url.PathEscape(id),
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Finalize transitions a draft invoice to open, assigns a sequential number,
// and stamps issuedAt.
func (c *Client) Finalize(ctx context.Context, id string) (*Invoice, error) {
	if err := requireID("Finalize", id); err != nil {
		return nil, err
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/invoices/" + url.PathEscape(id) + "/finalize",
		Body:   struct{}{},
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Void voids an invoice. Permitted from draft or open. Paid invoices cannot
// be voided. Pass nil for params to void without a reason.
func (c *Client) Void(ctx context.Context, id string, params *VoidParams) (*Invoice, error) {
	if err := requireID("Void", id); err != nil {
		return nil, err
	}

	var body any = struct{}{}
	if params != nil {
		body = params
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/invoices/" + url.PathEscape(id) + "/void",
		Body:   body,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// RecordPayment records a manual payment against an open invoice. Cumulative
// payments meeting the total transition the invoice to paid.
func (c *Client) RecordPayment(ctx context.Context, id string, params *PaymentParams) (*Invoice, error) {
	if err := requireID("RecordPayment", id); err != nil {
		return nil, err
	}
	if params == nil {
		return nil, missingBody("RecordPayment")
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/invoices/" + url.PathEscape(id) + "/payments",
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// AutoCharge off-session charges the customer's saved card for an open
// invoice. A decline is not an error — it returns a result with Outcome
// "failed" and a FailureCode, leaving the invoice in payment_failed. Only
// network / processor 5xx errors return an error.
func (c *Client) AutoCharge(ctx context.Context, id string) (*AutoChargeResult, error) {
	if err := requireID("AutoCharge", id); err != nil {
		return nil, err
	}

	var env autoChargeEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/invoices/" + url.PathEscape(id) + "/auto_charge",
		Body:   struct{}{},
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &AutoChargeResult{Invoice: env.Data, Outcome: env.Outcome, FailureCode: env.FailureCode}, nil
}

// RefundPayment refunds all or part of a recorded payment on a paid invoice.
// It is idempotent on params.IdempotencyKey: replays return the existing refund
// without contacting the processor again.
func (c *Client) RefundPayment(ctx context.Context, id, paymentID string, params *RefundParams) (*Invoice, error) {
	if err := requireID("RefundPayment", id); err != nil {
		return nil, err
	}
	if paymentID == "" {
		return nil, &threecommon.ValidationError{APIError: &threecommon.APIError{
			Code:    "missing_id",
			Message: "invoices.RefundPayment: paymentID must be non-empty",
		}}
	}
	if params == nil {
		return nil, missingBody("RefundPayment")
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/invoices/" + url.PathEscape(id) + "/payments/" + url.PathEscape(paymentID) + "/refunds",
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// DeleteDraft permanently deletes a draft invoice. Only legal while in draft
// (no number issued); finalized invoices must be voided instead so the audit
// trail stays intact.
func (c *Client) DeleteDraft(ctx context.Context, id string) (*DeleteDraftResult, error) {
	if err := requireID("DeleteDraft", id); err != nil {
		return nil, err
	}

	var env deletedEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodDelete,
		Path:   "/invoices/" + url.PathEscape(id),
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// ListAutoPaginate returns a [*pagination.Iter] that walks every invoice
// matching params. Pages are fetched lazily.
//
// On Go 1.23+, the iterator also supports range-over-func via [pagination.Iter.All].
func (c *Client) ListAutoPaginate(ctx context.Context, params *ListParams) *pagination.Iter[Invoice] {
	startPage := 0
	if params != nil && params.Page != nil {
		startPage = *params.Page
	}

	return pagination.NewIter(startPage, func(page int) ([]Invoice, bool, error) {
		pageParams := ListParams{}
		if params != nil {
			pageParams = *params
		}
		pageParams.Page = &page

		var out ListResponse
		if err := c.backend.Do(ctx, core.Request{
			Method: http.MethodGet,
			Path:   "/invoices",
			Query:  encodeListParams(&pageParams),
			Out:    &out,
		}); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil, false, err
			}
			return nil, false, err
		}
		return out.Data, out.HasMore, nil
	})
}

func requireID(method, id string) error {
	if id == "" {
		return &threecommon.ValidationError{APIError: &threecommon.APIError{
			Code:    "missing_id",
			Message: "invoices." + method + ": id must be non-empty",
		}}
	}
	return nil
}

func missingBody(method string) error {
	return &threecommon.ValidationError{APIError: &threecommon.APIError{
		Code:    "missing_body",
		Message: "invoices." + method + ": params must be non-nil",
	}}
}

func encodeListParams(p *ListParams) map[string]string {
	if p == nil {
		return nil
	}
	q := map[string]string{}
	if p.Page != nil {
		q["page"] = strconv.Itoa(*p.Page)
	}
	if p.PageSize != nil {
		q["pageSize"] = strconv.Itoa(*p.PageSize)
	}
	if p.Status != "" {
		q["status"] = string(p.Status)
	}
	if p.CustomerID != "" {
		q["customerId"] = p.CustomerID
	}
	if p.SubscriptionID != "" {
		q["subscriptionId"] = p.SubscriptionID
	}
	if p.IssuedAfter != "" {
		q["issuedAfter"] = p.IssuedAfter
	}
	if p.IssuedBefore != "" {
		q["issuedBefore"] = p.IssuedBefore
	}
	if p.Fields != "" {
		q["fields"] = p.Fields
	}
	if len(q) == 0 {
		return nil
	}
	return q
}
