// Package contacts provides the contacts resource client for the 3Common API.
//
// Most callers reach this package through
// [github.com/3-Common/sdk/sdk-go/client.API.Contacts]:
//
//	api, _ := client.New(threecommon.Config{APIKey: "..."})
//	page, err := api.Contacts.List(ctx, &contacts.ListParams{
//		Filter:   contacts.QuickFilterOptedIn,
//		PageSize: threecommon.Int(50),
//	})
//
// To use the resource standalone (without the aggregator), call [New]:
//
//	cli, _ := contacts.New(threecommon.Config{APIKey: "..."})
//	page, err := cli.List(ctx, nil)
//
// Type names inside this package omit the "Contact" prefix to avoid stutter
// (e.g. contacts.ListParams, not contacts.ContactListParams).
package contacts

import "github.com/3-Common/sdk/sdk-go/filters"

// Status is the lifecycle status of a contact.
//
//   - StatusOptedIn / StatusUnsubscribed: explicit consent state
//   - StatusUnknown: never recorded a choice
//   - StatusImported: created via CSV / bulk-upsert before consent was captured
//   - StatusDeleted: soft-deleted
type Status string

// Status values returned by the API. Unknown values from a future API
// version will surface as the raw string.
const (
	StatusDeleted      Status = "deleted"
	StatusImported     Status = "imported"
	StatusUnsubscribed Status = "unsubscribed"
	StatusOptedIn      Status = "opted-in"
	StatusUnknown      Status = "unknown"
)

// MergeResolution chooses how field-level conflicts are resolved when
// [UpdateParams.MergeWith] is set.
type MergeResolution string

// MergeResolution values.
const (
	// MergeResolutionSafe only fills fields that are empty on the target.
	MergeResolutionSafe MergeResolution = "safe-merge"
	// MergeResolutionOverwrite replaces target fields with source fields.
	MergeResolutionOverwrite MergeResolution = "overwrite-merge"
)

// QuickFilter is the simple status filter accepted by [ListParams.Filter].
type QuickFilter string

// QuickFilter values.
const (
	QuickFilterAll          QuickFilter = "all"
	QuickFilterOptedIn      QuickFilter = "opted-in"
	QuickFilterUnknown      QuickFilter = "unknown"
	QuickFilterUnsubscribed QuickFilter = "unsubscribed"
	QuickFilterImported     QuickFilter = "imported"
)

// ActivityType is the kind of event recorded against a contact in their
// activity feed.
type ActivityType string

// ActivityType values.
const (
	ActivityCheckoutSessionCompleted           ActivityType = "checkout_session_completed"
	ActivityProductSetCheckoutSessionCompleted ActivityType = "product_set_checkout_session_completed"
	ActivityOrderRefunded                      ActivityType = "order_refunded"
	ActivityTicketScanned                      ActivityType = "ticket_scanned"
	ActivityEmailSent                          ActivityType = "email_sent"
	ActivityInvoicePaid                        ActivityType = "invoice_paid"
)

// Contact is the resource shape returned by List, Retrieve, and Create
// (the "compact" projection). Custom-property keys (24-char hex ids) may
// appear in CustomProperties.
type Contact struct {
	ID                   string   `json:"id"`
	FirstName            string   `json:"firstName"`
	LastName             string   `json:"lastName"`
	FullName             string   `json:"fullName"`
	Email                string   `json:"email"`
	Phone                string   `json:"phone,omitempty"`
	VendorID             string   `json:"vendorId"`
	OrderSum             int64    `json:"orderSum"`
	GrossSum             int64    `json:"grossSum"`
	FirstOrder           *int64   `json:"firstOrder,omitempty"`
	LastOrder            *int64   `json:"lastOrder,omitempty"`
	CreatedAt            string   `json:"createdAt,omitempty"`
	Status               Status   `json:"status"`
	EventsAttendedIDs    []string `json:"eventsAttended_IDS"`
	ItemsPurchasedIDs    []string `json:"itemsPurchased_IDS"`
	ProductsPurchasedIDs []string `json:"productsPurchased_IDS"`
}

