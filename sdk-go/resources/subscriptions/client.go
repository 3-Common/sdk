package subscriptions

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

// Client is the subscriptions resource client. Construct one via [New] for
// standalone use, or use [github.com/3-Common/sdk/sdk-go/client.API.Subscriptions]
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

// List fetches a single page of subscriptions. To iterate every subscription
// matching a filter, use [Client.ListAutoPaginate].
func (c *Client) List(ctx context.Context, params *ListParams) (*ListResponse, error) {
	var out ListResponse
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/subscriptions",
		Query:  encodeListParams(params),
		Out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Retrieve fetches a single subscription by ID.
func (c *Client) Retrieve(ctx context.Context, id string, params *RetrieveParams) (*Subscription, error) {
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
		Path:   "/subscriptions/" + url.PathEscape(id),
		Query:  query,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Create makes a new subscription against an active recurring Price. Starts
// in trialing if TrialDays is set, else incomplete (awaiting first payment).
func (c *Client) Create(ctx context.Context, params *CreateParams) (*Subscription, error) {
	if params == nil {
		return nil, missingBody("Create")
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/subscriptions",
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Update applies a mid-cycle price/quantity change with Stripe-style daily
// proration, or flips forward-looking settings (notes, taxIds, taxRate,
// autoCharge, dunningEnabled, paymentDueDays).
func (c *Client) Update(ctx context.Context, id string, params *UpdateParams) (*UpdateResult, error) {
	if err := requireID("Update", id); err != nil {
		return nil, err
	}
	if params == nil {
		return nil, missingBody("Update")
	}

	var env updateEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPatch,
		Path:   "/subscriptions/" + url.PathEscape(id),
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &UpdateResult{
		Subscription: env.Data,
		Invoice:      env.Invoice,
		Proration:    env.Proration,
	}, nil
}

// Activate transitions an incomplete or trialing subscription to active.
func (c *Client) Activate(ctx context.Context, id string) (*Subscription, error) {
	if err := requireID("Activate", id); err != nil {
		return nil, err
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/subscriptions/" + url.PathEscape(id) + "/activate",
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Cancel schedules cancellation at the end of the current period. Idempotent.
// Pass nil for params to cancel without a reason.
func (c *Client) Cancel(ctx context.Context, id string, params *CancelParams) (*Subscription, error) {
	if err := requireID("Cancel", id); err != nil {
		return nil, err
	}

	var body any = struct{}{}
	if params != nil {
		body = params
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/subscriptions/" + url.PathEscape(id) + "/cancel",
		Body:   body,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// CancelImmediately is an admin override that terminates the subscription
// right now (status canceled, endedAt stamped). Pass nil for params to skip
// providing a reason.
func (c *Client) CancelImmediately(ctx context.Context, id string, params *CancelImmediatelyParams) (*Subscription, error) {
	if err := requireID("CancelImmediately", id); err != nil {
		return nil, err
	}

	var body any = struct{}{}
	if params != nil {
		body = params
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/subscriptions/" + url.PathEscape(id) + "/cancel-immediately",
		Body:   body,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// MarkUnpaid is an admin override that marks a subscription unpaid (terminal),
// bypassing dunning retries.
func (c *Client) MarkUnpaid(ctx context.Context, id string) (*Subscription, error) {
	if err := requireID("MarkUnpaid", id); err != nil {
		return nil, err
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/subscriptions/" + url.PathEscape(id) + "/mark-unpaid",
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Bill generates a draft invoice for the subscription's current period
// without advancing the period.
func (c *Client) Bill(ctx context.Context, id string) (*BillResult, error) {
	if err := requireID("Bill", id); err != nil {
		return nil, err
	}

	var env billEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/subscriptions/" + url.PathEscape(id) + "/bill",
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &BillResult{Subscription: env.Data, Invoice: env.Invoice}, nil
}

// Renew advances the subscription to its next billing period and generates an
// invoice. Transitions to canceled instead when CancelAtPeriodEnd was set.
func (c *Client) Renew(ctx context.Context, id string) (*RenewResult, error) {
	if err := requireID("Renew", id); err != nil {
		return nil, err
	}

	var env renewEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/subscriptions/" + url.PathEscape(id) + "/renew",
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &RenewResult{Subscription: env.Data, Invoice: env.Invoice}, nil
}

// PreviewUpcomingInvoice returns a non-persisted preview of the invoice the
// next renewal will generate. Returns nil when the subscription is set to
// cancel at period end.
func (c *Client) PreviewUpcomingInvoice(ctx context.Context, id string) (*InvoicePreview, error) {
	if err := requireID("PreviewUpcomingInvoice", id); err != nil {
		return nil, err
	}

	var env previewEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/subscriptions/" + url.PathEscape(id) + "/upcoming",
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return env.Data.Invoice, nil
}

// ListAutoPaginate returns a [*pagination.Iter] that walks every subscription
// matching params. Pages are fetched lazily.
//
// On Go 1.23+, the iterator also supports range-over-func via [pagination.Iter.All].
func (c *Client) ListAutoPaginate(ctx context.Context, params *ListParams) *pagination.Iter[Subscription] {
	startPage := 0
	if params != nil && params.Page != nil {
		startPage = *params.Page
	}

	return pagination.NewIter(startPage, func(page int) ([]Subscription, bool, error) {
		pageParams := ListParams{}
		if params != nil {
			pageParams = *params
		}
		pageParams.Page = &page

		var out ListResponse
		if err := c.backend.Do(ctx, core.Request{
			Method: http.MethodGet,
			Path:   "/subscriptions",
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
			Message: "subscriptions." + method + ": id must be non-empty",
		}}
	}
	return nil
}

func missingBody(method string) error {
	return &threecommon.ValidationError{APIError: &threecommon.APIError{
		Code:    "missing_body",
		Message: "subscriptions." + method + ": params must be non-nil",
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
	if p.ContactID != "" {
		q["contactId"] = p.ContactID
	}
	if p.PriceID != "" {
		q["priceId"] = p.PriceID
	}
	if p.Fields != "" {
		q["fields"] = p.Fields
	}
	if len(q) == 0 {
		return nil
	}
	return q
}
