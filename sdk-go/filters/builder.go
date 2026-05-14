package filters

import "encoding/json"

// FieldRef is the result of [Field]. Each operator method on a FieldRef
// produces a [*Condition].
type FieldRef struct {
	name string
}

// Field references a field by name. Operator methods on the result produce a
// [*Condition] suitable for use inside [And] / [Or] / [Combine].
//
// Panics if name is empty.
func Field(name string) FieldRef {
	if name == "" {
		panic("filters: field name must not be empty")
	}
	return FieldRef{name: name}
}

// IsEmpty asserts the field has no value.
func (f FieldRef) IsEmpty() *Condition { return f.cond(OpIsEmpty, nil) }

// IsNotEmpty asserts the field has a value.
func (f FieldRef) IsNotEmpty() *Condition { return f.cond(OpIsNotEmpty, nil) }

// IsEqualTo asserts equality.
func (f FieldRef) IsEqualTo(v any) *Condition { return f.cond(OpIsEqualTo, v) }

// IsNotEqualTo asserts inequality.
func (f FieldRef) IsNotEqualTo(v any) *Condition { return f.cond(OpIsNotEqualTo, v) }

// IsEqualToAnyOf asserts the field equals any of the given values.
func (f FieldRef) IsEqualToAnyOf(values ...any) *Condition {
	return f.cond(OpIsEqualToAnyOf, values)
}

// IsNotEqualToAnyOf asserts the field equals none of the given values.
func (f FieldRef) IsNotEqualToAnyOf(values ...any) *Condition {
	return f.cond(OpIsNotEqualToAnyOf, values)
}

// IsAnyOf is the strict-equality variant of [FieldRef.IsEqualToAnyOf].
func (f FieldRef) IsAnyOf(values ...any) *Condition { return f.cond(OpIsAnyOf, values) }

// IsNoneOf is the strict-equality variant of [FieldRef.IsNotEqualToAnyOf].
func (f FieldRef) IsNoneOf(values ...any) *Condition { return f.cond(OpIsNoneOf, values) }

// Contains asserts the field's string value contains v.
func (f FieldRef) Contains(v string) *Condition { return f.cond(OpContains, v) }

// ContainsExactly is the case-sensitive variant of [FieldRef.Contains].
func (f FieldRef) ContainsExactly(v string) *Condition { return f.cond(OpContainsExactly, v) }

// IsBefore asserts the field's date is strictly before v (ISO 8601).
func (f FieldRef) IsBefore(v string) *Condition { return f.cond(OpIsBefore, v) }

// IsAfter asserts the field's date is strictly after v (ISO 8601).
func (f FieldRef) IsAfter(v string) *Condition { return f.cond(OpIsAfter, v) }

// IsGreaterThan asserts strict numeric > v.
func (f FieldRef) IsGreaterThan(v float64) *Condition { return f.cond(OpIsGreaterThan, v) }

// IsGreaterThanOrEqualTo asserts numeric ≥ v.
func (f FieldRef) IsGreaterThanOrEqualTo(v float64) *Condition {
	return f.cond(OpIsGreaterThanOrEqualTo, v)
}

// IsLessThan asserts strict numeric < v.
func (f FieldRef) IsLessThan(v float64) *Condition { return f.cond(OpIsLessThan, v) }

// IsLessThanOrEqualTo asserts numeric ≤ v.
func (f FieldRef) IsLessThanOrEqualTo(v float64) *Condition {
	return f.cond(OpIsLessThanOrEqualTo, v)
}

// IsBetween asserts the field falls within r (inclusive). r.Start and r.End
// must agree on type.
func (f FieldRef) IsBetween(r Range) *Condition { return f.cond(OpIsBetween, r) }

func (f FieldRef) cond(op Operator, v any) *Condition {
	return &Condition{Field: f.name, Operator: op, Value: v}
}

// SerializableFilter is the result of [And] / [Or]. It satisfies the [*Group]
// shape (so it can be nested) and adds JSON-string convenience methods.
type SerializableFilter struct {
	Group
}

// Serialize returns the JSON-string form, ready to assign to a request's `Filters` query argument.
func (s *SerializableFilter) Serialize() (string, error) {
	groups := []*Group{&s.Group}
	out, err := json.Marshal(groups)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// MustSerialize is the panic-on-error variant of [SerializableFilter.Serialize].
// Safe to use with builder-produced filters; reserve [SerializableFilter.Serialize]
// for cases where you accept user-supplied conditions.
func (s *SerializableFilter) MustSerialize() string {
	out, err := s.Serialize()
	if err != nil {
		panic(err)
	}
	return out
}

// And combines conditions and/or sub-groups with AND logic. At least one
// argument is required; the function panics with an empty argument list to
// catch programmer errors at construction time.
func And(items ...any) *SerializableFilter {
	return newSerializable(LogicAnd, items)
}

// Or combines conditions and/or sub-groups with OR logic.
func Or(items ...any) *SerializableFilter {
	return newSerializable(LogicOr, items)
}

// CombinedFilters is the result of [Combine]: an array of top-level groups
// joined implicitly with AND. Use [And] / [Or] instead unless you specifically
// need a multi-group payload.
type CombinedFilters struct {
	groups []*Group
}

// Serialize marshals the combined groups.
func (c *CombinedFilters) Serialize() (string, error) {
	out, err := json.Marshal(c.groups)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// MustSerialize is the panic-on-error variant.
func (c *CombinedFilters) MustSerialize() string {
	out, err := c.Serialize()
	if err != nil {
		panic(err)
	}
	return out
}

// Combine joins multiple top-level groups.
func Combine(groups ...*Group) *CombinedFilters {
	if len(groups) == 0 {
		panic("filters: Combine requires at least one group")
	}
	return &CombinedFilters{groups: groups}
}

func newSerializable(logic Logic, items []any) *SerializableFilter {
	if len(items) == 0 {
		panic("filters: at least one condition or group is required")
	}
	// Defensive copy so callers mutating their slice can't surprise us.
	cp := make([]any, len(items))
	copy(cp, items)
	return &SerializableFilter{Group: Group{Logic: logic, Conditions: cp}}
}
