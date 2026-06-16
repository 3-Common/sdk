package properties

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

// Client is the properties resource client. Construct one via [New] for
// standalone use, or use
// [github.com/3-Common/sdk/sdk-go/client.API.Properties] when you need
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

// List fetches a single page of properties. To iterate every property matching
// a filter, use [Client.ListAutoPaginate].
func (c *Client) List(ctx context.Context, params *ListParams) (*ListResponse, error) {
	var out ListResponse
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/properties",
		Query:  encodeListParams(params),
		Out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Retrieve fetches a single property by id.
func (c *Client) Retrieve(ctx context.Context, id string) (*Property, error) {
	if err := requireID("Retrieve", id); err != nil {
		return nil, err
	}

	var env dataEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/properties/" + url.PathEscape(id),
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Create makes a new property. Type and ObjectType can only be set here and
// cannot be modified later. For "Select One" and "Select Multiple" types,
// Options is required and must have at least one entry.
func (c *Client) Create(ctx context.Context, params *CreateParams) (*Property, error) {
	if params == nil {
		return nil, missingBody("Create")
	}

	var env dataEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/properties",
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Update edits the property with the given id. Only the fields set on params
// are modified; Type and ObjectType cannot be changed. Returns the full
// property after the update is applied.
func (c *Client) Update(ctx context.Context, id string, params *UpdateParams) (*Property, error) {
	if err := requireID("Update", id); err != nil {
		return nil, err
	}
	if params == nil {
		return nil, missingBody("Update")
	}

	var env dataEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPatch,
		Path:   "/properties/" + url.PathEscape(id),
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// ListAutoPaginate returns a [*pagination.Iter] that walks every property
// matching params. Pages are fetched lazily - one HTTP call per page, only
// when the previous page's buffer drains.
func (c *Client) ListAutoPaginate(ctx context.Context, params *ListParams) *pagination.Iter[Property] {
	startPage := 0
	if params != nil && params.Page != nil {
		startPage = *params.Page
	}

	return pagination.NewIter(startPage, func(page int) ([]Property, bool, error) {
		pageParams := ListParams{}
		if params != nil {
			pageParams = *params
		}
		pageParams.Page = &page

		var out ListResponse
		if err := c.backend.Do(ctx, core.Request{
			Method: http.MethodGet,
			Path:   "/properties",
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
			Message: "properties." + method + ": id must be non-empty",
		}}
	}
	return nil
}

func missingBody(method string) error {
	return &threecommon.ValidationError{APIError: &threecommon.APIError{
		Code:    "missing_body",
		Message: "properties." + method + ": params must be non-nil",
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
	if p.ObjectType != "" {
		q["objectType"] = string(p.ObjectType)
	}
	if p.PropertyType != "" {
		q["propertyType"] = string(p.PropertyType)
	}
	if p.Status != "" {
		q["status"] = string(p.Status)
	}
	if p.Sort != "" {
		q["sort"] = p.Sort
	}
	if p.Order != "" {
		q["order"] = p.Order
	}
	if p.Search != "" {
		q["search"] = p.Search
	}
	if len(q) == 0 {
		return nil
	}
	return q
}
