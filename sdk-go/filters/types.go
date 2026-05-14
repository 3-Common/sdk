// Package filters provides the typed builder and wire types for the API's
// `filters` query parameter. Every endpoint that accepts filters consumes
// this same shape, so the package is shared across resource packages rather
// than duplicated.
//
// Build filters with [Field], combine groups with [And] / [Or] / [Combine],
// and serialize via [SerializableFilter.Serialize] before passing as the
// `Filters` query argument:
//
//	f := filters.And(
//		filters.Field("status").IsAnyOf("open"),
//		filters.Field("ticketSum").IsGreaterThan(10),
//	)
//	api.Events.List(ctx, &events.ListParams{Filters: f.Serialize()})
package filters

// Logic is the boolean connector for a [Group].
type Logic string

// Logic values.
const (
	LogicAnd Logic = "and"
	LogicOr  Logic = "or"
)

// Operator is the operation applied by a [Condition]. The full set is
// enumerated for compile-time safety; the API server rejects unknown values.
type Operator string

// Operator values supported by the API.
const (
	// Existence
	OpIsEmpty    Operator = "is_empty"
	OpIsNotEmpty Operator = "is_not_empty"

	// Equality
	OpIsEqualTo    Operator = "is_equal_to"
	OpIsNotEqualTo Operator = "is_not_equal_to"

	// Set membership
	OpIsEqualToAnyOf    Operator = "is_equal_to_any_of"
	OpIsNotEqualToAnyOf Operator = "is_not_equal_to_any_of"
	OpIsAnyOf           Operator = "is_any_of"
	OpIsNoneOf          Operator = "is_none_of"

	// Substring
	OpContains        Operator = "contains"
	OpContainsExactly Operator = "contains_exactly"

	// Date
	OpIsBefore Operator = "is_before"
	OpIsAfter  Operator = "is_after"

	// Numeric comparison
	OpIsGreaterThan          Operator = "is_greater_than"
	OpIsGreaterThanOrEqualTo Operator = "is_greater_than_or_equal_to"
	OpIsLessThan             Operator = "is_less_than"
	OpIsLessThanOrEqualTo    Operator = "is_less_than_or_equal_to"

	// Range (date or numeric)
	OpIsBetween Operator = "is_between"
)

// Range is the value type for [OpIsBetween]. Start and End must agree on
// type — both numbers or both ISO date strings.
type Range struct {
	Start any `json:"start"`
	End   any `json:"end"`
}

// Condition is a single field/operator/value triple. Value may be omitted for
// existence operators ([OpIsEmpty], [OpIsNotEmpty]).
type Condition struct {
	Field    string   `json:"field"`
	Operator Operator `json:"operator"`
	Value    any      `json:"value,omitempty"`
}

// Group is a set of conditions or sub-groups joined by [Logic].
type Group struct {
	Logic      Logic `json:"logic"`
	Conditions []any `json:"conditions"` // each element is a *Condition or *Group
}
