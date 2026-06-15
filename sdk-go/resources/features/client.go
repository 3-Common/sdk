package features

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

// Client is the features resource client. Construct one via [New] for
// standalone use, or use [github.com/3-Common/sdk/sdk-go/client.API.Features]
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

// List fetches a single page of features. To iterate every feature matching a
// filter, use [Client.ListAutoPaginate].
func (c *Client) List(ctx context.Context, params *ListParams) (*ListResponse, error) {
	var out ListResponse
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/features",
		Query:  encodeListParams(params),
		Out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Resolve resolves the current value of a feature for a customer by walking
// active subscriptions → prices → feature grants. Returns a
// [*threecommon.NotFoundError] when the feature key is unknown.
func (c *Client) Resolve(ctx context.Context, params *ResolveParams) (*ResolvedFeature, error) {
	if params == nil {
		return nil, missingBody("Resolve")
	}

	var env resolveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/features/resolve",
		Query:  encodeResolveParams(params),
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Retrieve fetches a single feature by ID.
func (c *Client) Retrieve(ctx context.Context, id string, params *RetrieveParams) (*Feature, error) {
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
		Path:   "/features/" + url.PathEscape(id),
		Query:  query,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Create makes a new feature in the catalog. The Key is the stable
// application-facing identifier; Type decides how prices grant the feature and
// how it resolves.
func (c *Client) Create(ctx context.Context, params *CreateParams) (*Feature, error) {
	if params == nil {
		return nil, missingBody("Create")
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/features",
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Update applies a partial update to a feature. Mutable fields: Name,
// Description, EnumValues, Metadata. Key and Type are immutable — archive and
// create a new feature to change them.
func (c *Client) Update(ctx context.Context, id string, params *UpdateParams) (*Feature, error) {
	if err := requireID("Update", id); err != nil {
		return nil, err
	}
	if params == nil {
		return nil, missingBody("Update")
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPatch,
		Path:   "/features/" + url.PathEscape(id),
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Archive soft-archives a feature. Idempotent.
func (c *Client) Archive(ctx context.Context, id string) (*Feature, error) {
	if err := requireID("Archive", id); err != nil {
		return nil, err
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/features/" + url.PathEscape(id) + "/archive",
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Unarchive reactivates a previously archived feature. Idempotent.
func (c *Client) Unarchive(ctx context.Context, id string) (*Feature, error) {
	if err := requireID("Unarchive", id); err != nil {
		return nil, err
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/features/" + url.PathEscape(id) + "/unarchive",
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// ListAutoPaginate returns a [*pagination.Iter] that walks every feature
// matching params. Pages are fetched lazily — one HTTP call per page, only
// when the consumer drains the previous page's buffer.
//
// On Go 1.23+, the iterator also supports range-over-func via [pagination.Iter.All].
func (c *Client) ListAutoPaginate(ctx context.Context, params *ListParams) *pagination.Iter[Feature] {
	startPage := 0
	if params != nil && params.Page != nil {
		startPage = *params.Page
	}

	return pagination.NewIter(startPage, func(page int) ([]Feature, bool, error) {
		// Build a fresh ListParams per page so we don't mutate caller state.
		pageParams := ListParams{}
		if params != nil {
			pageParams = *params
		}
		pageParams.Page = &page

		var out ListResponse
		if err := c.backend.Do(ctx, core.Request{
			Method: http.MethodGet,
			Path:   "/features",
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
			Message: "features." + method + ": id must be non-empty",
		}}
	}
	return nil
}

func missingBody(method string) error {
	return &threecommon.ValidationError{APIError: &threecommon.APIError{
		Code:    "missing_body",
		Message: "features." + method + ": params must be non-nil",
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

// encodeResolveParams converts ResolveParams into a query map.
func encodeResolveParams(p *ResolveParams) map[string]string {
	q := map[string]string{}
	if p.ContactID != "" {
		q["contactId"] = p.ContactID
	}
	if p.FeatureKey != "" {
		q["featureKey"] = p.FeatureKey
	}
	if len(q) == 0 {
		return nil
	}
	return q
}
