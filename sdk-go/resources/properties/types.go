// Package properties provides the properties resource client for the 3Common
// API. Properties are the custom-field definitions that can be attached to
// events, orders, tickets, and contacts.
//
// Most callers reach this package through
// [github.com/3-Common/sdk/sdk-go/client.API.Properties]:
//
//	api, _ := client.New(threecommon.Config{APIKey: "..."})
//	page, err := api.Properties.List(ctx, &properties.ListParams{
//		ObjectType: properties.ObjectTypeContact,
//		Status:     properties.StatusActive,
//	})
//
// To use the resource standalone (without the aggregator), call [New]:
//
//	cli, _ := properties.New(threecommon.Config{APIKey: "..."})
//	page, err := cli.List(ctx, nil)
//
// Type names inside this package omit the "Property" prefix to avoid stutter
// (e.g. properties.ListParams, not properties.PropertyListParams).
package properties

import "encoding/json"

// Type is the data type of a property. "type" can only be set when the
// property is created; it cannot be modified afterwards.
type Type string

// Type values returned by the API. Unknown values from a future API version
// will surface as the raw string.
const (
	TypeText           Type = "Text"
	TypeMultiLineText  Type = "Multi-line Text"
	TypeSelectOne      Type = "Select One"
	TypeYesNo          Type = "Yes/No"
	TypeSelectMultiple Type = "Select Multiple"
	TypeDate           Type = "Date"
	TypeFile           Type = "File"
	TypeEmail          Type = "Email"
	TypePhone          Type = "Phone"
)

// Status is the lifecycle status of a property. "archived" properties are
// soft-deleted: existing references remain valid, but only "active" properties
// should be used in new workflows, forms, etc.
type Status string

// Status values.
const (
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
)

// ObjectType is the type of object a property belongs to. Like Type, it can
// only be set at creation and cannot be modified afterwards.
//
//   - ObjectTypeEvent:   properties on events
//   - ObjectTypeOrder:   properties on orders (buyer-level)
//   - ObjectTypeTicket:  properties on individual products within an order
//   - ObjectTypeContact: properties on customer contact records
type ObjectType string

// ObjectType values.
const (
	ObjectTypeEvent   ObjectType = "event"
	ObjectTypeOrder   ObjectType = "order"
	ObjectTypeTicket  ObjectType = "ticket"
	ObjectTypeContact ObjectType = "contact"
)

// Option is one selectable choice on a "Select One" or "Select Multiple"
// property.
//
// WARNING: an option's Value is the identity persisted on every instance that
// selected it. Changing an existing option's Value does NOT migrate those
// instances. Rename an option by changing its Label only.
type Option struct {
	// Value is the value stored when this option is selected.
	Value string `json:"value"`
	// Label is the display label for this option.
	Label string `json:"label"`
}

// Property is the resource shape returned by List, Retrieve, Create, and
// Update. Options is present only for "Select One" and "Select Multiple"
// properties.
type Property struct {
	ID          string     `json:"id"`
	Type        Type       `json:"type"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Status      Status     `json:"status"`
	ObjectType  ObjectType `json:"objectType"`
	Options     []Option   `json:"options,omitempty"`
}

// ListParams are the query parameters accepted by [Client.List] and
// [Client.ListAutoPaginate].
type ListParams struct {
	// Page is the 0-indexed page number. Nil uses the server default (0).
	Page *int

	// PageSize is the items-per-page cap (1-100). Nil uses the server
	// default (20).
	PageSize *int

	// ObjectType filters by the type of object the property belongs to.
	ObjectType ObjectType

	// PropertyType filters by property data type.
	PropertyType Type

	// Status filters by property status.
	Status Status

	// Sort is the field to sort by: "name", "description", "type",
	// "objectType", or "status". Defaults to "name".
	Sort string

	// Order is "asc" or "desc". Defaults to "asc".
	Order string

	// Search performs a case-insensitive match against the property name.
	Search string
}

// CreateParams is the body shape accepted by [Client.Create]. Type and
// ObjectType can only be set here; they cannot be modified later. For
// "Select One" and "Select Multiple" types, Options is required and must have
// at least one entry.
type CreateParams struct {
	Type        Type       `json:"type"`
	Name        string     `json:"name"`
	Status      Status     `json:"status"`
	ObjectType  ObjectType `json:"objectType"`
	Description string     `json:"description,omitempty"`
	Options     []Option   `json:"options,omitempty"`
}

// UpdateParams is the body shape accepted by [Client.Update]. Only the fields
// that are set are sent; the rest are left unchanged server-side. Type and
// ObjectType cannot be modified on an existing property.
type UpdateParams struct {
	// Name renames the property when non-empty.
	Name string

	// Status changes the property status when non-empty. Set to
	// [StatusArchived] to retire a property.
	Status Status

	// Options replaces the option set for "Select One" / "Select Multiple"
	// properties when non-nil.
	Options []Option

	// Description sets a new description when non-nil. To remove the
	// description entirely, leave this nil and set ClearDescription instead.
	Description *string

	// ClearDescription sends an explicit null for description, removing it.
	// Takes precedence over Description.
	ClearDescription bool
}

// MarshalJSON emits only the fields that are set, sending an explicit null for
// description when ClearDescription is true.
func (p UpdateParams) MarshalJSON() ([]byte, error) {
	m := make(map[string]any)
	if p.Name != "" {
		m["name"] = p.Name
	}
	if p.Status != "" {
		m["status"] = p.Status
	}
	if p.Options != nil {
		m["options"] = p.Options
	}
	switch {
	case p.ClearDescription:
		m["description"] = nil
	case p.Description != nil:
		m["description"] = *p.Description
	}
	return json.Marshal(m)
}

// ListResponse is the body returned by GET /v1/properties.
type ListResponse struct {
	Data    []Property `json:"data"`
	HasMore bool       `json:"hasMore"`
}

// dataEnvelope is the {"data": Property} shape used by the detail-returning
// endpoints (Retrieve, Create, Update).
type dataEnvelope struct {
	Data Property `json:"data"`
}
