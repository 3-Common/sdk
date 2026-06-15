package prices

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

// Client is the prices resource client. Construct one via [New] for standalone
// use, or use [github.com/3-Common/sdk/sdk-go/client.API.Prices] when you need
// multiple resources sharing a single backend.
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

// List fetches a single page of prices. To iterate every price matching a
// filter, use [Client.ListAutoPaginate].
func (c *Client) List(ctx context.Context, params *ListParams) (*ListResponse, error) {
	var out ListResponse
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/prices",
		Query:  encodeListParams(params),
		Out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Retrieve fetches a single price by ID.
func (c *Client) Retrieve(ctx context.Context, id string, params *RetrieveParams) (*Price, error) {
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
		Path:   "/prices/" + url.PathEscape(id),
		Query:  query,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Create makes a new price for a product. Defines cadence (one_time or
// recurring), per-unit cost, and an optional array of typed feature grants.
func (c *Client) Create(ctx context.Context, params *CreateParams) (*Price, error) {
	if params == nil {
		return nil, missingBody("Create")
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/prices",
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Update applies a partial update to a price. Mutable fields: UnitAmount,
// Recurring, Features, Nickname, Metadata. To switch type/currency/product,
// archive and create a new price.
func (c *Client) Update(ctx context.Context, id string, params *UpdateParams) (*Price, error) {
	if err := requireID("Update", id); err != nil {
		return nil, err
	}
	if params == nil {
		return nil, missingBody("Update")
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPatch,
		Path:   "/prices/" + url.PathEscape(id),
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Archive soft-archives a price. Idempotent. Existing subscriptions are
// unaffected; new subscriptions cannot select this price until unarchived.
func (c *Client) Archive(ctx context.Context, id string) (*Price, error) {
	if err := requireID("Archive", id); err != nil {
		return nil, err
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/prices/" + url.PathEscape(id) + "/archive",
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Unarchive reactivates a previously archived price. Idempotent.
func (c *Client) Unarchive(ctx context.Context, id string) (*Price, error) {
	if err := requireID("Unarchive", id); err != nil {
		return nil, err
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/prices/" + url.PathEscape(id) + "/unarchive",
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// ListAutoPaginate returns a [*pagination.Iter] that walks every price matching
// params. Pages are fetched lazily — one HTTP call per page, only when the
// consumer drains the previous page's buffer.
//
// On Go 1.23+, the iterator also supports range-over-func via [pagination.Iter.All].
func (c *Client) ListAutoPaginate(ctx context.Context, params *ListParams) *pagination.Iter[Price] {
	startPage := 0
	if params != nil && params.Page != nil {
		startPage = *params.Page
	}

	return pagination.NewIter(startPage, func(page int) ([]Price, bool, error) {
		// Build a fresh ListParams per page so we don't mutate caller state.
		pageParams := ListParams{}
		if params != nil {
			pageParams = *params
		}
		pageParams.Page = &page

		var out ListResponse
		if err := c.backend.Do(ctx, core.Request{
			Method: http.MethodGet,
			Path:   "/prices",
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
			Message: "prices." + method + ": id must be non-empty",
		}}
	}
	return nil
}

func missingBody(method string) error {
	return &threecommon.ValidationError{APIError: &threecommon.APIError{
		Code:    "missing_body",
		Message: "prices." + method + ": params must be non-nil",
	}}
}

// encodeListParams converts ListParams into a query map. nil-safe; nil pointer
// fields and empty strings are omitted. Booleans render lowercase.
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
	if p.ProductID != "" {
		q["productId"] = p.ProductID
	}
	if p.Type != "" {
		q["type"] = string(p.Type)
	}
	if p.Active != nil {
		q["active"] = strconv.FormatBool(*p.Active)
	}
	if p.Fields != "" {
		q["fields"] = p.Fields
	}
	if len(q) == 0 {
		return nil
	}
	return q
}
