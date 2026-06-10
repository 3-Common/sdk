// Package features provides the features resource client for the 3Common API.
//
// Most callers reach this package through
// [github.com/3-Common/sdk/sdk-go/client.API.Features]:
//
//	api, _ := client.New(threecommon.Config{APIKey: "..."})
//	feature, err := api.Features.Create(ctx, &features.CreateParams{
//		Key:  "api_calls",
//		Name: "API calls",
//		Type: features.TypeQuantity,
//	})
//
// To use the resource standalone (without the aggregator), call [New]:
//
//	fs, _ := features.New(threecommon.Config{APIKey: "..."})
//	result, err := fs.List(ctx, nil)
//
// Type names inside this package omit the "Feature" prefix to avoid stutter
// (e.g. features.ListParams, not features.FeatureListParams).
package features

// Type is the value shape of a feature.
type Type string

// Type values returned by the API. Unknown values from a future API version
// will surface as the raw string.
const (
	// TypeBoolean is a pure on/off feature.
	TypeBoolean Type = "boolean"
	// TypeQuantity is a countable feature; drives entitlement balance.
	TypeQuantity Type = "quantity"
	// TypeEnum is one of a fixed ordered set of values.
	TypeEnum Type = "enum"
	// TypeDuration is a number of days (or unlimited).
	TypeDuration Type = "duration"
)

// Feature is one entry in the host's feature catalog. Pointer fields and
// `omitempty` strings are populated only when the server returned them — list
// responses with a `Fields` filter omit unrequested values.
type Feature struct {
	ID          string            `json:"id"`
	HostID      string            `json:"hostId,omitempty"`
	Key         string            `json:"key,omitempty"`
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Type        Type              `json:"type,omitempty"`
	EnumValues  []string          `json:"enumValues,omitempty"`
	Active      *bool             `json:"active,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   string            `json:"createdAt,omitempty"`
	UpdatedAt   string            `json:"updatedAt,omitempty"`
}

// ResolvedValue is the resolved type-specific value of a feature for a
// customer. Only the fields relevant to Type are populated:
//
//   - boolean:  Enabled
//   - quantity: Quantity (nil = unlimited), Balance (live entitlement balance, if any)
//   - enum:     EnumValue
//   - duration: DurationDays (nil = unlimited)
type ResolvedValue struct {
	Type         Type    `json:"type"`
	Enabled      *bool   `json:"enabled,omitempty"`
	Quantity     *int64  `json:"quantity,omitempty"`
	Balance      *int64  `json:"balance,omitempty"`
	EnumValue    *string `json:"enumValue,omitempty"`
	DurationDays *int64  `json:"durationDays,omitempty"`
}

// ResolvedFeature is the resolved state of a feature for a customer, returned
// by GET /v1/features/resolve. It combines the catalog feature, the resolved
// value, and the subscriptions that contributed it.
type ResolvedFeature struct {
	Feature                     Feature       `json:"feature"`
	Value                       ResolvedValue `json:"value"`
	ContributingSubscriptionIDs []string      `json:"contributingSubscriptionIds"`
}

// ListParams are the query parameters accepted by [Client.List] and [Client.ListAutoPaginate].
type ListParams struct {
	// Page is the 0-indexed page number. Nil uses the server default (0).
	Page *int

	// PageSize is the items-per-page cap. Nil uses the server default.
	PageSize *int

	// Type filters by value shape. Empty includes all.
	Type Type

	// Active, when set, returns only active (true) or only archived (false)
	// features. Nil applies no filter.
	Active *bool

	// Fields is a comma-separated list of fields to include in the response.
	// Empty returns all fields.
	Fields string
}

// RetrieveParams are the query parameters accepted by [Client.Retrieve].
type RetrieveParams struct {
	// Fields is a comma-separated list of fields to include in the response.
	Fields string
}

// ResolveParams are the query parameters accepted by [Client.Resolve].
// ContactID and FeatureKey are required.
type ResolveParams struct {
	// ContactID is the CRM contact id.
	ContactID string

	// FeatureKey is the feature catalog key.
	FeatureKey string
}

// CreateParams is the body shape accepted by [Client.Create].
type CreateParams struct {
	Key         string            `json:"key"`
	Name        string            `json:"name"`
	Type        Type              `json:"type"`
	Description string            `json:"description,omitempty"`
	EnumValues  []string          `json:"enumValues,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// UpdateParams is the body shape accepted by [Client.Update]. Mutable fields:
// Name, Description, EnumValues, Metadata. Key and Type are immutable.
type UpdateParams struct {
	Name        *string           `json:"name,omitempty"`
	Description *string           `json:"description,omitempty"`
	EnumValues  []string          `json:"enumValues,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ListResponse is the body returned by GET /v1/features.
type ListResponse struct {
	Data    []Feature `json:"data"`
	HasMore bool      `json:"hasMore"`
}

// retrieveEnvelope is the {"data": Feature} shape used by the detail, create,
// update, archive, and unarchive endpoints.
type retrieveEnvelope struct {
	Data Feature `json:"data"`
}

// resolveEnvelope is the {"data": ResolvedFeature} shape returned by
// GET /v1/features/resolve.
type resolveEnvelope struct {
	Data ResolvedFeature `json:"data"`
}
