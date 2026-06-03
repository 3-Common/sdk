// Package prices provides the prices resource client for the 3Common API.
//
// Most callers reach this package through
// [github.com/3-Common/sdk/sdk-go/client.API.Prices]:
//
//	api, _ := client.New(threecommon.Config{APIKey: "..."})
//	price, err := api.Prices.Create(ctx, &prices.CreateParams{
//		ProductID:  "prod_7",
//		Type:       prices.TypeRecurring,
//		Currency:   prices.CurrencyUSD,
//		UnitAmount: 1500,
//		Recurring:  &prices.Recurring{Interval: prices.IntervalMonth, IntervalCount: 1},
//	})
//
// To use the resource standalone (without the aggregator), call [New]:
//
//	pr, _ := prices.New(threecommon.Config{APIKey: "..."})
//	result, err := pr.List(ctx, nil)
//
// Type names inside this package omit the "Price" prefix to avoid stutter
// (e.g. prices.ListParams, not prices.PriceListParams).
package prices

import "encoding/json"

// Type is the cadence of a price.
type Type string

// Type values returned by the API. Unknown values from a future API version
// will surface as the raw string.
const (
	// TypeRecurring is billed on a fixed cadence (subscription-backed).
	TypeRecurring Type = "recurring"
	// TypeOneTime is a single charge, typically an add-on / top-up pack.
	TypeOneTime Type = "one_time"
)

// Currency is the settlement currency of a price.
type Currency string

// Currency values returned by the API.
const (
	CurrencyUSD Currency = "USD"
	CurrencyCAD Currency = "CAD"
)

// Interval is the cadence unit of a recurring price.
type Interval string

// Interval values returned by the API.
const (
	IntervalDay   Interval = "day"
	IntervalWeek  Interval = "week"
	IntervalMonth Interval = "month"
	IntervalYear  Interval = "year"
)

// Recurring is the cadence descriptor, present when Type is recurring.
type Recurring struct {
	Interval      Interval `json:"interval"`
	IntervalCount int64    `json:"intervalCount"`
}

// FeatureType is the kind of a price feature grant.
type FeatureType string

// FeatureType values returned by the API.
const (
	FeatureTypeBoolean  FeatureType = "boolean"
	FeatureTypeQuantity FeatureType = "quantity"
	FeatureTypeEnum     FeatureType = "enum"
	FeatureTypeDuration FeatureType = "duration"
)

// Feature is one typed feature grant on a price — a tagged union keyed on
// Type. Only the fields relevant to Type are meaningful; set Type and the
// matching fields when building a create/update body. On the wire each variant
// serializes to exactly its own shape.
//
//   - boolean:  Enabled
//   - quantity: Quantity (nil = unlimited), RolloverEnabled, RolloverCap, ExpireOnCancel
//   - enum:     EnumValue
//   - duration: DurationDays (nil = unlimited)
type Feature struct {
	FeatureKey string      `json:"featureKey"`
	Type       FeatureType `json:"type"`

	// Boolean grants.
	Enabled *bool `json:"enabled,omitempty"`

	// Quantity grants. Quantity nil means unlimited.
	Quantity        *int64 `json:"quantity,omitempty"`
	RolloverEnabled *bool  `json:"rolloverEnabled,omitempty"`
	RolloverCap     *int64 `json:"rolloverCap,omitempty"`
	ExpireOnCancel  *bool  `json:"expireOnCancel,omitempty"`

	// Enum grants.
	EnumValue string `json:"enumValue,omitempty"`

	// Duration grants. DurationDays nil means unlimited / "all time".
	DurationDays *int64 `json:"durationDays,omitempty"`
}

