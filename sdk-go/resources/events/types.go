// Package events provides the events resource client for the 3Common API.
//
// Most callers reach this package through
// [github.com/3-Common/sdk/sdk-go/client.API.Events]:
//
//	api, _ := client.New(threecommon.Config{APIKey: "..."})
//	result, err := api.Events.List(ctx, &events.ListParams{Status: events.StatusOpen})
//
// To use the resource standalone (without the aggregator), call [New]:
//
//	ev, _ := events.New(threecommon.Config{APIKey: "..."})
//	result, err := ev.List(ctx, nil)
//
// Type names inside this package omit the "Event" prefix to avoid stutter
// (e.g. events.ListParams, not events.EventListParams).
package events

import "github.com/3-Common/sdk/sdk-go/filters"

// Status is the lifecycle status of an event.
type Status string

// Status values returned by the API. Unknown values from a future API
// version will surface as the raw string.
const (
	StatusDraft       Status = "draft"
	StatusOpen        Status = "open"
	StatusClosed      Status = "closed"
	StatusUnpublished Status = "unpublished"
	StatusCancelled   Status = "cancelled"
	StatusPostponed   Status = "postponed"
	StatusSchedule    Status = "schedule"
)

// Event is the resource shape returned by the API. Pointer fields are
// populated only when the server returned them — list responses with a
// `Fields` filter omit unrequested values.
type Event struct {
	ID            string `json:"id"`
	Name          string `json:"name,omitempty"`
	Type          string `json:"type,omitempty"`
	Schedule      string `json:"schedule,omitempty"`
	Start         string `json:"start,omitempty"` // ISO 8601
	Status        Status `json:"status,omitempty"`
	ItemsSold     *int64 `json:"itemsSold,omitempty"`
	RevenueCents  *int64 `json:"revenueCents,omitempty"`
	MinPriceCents *int64 `json:"minPriceCents,omitempty"`
	MaxPriceCents *int64 `json:"maxPriceCents,omitempty"`
	Currency      string `json:"currency,omitempty"`
	IsPublic      *bool  `json:"isPublic,omitempty"`
	IsVirtual     *bool  `json:"isVirtual,omitempty"`
}

// ListParams are the query parameters accepted by [Client.List] and [Client.ListAutoPaginate].
type ListParams struct {
	// Page is the 0-indexed page number. Nil uses the server default (0).
	Page *int

	// PageSize is the items-per-page cap. Nil uses the server default (20);
	// the server caps at 50.
	PageSize *int

	// Status filters by lifecycle status. Empty includes all statuses.
	Status Status

	// Search is a case-insensitive partial match on name or address.
	Search string

	// StartBefore is an ISO 8601 timestamp; only events starting on or before this date are returned.
	StartBefore string

	// StartAfter is an ISO 8601 timestamp; only events starting on or after this date are returned.
	StartAfter string

	// SortField controls the sort order. Defaults to "start".
	SortField string

	// SortDirection is "asc" or "desc". Defaults to "desc".
	SortDirection string

	// Filters is a JSON-encoded filter group, produced by
	// [github.com/3-Common/sdk/sdk-go/filters.SerializableFilter.Serialize].
	//
	//	import "github.com/3-Common/sdk/sdk-go/filters"
	//
	//	f := filters.And(
	//		filters.Field("status").IsAnyOf("open"),
	//		filters.Field("ticketSum").IsGreaterThan(10),
	//	)
	//	out, _ := f.Serialize()
	//	api.Events.List(ctx, &events.ListParams{Filters: out})
	Filters string

	// Fields is a comma-separated list of fields to include in the response.
	// Empty returns all fields.
	Fields string
}

// FilterWith is a convenience that serializes f and assigns it to
// [ListParams.Filters]. Returns the modified params for chaining.
//
//	api.Events.List(ctx, (&events.ListParams{Status: events.StatusOpen}).
//		FilterWith(filters.And(filters.Field("ticketSum").IsGreaterThan(10))))
func (p *ListParams) FilterWith(f *filters.SerializableFilter) *ListParams {
	if f != nil {
		p.Filters = f.MustSerialize()
	}
	return p
}

// RetrieveParams are the query parameters accepted by [Client.Retrieve].
type RetrieveParams struct {
	// Fields is a comma-separated list of fields to include in the response.
	Fields string
}

// UpdateParams is the body shape accepted by [Client.Update]. Only fields with non-nil pointers are sent.
type UpdateParams struct {
	Name *string `json:"name,omitempty"`
}

// ListResponse is the body returned by GET /v1/events.
type ListResponse struct {
	Data    []Event `json:"data"`
	HasMore bool    `json:"hasMore"`
}

// retrieveEnvelope is the {"data": Event} shape used by detail and update endpoints.
type retrieveEnvelope struct {
	Data Event `json:"data"`
}
