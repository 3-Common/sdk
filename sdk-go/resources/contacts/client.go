package contacts

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

// Client is the contacts resource client. Construct one via [New] for
// standalone use, or use [github.com/3-Common/sdk/sdk-go/client.API.Contacts]
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

// List fetches a single page of contacts. To iterate every contact matching
// a filter, use [Client.ListAutoPaginate].
func (c *Client) List(ctx context.Context, params *ListParams) (*ListResponse, error) {
	var out ListResponse
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/contacts",
		Query:  encodeListParams(params),
		Out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Count returns the total contact count for the host.
func (c *Client) Count(ctx context.Context) (int64, error) {
	var env countEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/contacts/count",
		Out:    &env,
	}); err != nil {
		return 0, err
	}
	return env.Data.Count, nil
}

// Retrieve fetches a single contact by id.
func (c *Client) Retrieve(ctx context.Context, id string) (*Contact, error) {
	if err := requireID("Retrieve", id); err != nil {
		return nil, err
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/contacts/" + url.PathEscape(id),
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Create makes a new contact. Returns a 409 ConflictError if a contact with
// the same email already exists for the host.
func (c *Client) Create(ctx context.Context, params *CreateParams) (*Contact, error) {
	if params == nil {
		return nil, missingBody("Create")
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/contacts",
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Update changes a contact's profile fields. Returns the richer
// order-details projection ([WithOrderDetails]), not the compact [Contact]
// shape that Retrieve returns. Optionally absorbs a second contact when
// MergeWith + Resolution are set together.
func (c *Client) Update(ctx context.Context, id string, params *UpdateParams) (*WithOrderDetails, error) {
	if err := requireID("Update", id); err != nil {
		return nil, err
	}
	if params == nil {
		return nil, missingBody("Update")
	}

	var env orderDetailsEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPatch,
		Path:   "/contacts/" + url.PathEscape(id),
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Delete permanently removes a contact. Echoes the removed contact's id.
func (c *Client) Delete(ctx context.Context, id string) (*DeleteResult, error) {
	if err := requireID("Delete", id); err != nil {
		return nil, err
	}

	var env deleteEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodDelete,
		Path:   "/contacts/" + url.PathEscape(id),
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// BulkUpsert upserts up to 10,000 contacts in one round-trip. Deduplicated
// server-side by email; existing rows are updated rather than rejected.
func (c *Client) BulkUpsert(ctx context.Context, params *BulkUpsertParams) (*BulkUpsertResult, error) {
	if params == nil {
		return nil, missingBody("BulkUpsert")
	}

	var env bulkUpsertEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/contacts/bulk",
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// ListActivity returns a paginated activity log for a contact (checkouts,
// refunds, scans, emails, invoice payments).
func (c *Client) ListActivity(ctx context.Context, id string, params *ActivityListParams) (*ListActivityResponse, error) {
	if err := requireID("ListActivity", id); err != nil {
		return nil, err
	}

	var out ListActivityResponse
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/contacts/" + url.PathEscape(id) + "/activity",
		Query:  encodeActivityParams(params),
		Out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListAutoPaginate returns a [*pagination.Iter] that walks every contact
// matching params. Pages are fetched lazily — one HTTP call per page, only
// when the previous page's buffer drains.
func (c *Client) ListAutoPaginate(ctx context.Context, params *ListParams) *pagination.Iter[Contact] {
	startPage := 0
	if params != nil && params.PageNumber != nil {
		startPage = *params.PageNumber
	}

	return pagination.NewIter(startPage, func(page int) ([]Contact, bool, error) {
		pageParams := ListParams{}
		if params != nil {
			pageParams = *params
		}
		pageParams.PageNumber = &page

		var out ListResponse
		if err := c.backend.Do(ctx, core.Request{
			Method: http.MethodGet,
			Path:   "/contacts",
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

// ListActivityAutoPaginate returns a [*pagination.Iter] that walks every
// activity record for a contact.
func (c *Client) ListActivityAutoPaginate(ctx context.Context, id string, params *ActivityListParams) *pagination.Iter[Activity] {
	if id == "" {
		// Return an iterator that fails on first Next() with the validation
		// error — matches the events/invoices pattern of surfacing missing-id
		// validation through iter.Err().
		err := requireID("ListActivityAutoPaginate", id)
		return pagination.NewIter(0, func(_ int) ([]Activity, bool, error) {
			return nil, false, err
		})
	}

	startPage := 0
	if params != nil && params.PageNumber != nil {
		startPage = *params.PageNumber
	}

	return pagination.NewIter(startPage, func(page int) ([]Activity, bool, error) {
		pageParams := ActivityListParams{}
		if params != nil {
			pageParams = *params
		}
		pageParams.PageNumber = &page

		var out ListActivityResponse
		if err := c.backend.Do(ctx, core.Request{
			Method: http.MethodGet,
			Path:   "/contacts/" + url.PathEscape(id) + "/activity",
			Query:  encodeActivityParams(&pageParams),
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

// RetrievePaymentMethod fetches the saved card on file for a contact, or nil
// when none is saved. One card is supported per contact.
func (c *Client) RetrievePaymentMethod(ctx context.Context, id string) (*PaymentMethod, error) {
	if err := requireID("RetrievePaymentMethod", id); err != nil {
		return nil, err
	}

	var env paymentMethodEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/contacts/" + url.PathEscape(id) + "/payment-methods",
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return env.Data, nil
}

// AttachPaymentMethod persists the card from a confirmed Stripe SetupIntent
// against the contact. The SetupIntent is re-verified server-side before the
// card is saved, replacing any existing card on file. The returned
// [AttachPaymentMethodResult] carries both the saved payment method and a
// ReplacedExisting flag reporting whether an existing card was replaced.
func (c *Client) AttachPaymentMethod(ctx context.Context, id string, params *AttachPaymentMethodParams) (*AttachPaymentMethodResult, error) {
	if err := requireID("AttachPaymentMethod", id); err != nil {
		return nil, err
	}
	if params == nil {
		return nil, missingBody("AttachPaymentMethod")
	}

	var out AttachPaymentMethodResult
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/contacts/" + url.PathEscape(id) + "/payment-methods",
		Body:   params,
		Out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreatePaymentMethodSetupIntent begins saving a card for a contact. It returns
// a Stripe SetupIntent whose ClientSecret is confirmed client-side with Stripe
// Elements; once confirmed, call [Client.AttachPaymentMethod] with the returned
// SetupIntentID to persist the card. Sends no request body.
func (c *Client) CreatePaymentMethodSetupIntent(ctx context.Context, id string) (*PaymentMethodSetupIntent, error) {
	if err := requireID("CreatePaymentMethodSetupIntent", id); err != nil {
		return nil, err
	}

	var env setupIntentEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/contacts/" + url.PathEscape(id) + "/payment-methods/setup-intent",
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// RemovePaymentMethod detaches the saved card from Stripe and removes it from
// the contact. methodID is the payment-method id (the deeper path segment).
func (c *Client) RemovePaymentMethod(ctx context.Context, id, methodID string) (*RemovedPaymentMethod, error) {
	if err := requireID("RemovePaymentMethod", id); err != nil {
		return nil, err
	}
	if err := requireMethodID("RemovePaymentMethod", methodID); err != nil {
		return nil, err
	}

	var env removedPaymentMethodEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodDelete,
		Path:   "/contacts/" + url.PathEscape(id) + "/payment-methods/" + url.PathEscape(methodID),
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func requireID(method, id string) error {
	if id == "" {
		return &threecommon.ValidationError{APIError: &threecommon.APIError{
			Code:    "missing_id",
			Message: "contacts." + method + ": id must be non-empty",
		}}
	}
	return nil
}

func requireMethodID(method, methodID string) error {
	if methodID == "" {
		return &threecommon.ValidationError{APIError: &threecommon.APIError{
			Code:    "missing_method_id",
			Message: "contacts." + method + ": methodId must be non-empty",
		}}
	}
	return nil
}

func missingBody(method string) error {
	return &threecommon.ValidationError{APIError: &threecommon.APIError{
		Code:    "missing_body",
		Message: "contacts." + method + ": params must be non-nil",
	}}
}

func encodeListParams(p *ListParams) map[string]string {
	if p == nil {
		return nil
	}
	q := map[string]string{}
	if p.PageNumber != nil {
		q["pageNumber"] = strconv.Itoa(*p.PageNumber)
	}
	if p.PageSize != nil {
		q["pageSize"] = strconv.Itoa(*p.PageSize)
	}
	if p.SortField != "" {
		q["sortField"] = p.SortField
	}
	if p.SortDirection != "" {
		q["sortDirection"] = p.SortDirection
	}
	if p.Filter != "" {
		q["filter"] = string(p.Filter)
	}
	if p.Filters != "" {
		q["filters"] = p.Filters
	}
	if p.Search != "" {
		q["search"] = p.Search
	}
	if len(q) == 0 {
		return nil
	}
	return q
}

func encodeActivityParams(p *ActivityListParams) map[string]string {
	if p == nil {
		return nil
	}
	q := map[string]string{}
	if p.PageNumber != nil {
		q["pageNumber"] = strconv.Itoa(*p.PageNumber)
	}
	if p.PageSize != nil {
		q["pageSize"] = strconv.Itoa(*p.PageSize)
	}
	if p.Filter != "" {
		q["filter"] = string(p.Filter)
	}
	if p.Sort != "" {
		q["sort"] = p.Sort
	}
	if len(q) == 0 {
		return nil
	}
	return q
}
