// Package entitlements provides the entitlements resource client for the
// 3Common API.
//
// Most callers reach this package through
// [github.com/3-Common/sdk/sdk-go/client.API.Entitlements]:
//
//	api, _ := client.New(threecommon.Config{APIKey: "..."})
//	ent, err := api.Entitlements.Lookup(ctx, &entitlements.LookupParams{
//		ContactID:  "cnt_7",
//		FeatureKey: "api_calls",
//	})
//
// To use the resource standalone (without the aggregator), call [New]:
//
//	ents, _ := entitlements.New(threecommon.Config{APIKey: "..."})
//	result, err := ents.List(ctx, nil)
//
// Type names inside this package omit the "Entitlement" prefix to avoid
// stutter (e.g. entitlements.ListParams, not entitlements.EntitlementListParams).
package entitlements

// GrantSource is the origin of an entitlement grant.
type GrantSource string

// GrantSource values returned by the API. Unknown values from a future API
// version will surface as the raw string.
const (
	// GrantSourceSubscriptionRecurring is a cycle grant from a subscription renewal.
	GrantSourceSubscriptionRecurring GrantSource = "subscription_recurring"
	// GrantSourceOneTimeAddon is a top-up purchase (consumed first by Consume).
	GrantSourceOneTimeAddon GrantSource = "one_time_addon"
	// GrantSourceManual is an admin-applied grant.
	GrantSourceManual GrantSource = "manual"
)

// Grant is one entry in an entitlement's grant history.
type Grant struct {
	ID        string      `json:"id"`
	Source    GrantSource `json:"source"`
	SourceID  string      `json:"sourceId,omitempty"`
	PriceID   string      `json:"priceId,omitempty"`
	Amount    int64       `json:"amount"`
	Remaining int64       `json:"remaining"`
	AddedAt   string      `json:"addedAt"`
}

// Entitlement is the resource shape returned by the API. Pointer fields and
// `omitempty` strings are populated only when the server returned them — list
// responses with a `Fields` filter omit unrequested values.
type Entitlement struct {
	ID            string            `json:"id"`
	HostID        string            `json:"hostId,omitempty"`
	ContactID     string            `json:"contactId,omitempty"`
	FeatureKey    string            `json:"featureKey,omitempty"`
	Balance       *int64            `json:"balance,omitempty"`
	Grants        []Grant           `json:"grants,omitempty"`
	TotalGranted  *int64            `json:"totalGranted,omitempty"`
	TotalConsumed *int64            `json:"totalConsumed,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	CreatedAt     string            `json:"createdAt,omitempty"`
	UpdatedAt     string            `json:"updatedAt,omitempty"`
}

// ListParams are the query parameters accepted by [Client.List] and [Client.ListAutoPaginate].
type ListParams struct {
	// Page is the 0-indexed page number. Nil uses the server default (0).
	Page *int

	// PageSize is the items-per-page cap. Nil uses the server default.
	PageSize *int

	// ContactID filters by recipient contact id.
	ContactID string

	// FeatureKey filters by feature.
	FeatureKey string

	// MinBalance keeps only entitlements whose balance is >= this value. Nil
	// applies no floor.
	MinBalance *int64

	// Fields is a comma-separated list of fields to include in the response.
	// Empty returns all fields.
	Fields string
}

// RetrieveParams are the query parameters accepted by [Client.Retrieve].
type RetrieveParams struct {
	// Fields is a comma-separated list of fields to include in the response.
	Fields string
}

// LookupParams are the query parameters accepted by [Client.Lookup]. ContactID
// and FeatureKey are required.
type LookupParams struct {
	// ContactID is the CRM contact id.
	ContactID string

	// FeatureKey is the feature key.
	FeatureKey string

	// Fields is a comma-separated list of fields to include in the response.
	Fields string
}

// GrantParams is the body shape accepted by [Client.Grant].
type GrantParams struct {
	ContactID  string            `json:"contactId"`
	FeatureKey string            `json:"featureKey"`
	Amount     int64             `json:"amount"`
	GrantID    string            `json:"grantId"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// ConsumeParams is the body shape accepted by [Client.Consume].
type ConsumeParams struct {
	ContactID  string `json:"contactId"`
	FeatureKey string `json:"featureKey"`
	Amount     int64  `json:"amount"`
	Reason     string `json:"reason,omitempty"`
}

// ListResponse is the body returned by GET /v1/entitlements.
type ListResponse struct {
	Data    []Entitlement `json:"data"`
	HasMore bool          `json:"hasMore"`
}

// retrieveEnvelope is the {"data": Entitlement} shape used by the detail,
// lookup, grant, and consume endpoints.
type retrieveEnvelope struct {
	Data Entitlement `json:"data"`
}