// Property is one custom-property entry on the richer order-details
// projection.
type Property struct {
	PropertyID string `json:"property_id"`
	// Value can be a string, []string, or bool depending on the property's
	// configured type. Decoded as the matching Go variant; pass the right
	// shape when constructing.
	Value any `json:"value"`
}

// WithOrderDetails is the richer projection returned by Update. Includes
// raw events_attended / items_purchased / products_purchased arrays plus
// the properties array, on top of everything in [Contact]. The id field on
// this projection is `_id` (Mongo-style), not `id`.
type WithOrderDetails struct {
	ID                string     `json:"_id"`
	Email             string     `json:"email"`
	VendorID          string     `json:"vendorId"`
	FirstName         string     `json:"firstName"`
	LastName          string     `json:"lastName"`
	FullName          string     `json:"fullName"`
	Phone             *string    `json:"phone,omitempty"`
	Status            Status     `json:"status"`
	GrossSum          int64      `json:"grossSum"`
	OrderSum          int64      `json:"orderSum"`
	LeastRecentOrder  string     `json:"leastRecentOrder,omitempty"`
	MostRecentOrder   string     `json:"mostRecentOrder,omitempty"`
	EventsAttended    []string   `json:"events_attended"`
	ItemsPurchased    []string   `json:"items_purchased"`
	ProductsPurchased []string   `json:"products_purchased"`
	Properties        []Property `json:"properties,omitempty"`
	CreatedAt         string     `json:"createdAt,omitempty"`
	UpdatedAt         string     `json:"updatedAt,omitempty"`
}

// Activity is a single activity record in a contact's activity feed.
type Activity struct {
	ID        string         `json:"_id"`
	VendorID  string         `json:"vendor_id"`
	Email     string         `json:"email"`
	ContactID string         `json:"contact_id,omitempty"`
	Type      ActivityType   `json:"type"`
	Data      map[string]any `json:"data"`
	CreatedAt string         `json:"createdAt"`
	UpdatedAt string         `json:"updatedAt"`
}

// ListParams are the query parameters accepted by [Client.List] and
// [Client.ListAutoPaginate].
type ListParams struct {
	// PageNumber is the 0-indexed page number. Nil uses the server default (0).
	PageNumber *int

	// PageSize is the items-per-page cap. Nil uses the server default (20);
	// the server caps at 500.
	PageSize *int

	// SortField is the field to sort by. Index-backed values:
	// `mostRecentOrder`, `orderSum`, `grossSum`. Other field names and
	// 24-char hex custom-property ids are accepted and sorted in-memory.
	// Defaults to `mostRecentOrder`.
	SortField string

	// SortDirection is "asc" or "desc". Defaults to "desc" when SortField
	// is provided.
	SortDirection string

	// Filter is a quick status filter. Empty (or `"all"`) includes all
	// statuses. ANDed with `Filters` when both are supplied.
	Filter QuickFilter

	// Filters is a JSON-encoded filter group produced by
	// [github.com/3-Common/sdk/sdk-go/filters.SerializableFilter.Serialize].
	Filters string

	// Search is a free-text query over email, firstName, lastName, fullName.
	Search string
}

// FilterWith is a convenience that serializes f and assigns it to
// [ListParams.Filters]. Returns the modified params for chaining.
//
//	api.Contacts.List(ctx, (&contacts.ListParams{Filter: contacts.QuickFilterOptedIn}).
//		FilterWith(filters.And(filters.Field("grossSum").IsGreaterThan(1000))))
func (p *ListParams) FilterWith(f *filters.SerializableFilter) *ListParams {
	if f != nil {
		p.Filters = f.MustSerialize()
	}
	return p
}