// MarshalJSON emits only the fields that belong to the feature's Type, so each
// variant matches the API's discriminated union. Quantity and DurationDays are
// emitted even when nil (as null = unlimited) since the API requires them.
func (f Feature) MarshalJSON() ([]byte, error) {
	switch f.Type {
	case FeatureTypeBoolean:
		return json.Marshal(struct {
			FeatureKey string      `json:"featureKey"`
			Type       FeatureType `json:"type"`
			Enabled    bool        `json:"enabled"`
		}{f.FeatureKey, f.Type, f.Enabled != nil && *f.Enabled})
	case FeatureTypeQuantity:
		return json.Marshal(struct {
			FeatureKey      string      `json:"featureKey"`
			Type            FeatureType `json:"type"`
			Quantity        *int64      `json:"quantity"`
			RolloverEnabled bool        `json:"rolloverEnabled"`
			RolloverCap     *int64      `json:"rolloverCap,omitempty"`
			ExpireOnCancel  *bool       `json:"expireOnCancel,omitempty"`
		}{f.FeatureKey, f.Type, f.Quantity, f.RolloverEnabled != nil && *f.RolloverEnabled, f.RolloverCap, f.ExpireOnCancel})
	case FeatureTypeEnum:
		return json.Marshal(struct {
			FeatureKey string      `json:"featureKey"`
			Type       FeatureType `json:"type"`
			EnumValue  string      `json:"enumValue"`
		}{f.FeatureKey, f.Type, f.EnumValue})
	case FeatureTypeDuration:
		return json.Marshal(struct {
			FeatureKey   string      `json:"featureKey"`
			Type         FeatureType `json:"type"`
			DurationDays *int64      `json:"durationDays"`
		}{f.FeatureKey, f.Type, f.DurationDays})
	default:
		// Unknown/future variant: emit the struct as-is (omitempty applies).
		type featureAlias Feature
		return json.Marshal(featureAlias(f))
	}
}

// Price is the resource shape returned by the API. Pointer fields and
// `omitempty` strings are populated only when the server returned them — list
// responses with a `Fields` filter omit unrequested values.
type Price struct {
	ID         string            `json:"id"`
	HostID     string            `json:"hostId,omitempty"`
	ProductID  string            `json:"productId,omitempty"`
	Type       Type              `json:"type,omitempty"`
	Currency   Currency          `json:"currency,omitempty"`
	UnitAmount *int64            `json:"unitAmount,omitempty"`
	Recurring  *Recurring        `json:"recurring,omitempty"`
	Features   []Feature         `json:"features,omitempty"`
	Nickname   string            `json:"nickname,omitempty"`
	Active     *bool             `json:"active,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  string            `json:"createdAt,omitempty"`
	UpdatedAt  string            `json:"updatedAt,omitempty"`
}

// ListParams are the query parameters accepted by [Client.List] and [Client.ListAutoPaginate].
type ListParams struct {
	// Page is the 0-indexed page number. Nil uses the server default (0).
	Page *int

	// PageSize is the items-per-page cap. Nil uses the server default.
	PageSize *int

	// ProductID filters by parent product.
	ProductID string

	// Type filters by cadence. Empty includes all.
	Type Type

	// Active, when set, returns only active (true) or only archived (false)
	// prices. Nil applies no filter.
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

// CreateParams is the body shape accepted by [Client.Create]. Recurring is
// required when Type is recurring and forbidden when Type is one_time.
type CreateParams struct {
	ProductID  string            `json:"productId"`
	Type       Type              `json:"type"`
	Currency   Currency          `json:"currency"`
	UnitAmount int64             `json:"unitAmount"`
	Recurring  *Recurring        `json:"recurring,omitempty"`
	Features   []Feature         `json:"features,omitempty"`
	Nickname   string            `json:"nickname,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// UpdateParams is the body shape accepted by [Client.Update]. Only fields with
// non-zero/non-nil values are sent; mutable fields are unitAmount, recurring,
// features, nickname, and metadata.
type UpdateParams struct {
	UnitAmount *int64            `json:"unitAmount,omitempty"`
	Recurring  *Recurring        `json:"recurring,omitempty"`
	Features   []Feature         `json:"features,omitempty"`
	Nickname   *string           `json:"nickname,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// ListResponse is the body returned by GET /v1/prices.
type ListResponse struct {
	Data    []Price `json:"data"`
	HasMore bool    `json:"hasMore"`
}

// retrieveEnvelope is the {"data": Price} shape used by the detail, create,
// update, archive, and unarchive endpoints.
type retrieveEnvelope struct {
	Data Price `json:"data"`
}
