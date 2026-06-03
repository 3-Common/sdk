package entitlements

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

// Client is the entitlements resource client. Construct one via [New] for
// standalone use, or use [github.com/3-Common/sdk/sdk-go/client.API.Entitlements]
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

// List fetches a single page of entitlement balance records. To iterate every
// entitlement matching a filter, use [Client.ListAutoPaginate].
func (c *Client) List(ctx context.Context, params *ListParams) (*ListResponse, error) {
	var out ListResponse
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/entitlements",
		Query:  encodeListParams(params),
		Out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Retrieve fetches a single entitlement record by ID, including grant history.
func (c *Client) Retrieve(ctx context.Context, id string, params *RetrieveParams) (*Entitlement, error) {
	if id == "" {
		return nil, &threecommon.ValidationError{APIError: &threecommon.APIError{
			Code:    "missing_id",
			Message: "entitlements.Retrieve: id must be non-empty",
		}}
	}

	var query map[string]string
	if params != nil && params.Fields != "" {
		query = map[string]string{"fields": params.Fields}
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/entitlements/" + url.PathEscape(id),
		Query:  query,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Lookup returns the unique entitlement for a (ContactID, FeatureKey) pair.
// Returns a [*threecommon.NotFoundError] when no record exists yet.
func (c *Client) Lookup(ctx context.Context, params *LookupParams) (*Entitlement, error) {
	if params == nil {
		return nil, missingBody("Lookup")
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/entitlements/lookup",
		Query:  encodeLookupParams(params),
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Grant adds a manual entitlement grant for a (ContactID, FeatureKey) — useful
// for admin top-ups, comp credits, or migration. Idempotent on GrantID.
func (c *Client) Grant(ctx context.Context, params *GrantParams) (*Entitlement, error) {
	if params == nil {
		return nil, missingBody("Grant")
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/entitlements/grants",
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Consume debits units from a customer's entitlement balance. Returns a
// [*threecommon.ConflictError] on insufficient balance. Debits one_time_addon
// grants first, then manual, then subscription_recurring (FIFO within source).
func (c *Client) Consume(ctx context.Context, params *ConsumeParams) (*Entitlement, error) {
	if params == nil {
		return nil, missingBody("Consume")
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/entitlements/consume",
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// ListAutoPaginate returns a [*pagination.Iter] that walks every entitlement
// matching params. Pages are fetched lazily — one HTTP call per page, only
// when the consumer drains the previous page's buffer.
//
// On Go 1.23+, the iterator also supports range-over-func via [pagination.Iter.All].
func (c *Client) ListAutoPaginate(ctx context.Context, params *ListParams) *pagination.Iter[Entitlement] {
	startPage := 0
	if params != nil && params.Page != nil {
		startPage = *params.Page
	}

	return pagination.NewIter(startPage, func(page int) ([]Entitlement, bool, error) {
		// Build a fresh ListParams per page so we don't mutate caller state.
		pageParams := ListParams{}
		if params != nil {
			pageParams = *params
		}
		pageParams.Page = &page

		var out ListResponse
		if err := c.backend.Do(ctx, core.Request{
			Method: http.MethodGet,
			Path:   "/entitlements",
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

func missingBody(method string) error {
	return &threecommon.ValidationError{APIError: &threecommon.APIError{
		Code:    "missing_body",
		Message: "entitlements." + method + ": params must be non-nil",
	}}
}

// encodeListParams converts ListParams into a query map. nil-safe; nil pointer
// fields and empty strings are omitted.
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
	if p.ContactID != "" {
		q["contactId"] = p.ContactID
	}
	if p.FeatureKey != "" {
		q["featureKey"] = p.FeatureKey
	}
	if p.MinBalance != nil {
		q["minBalance"] = strconv.FormatInt(*p.MinBalance, 10)
	}
	if p.Fields != "" {
		q["fields"] = p.Fields
	}
	if len(q) == 0 {
		return nil
	}
	return q
}

// encodeLookupParams converts LookupParams into a query map.
func encodeLookupParams(p *LookupParams) map[string]string {
	q := map[string]string{}
	if p.ContactID != "" {
		q["contactId"] = p.ContactID
	}
	if p.FeatureKey != "" {
		q["featureKey"] = p.FeatureKey
	}
	if p.Fields != "" {
		q["fields"] = p.Fields
	}
	if len(q) == 0 {
		return nil
	}
	return q
}