// ActivityListParams are the query parameters accepted by
// [Client.ListActivity] and [Client.ListActivityAutoPaginate].
type ActivityListParams struct {
	// PageNumber is the 0-indexed page number. Nil uses the server default (0).
	PageNumber *int

	// PageSize is the items-per-page cap. Nil uses the server default (20).
	PageSize *int

	// Filter restricts to a single activity type.
	Filter ActivityType

	// Sort is "oldest" to reverse the default newest-first order.
	Sort string
}

// CreateParams is the body shape accepted by [Client.Create].
type CreateParams struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

// ContactUpdate is the nested object inside [UpdateParams]. All four
// non-phone fields are required by the API.
type ContactUpdate struct {
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Email     string  `json:"email"`
	Phone     *string `json:"phone,omitempty"`
	Status    Status  `json:"status"`
}

// UpdateParams is the body shape accepted by [Client.Update]. The Contact
// object carries the new field values; MergeWith + Resolution are set
// together when an email change collides with another existing contact.
type UpdateParams struct {
	Contact    ContactUpdate   `json:"contact"`
	MergeWith  string          `json:"mergeWith,omitempty"`
	Resolution MergeResolution `json:"resolution,omitempty"`
}

// BulkUpsertItem is one row in [BulkUpsertParams.Contacts]. Wider than
// [CreateParams] to support CSV-import flows that carry status, custom
// properties, and the association arrays.
type BulkUpsertItem struct {
	Email                string     `json:"email"`
	FirstName            string     `json:"firstName,omitempty"`
	LastName             string     `json:"lastName,omitempty"`
	Phone                *string    `json:"phone,omitempty"`
	Status               Status     `json:"status,omitempty"`
	Properties           []Property `json:"properties,omitempty"`
	EventsAttendedIDs    []string   `json:"eventsAttended_IDS,omitempty"`
	ItemsPurchasedIDs    []string   `json:"itemsPurchased_IDS,omitempty"`
	ProductsPurchasedIDs []string   `json:"productsPurchased_IDS,omitempty"`
}

// BulkUpsertParams is the body shape accepted by [Client.BulkUpsert].
type BulkUpsertParams struct {
	Contacts []BulkUpsertItem `json:"contacts"`
}

// ListResponse is the body returned by GET /v1/contacts.
type ListResponse struct {
	Data       []Contact `json:"data"`
	HasMore    bool      `json:"hasMore"`
	PageNumber int       `json:"pageNumber"`
	PageSize   int       `json:"pageSize"`
}

// ListActivityResponse is the body returned by GET /v1/contacts/{id}/activity.
type ListActivityResponse struct {
	Data       []Activity `json:"data"`
	HasMore    bool       `json:"hasMore"`
	PageNumber int        `json:"pageNumber"`
	PageSize   int        `json:"pageSize"`
}

// CountResult is the data shape unwrapped from GET /v1/contacts/count.
type CountResult struct {
	Count int64 `json:"count"`
}

// BulkUpsertResult is the data shape unwrapped from POST /v1/contacts/bulk.
type BulkUpsertResult struct {
	Affected int64 `json:"affected"`
}

// DeleteResult is the data shape unwrapped from DELETE /v1/contacts/{id}.
// Echoes the removed contact's id.
type DeleteResult struct {
	ID string `json:"id"`
}

// retrieveEnvelope is the {"data": Contact} shape used by detail-returning endpoints.
type retrieveEnvelope struct {
	Data Contact `json:"data"`
}

// orderDetailsEnvelope is the {"data": WithOrderDetails} shape used by Update.
type orderDetailsEnvelope struct {
	Data WithOrderDetails `json:"data"`
}

// countEnvelope is the {"data": {count}} shape used by Count.
type countEnvelope struct {
	Data CountResult `json:"data"`
}

// bulkUpsertEnvelope is the {"data": {affected}} shape used by BulkUpsert.
type bulkUpsertEnvelope struct {
	Data BulkUpsertResult `json:"data"`
}

// deleteEnvelope is the {"data": {id}} shape used by Delete.
type deleteEnvelope struct {
	Data DeleteResult `json:"data"`
}
