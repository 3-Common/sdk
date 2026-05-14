package events

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

// Client is the events resource client. Construct one via [New] for
// standalone use, or use [github.com/3-Common/sdk/sdk-go/client.API.Events]
// when you need multiple resources sharing a single backend.
type Client struct {
	backend *core.Client
}

// New constructs a [*Client] from a [threecommon.Config]. Validates the
// config and returns a [*threecommon.ValidationError] for missing or invalid
// fields.
func New(cfg threecommon.Config) (*Client, error) {
	backend, err := core.NewFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{backend: backend}, nil
}

// FromBackend wraps an existing backend. Used by the
// [github.com/3-Common/sdk/sdk-go/client] aggregator package; users should not call this directly.
func FromBackend(b *core.Client) *Client {
	return &Client{backend: b}
}

// List fetches a single page of events. To iterate every event matching a filter, use [Client.ListAutoPaginate].
func (c *Client) List(ctx context.Context, params *ListParams) (*ListResponse, error) {
	var out ListResponse
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/events",
		Query:  encodeListParams(params),
		Out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Retrieve fetches a single event by ID.
func (c *Client) Retrieve(ctx context.Context, id string, params *RetrieveParams) (*Event, error) {
	if id == "" {
		return nil, &threecommon.ValidationError{APIError: &threecommon.APIError{
			Code:    "missing_id",
			Message: "events.Retrieve: id must be non-empty",
		}}
	}

	var query map[string]string
	if params != nil && params.Fields != "" {
		query = map[string]string{"fields": params.Fields}
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/events/" + url.PathEscape(id),
		Query:  query,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Update applies a partial-update to an event. Fields not present in params
// (i.e. nil-pointer fields) are left unchanged.
func (c *Client) Update(ctx context.Context, id string, params *UpdateParams) (*Event, error) {
	if id == "" {
		return nil, &threecommon.ValidationError{APIError: &threecommon.APIError{
			Code:    "missing_id",
			Message: "events.Update: id must be non-empty",
		}}
	}
	if params == nil {
		return nil, &threecommon.ValidationError{APIError: &threecommon.APIError{
			Code:    "missing_body",
			Message: "events.Update: params must be non-nil",
		}}
	}

	var env retrieveEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPatch,
		Path:   "/events/" + url.PathEscape(id),
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// ListAutoPaginate returns a [*threecommon.Iter] that walks every event
// matching params. Pages are fetched lazily — one HTTP call per page, only
// when the consumer drains the previous page's buffer.
//
// On Go 1.23+, the iterator also supports range-over-func via [threecommon.Iter.All].
func (c *Client) ListAutoPaginate(ctx context.Context, params *ListParams) *pagination.Iter[Event] {
	startPage := 0
	if params != nil && params.Page != nil {
		startPage = *params.Page
	}

	return pagination.NewIter(startPage, func(page int) ([]Event, bool, error) {
		// Build a fresh ListParams per page so we don't mutate caller state.
		pageParams := ListParams{}
		if params != nil {
			pageParams = *params
		}
		pageParams.Page = &page

		// Reuse the parent ctx; iteration honors caller-side cancellation.
		var out ListResponse
		if err := c.backend.Do(ctx, core.Request{
			Method: http.MethodGet,
			Path:   "/events",
			Query:  encodeListParams(&pageParams),
			Out:    &out,
		}); err != nil {
			// Surface context errors as ConnectionError already (handled by
			// httpclient); just pass through.
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil, false, err
			}
			return nil, false, err
		}
		return out.Data, out.HasMore, nil
	})
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
	if p.Status != "" {
		q["status"] = string(p.Status)
	}
	if p.Search != "" {
		q["search"] = p.Search
	}
	if p.StartBefore != "" {
		q["startBefore"] = p.StartBefore
	}
	if p.StartAfter != "" {
		q["startAfter"] = p.StartAfter
	}
	if p.SortField != "" {
		q["sortField"] = p.SortField
	}
	if p.SortDirection != "" {
		q["sortDirection"] = p.SortDirection
	}
	if p.Filters != "" {
		q["filters"] = p.Filters
	}
	if p.Fields != "" {
		q["fields"] = p.Fields
	}
	if len(q) == 0 {
		return nil
	}
	return q
}
