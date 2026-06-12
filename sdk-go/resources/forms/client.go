package forms

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/internal/core"
	"github.com/3-Common/sdk/sdk-go/pagination"
)

// Client is the forms resource client. Construct one via [New] for standalone
// use, or use [github.com/3-Common/sdk/sdk-go/client.API.Forms] when you need
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

// List fetches a single page of forms. To iterate every form matching a
// filter, use [Client.ListAutoPaginate].
func (c *Client) List(ctx context.Context, params *ListParams) (*ListResponse, error) {
	var out ListResponse
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/forms",
		Query:  encodeListParams(params),
		Out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Retrieve fetches a single form, including its element tree, by id.
func (c *Client) Retrieve(ctx context.Context, id string) (*Form, error) {
	if err := requireID("Retrieve", "id", id); err != nil {
		return nil, err
	}

	var env formEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodGet,
		Path:   "/forms/" + url.PathEscape(id),
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Create makes a new form. Name and Type are required.
func (c *Client) Create(ctx context.Context, params *CreateParams) (*Form, error) {
	if params == nil {
		return nil, missingBody("Create")
	}

	var env formEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/forms",
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Update changes a form's settings. Only the fields set on params are
// changed; nullable fields accept [threecommon.Null] to clear the setting
// server-side.
func (c *Client) Update(ctx context.Context, id string, params *UpdateParams) (*Form, error) {
	if err := requireID("Update", "id", id); err != nil {
		return nil, err
	}
	if params == nil {
		return nil, missingBody("Update")
	}

	var env formEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPatch,
		Path:   "/forms/" + url.PathEscape(id),
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Duplicate copies a form. The optional params override the copy's name and
// status.
func (c *Client) Duplicate(ctx context.Context, id string, params *DuplicateParams) (*Form, error) {
	if err := requireID("Duplicate", "id", id); err != nil {
		return nil, err
	}

	req := core.Request{
		Method: http.MethodPost,
		Path:   "/forms/" + url.PathEscape(id) + "/duplicate",
	}
	// Avoid marshaling a typed-nil *DuplicateParams to a literal "null" body;
	// the override fields are optional, so an absent params means no body.
	if params != nil {
		req.Body = params
	}

	var env formEnvelope
	req.Out = &env
	if err := c.backend.Do(ctx, req); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// AddElement appends an element to a form and returns the created element.
func (c *Client) AddElement(ctx context.Context, id string, params *AddElementParams) (*Element, error) {
	if err := requireID("AddElement", "id", id); err != nil {
		return nil, err
	}
	if params == nil {
		return nil, missingBody("AddElement")
	}

	var env elementEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/forms/" + url.PathEscape(id) + "/elements",
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// UpdateElement edits an existing element. Only the fields set on params are
// changed; every field except Prompt accepts [threecommon.Null] to clear the
// setting server-side.
func (c *Client) UpdateElement(ctx context.Context, id, elementID string, params *UpdateElementParams) (*Element, error) {
	if err := requireID("UpdateElement", "id", id); err != nil {
		return nil, err
	}
	if err := requireID("UpdateElement", "elementId", elementID); err != nil {
		return nil, err
	}
	if params == nil {
		return nil, missingBody("UpdateElement")
	}

	var env elementEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPatch,
		Path:   "/forms/" + url.PathEscape(id) + "/elements/" + url.PathEscape(elementID),
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// DeleteElement removes an element from a form. Echoes the removed element's id.
func (c *Client) DeleteElement(ctx context.Context, id, elementID string) (*DeleteElementResult, error) {
	if err := requireID("DeleteElement", "id", id); err != nil {
		return nil, err
	}
	if err := requireID("DeleteElement", "elementId", elementID); err != nil {
		return nil, err
	}

	var env deleteElementEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodDelete,
		Path:   "/forms/" + url.PathEscape(id) + "/elements/" + url.PathEscape(elementID),
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// MoveElement repositions an element within the form. Returns the updated form.
func (c *Client) MoveElement(ctx context.Context, id, elementID string, params *MoveElementParams) (*Form, error) {
	if err := requireID("MoveElement", "id", id); err != nil {
		return nil, err
	}
	if err := requireID("MoveElement", "elementId", elementID); err != nil {
		return nil, err
	}
	if params == nil {
		return nil, missingBody("MoveElement")
	}

	var env formEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPut,
		Path:   "/forms/" + url.PathEscape(id) + "/elements/" + url.PathEscape(elementID) + "/position",
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// AddLogicRule adds a conditional-visibility rule to a selection or Yes/No
// element. Returns the updated source element.
func (c *Client) AddLogicRule(ctx context.Context, id, elementID string, params *AddLogicRuleParams) (*Element, error) {
	if err := requireID("AddLogicRule", "id", id); err != nil {
		return nil, err
	}
	if err := requireID("AddLogicRule", "elementId", elementID); err != nil {
		return nil, err
	}
	if params == nil {
		return nil, missingBody("AddLogicRule")
	}

	var env elementEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPost,
		Path:   "/forms/" + url.PathEscape(id) + "/elements/" + url.PathEscape(elementID) + "/logic-rules",
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// RemoveLogicRule removes the logic rule on elementID that reveals
// targetElementID. Returns the updated source element.
func (c *Client) RemoveLogicRule(ctx context.Context, id, elementID, targetElementID string) (*Element, error) {
	if err := requireID("RemoveLogicRule", "id", id); err != nil {
		return nil, err
	}
	if err := requireID("RemoveLogicRule", "elementId", elementID); err != nil {
		return nil, err
	}
	if err := requireID("RemoveLogicRule", "targetElementId", targetElementID); err != nil {
		return nil, err
	}

	var env elementEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodDelete,
		Path:   "/forms/" + url.PathEscape(id) + "/elements/" + url.PathEscape(elementID) + "/logic-rules/" + url.PathEscape(targetElementID),
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// EnableOtherOption enables the free-text "Other" option on a selection
// element. Returns the updated element.
func (c *Client) EnableOtherOption(ctx context.Context, id, elementID string, params *EnableOtherOptionParams) (*Element, error) {
	if err := requireID("EnableOtherOption", "id", id); err != nil {
		return nil, err
	}
	if err := requireID("EnableOtherOption", "elementId", elementID); err != nil {
		return nil, err
	}
	if params == nil {
		return nil, missingBody("EnableOtherOption")
	}

	var env elementEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodPut,
		Path:   "/forms/" + url.PathEscape(id) + "/elements/" + url.PathEscape(elementID) + "/other-option",
		Body:   params,
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// DisableOtherOption disables the "Other" option on a selection element.
// Returns the updated element.
func (c *Client) DisableOtherOption(ctx context.Context, id, elementID string) (*Element, error) {
	if err := requireID("DisableOtherOption", "id", id); err != nil {
		return nil, err
	}
	if err := requireID("DisableOtherOption", "elementId", elementID); err != nil {
		return nil, err
	}

	var env elementEnvelope
	if err := c.backend.Do(ctx, core.Request{
		Method: http.MethodDelete,
		Path:   "/forms/" + url.PathEscape(id) + "/elements/" + url.PathEscape(elementID) + "/other-option",
		Out:    &env,
	}); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// ListAutoPaginate returns a [*pagination.Iter] that walks every form matching
// params. Pages are fetched lazily, one HTTP call per page, only when the
// previous page's buffer drains.
func (c *Client) ListAutoPaginate(ctx context.Context, params *ListParams) *pagination.Iter[FormSummary] {
	startPage := 0
	if params != nil && params.Page != nil {
		startPage = *params.Page
	}

	return pagination.NewIter(startPage, func(page int) ([]FormSummary, bool, error) {
		pageParams := ListParams{}
		if params != nil {
			pageParams = *params
		}
		pageParams.Page = &page

		var out ListResponse
		if err := c.backend.Do(ctx, core.Request{
			Method: http.MethodGet,
			Path:   "/forms",
			Query:  encodeListParams(&pageParams),
			Out:    &out,
		}); err != nil {
			return nil, false, err
		}
		return out.Data, out.HasMore, nil
	})
}

func requireID(method, field, id string) error {
	if id == "" {
		return &threecommon.ValidationError{APIError: &threecommon.APIError{
			Code:    "missing_id",
			Message: "forms." + method + ": " + field + " must be non-empty",
		}}
	}
	return nil
}

func missingBody(method string) error {
	return &threecommon.ValidationError{APIError: &threecommon.APIError{
		Code:    "missing_body",
		Message: "forms." + method + ": params must be non-nil",
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
	if p.Type != "" {
		q["type"] = string(p.Type)
	}
	if len(q) == 0 {
		return nil
	}
	return q
}
