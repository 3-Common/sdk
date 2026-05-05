package filters_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/3-Common/sdk/sdk-go/filters"
)

func TestField_OperatorMethodsProduceConditions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		got  *filters.Condition
		want filters.Operator
	}{
		{"is_empty", filters.Field("x").IsEmpty(), filters.OpIsEmpty},
		{"is_not_empty", filters.Field("x").IsNotEmpty(), filters.OpIsNotEmpty},
		{"is_equal_to", filters.Field("x").IsEqualTo(1), filters.OpIsEqualTo},
		{"is_not_equal_to", filters.Field("x").IsNotEqualTo(1), filters.OpIsNotEqualTo},
		{"is_equal_to_any_of", filters.Field("x").IsEqualToAnyOf(1, 2), filters.OpIsEqualToAnyOf},
		{"is_not_equal_to_any_of", filters.Field("x").IsNotEqualToAnyOf(1), filters.OpIsNotEqualToAnyOf},
		{"is_any_of", filters.Field("x").IsAnyOf("a", "b"), filters.OpIsAnyOf},
		{"is_none_of", filters.Field("x").IsNoneOf("a"), filters.OpIsNoneOf},
		{"contains", filters.Field("x").Contains("a"), filters.OpContains},
		{"contains_exactly", filters.Field("x").ContainsExactly("a"), filters.OpContainsExactly},
		{"is_before", filters.Field("x").IsBefore("2026-01-01"), filters.OpIsBefore},
		{"is_after", filters.Field("x").IsAfter("2026-01-01"), filters.OpIsAfter},
		{"is_greater_than", filters.Field("x").IsGreaterThan(5), filters.OpIsGreaterThan},
		{"is_greater_or_eq", filters.Field("x").IsGreaterThanOrEqualTo(5), filters.OpIsGreaterThanOrEqualTo},
		{"is_less_than", filters.Field("x").IsLessThan(5), filters.OpIsLessThan},
		{"is_less_or_eq", filters.Field("x").IsLessThanOrEqualTo(5), filters.OpIsLessThanOrEqualTo},
		{"is_between", filters.Field("x").IsBetween(filters.Range{Start: 1, End: 10}), filters.OpIsBetween},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, "x", tc.got.Field)
			assert.Equal(t, tc.want, tc.got.Operator)
		})
	}
}

func TestField_PanicsOnEmptyName(t *testing.T) {
	t.Parallel()

	assert.PanicsWithValue(t, "filters: field name must not be empty", func() {
		filters.Field("")
	})
}

func TestAnd_SerializesToWireFormat(t *testing.T) {
	t.Parallel()

	f := filters.And(
		filters.Field("status").IsAnyOf("open"),
		filters.Field("ticketSum").IsGreaterThan(10),
	)

	got, err := f.Serialize()
	require.NoError(t, err)

	var parsed []map[string]any
	require.NoError(t, json.Unmarshal([]byte(got), &parsed))
	require.Len(t, parsed, 1)
	assert.Equal(t, "and", parsed[0]["logic"])

	conds := parsed[0]["conditions"].([]any)
	assert.Len(t, conds, 2)
}

func TestOr_AcceptsNestedGroups(t *testing.T) {
	t.Parallel()

	inner := filters.And(
		filters.Field("ticketSum").IsGreaterThan(0),
	)
	outer := filters.Or(
		filters.Field("status").IsEqualTo("open"),
		&inner.Group,
	)

	got := outer.MustSerialize()
	assert.Contains(t, got, `"logic":"or"`)
	assert.Contains(t, got, `"logic":"and"`) // nested
}

func TestAnd_PanicsOnEmpty(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() { _ = filters.And() })
	assert.Panics(t, func() { _ = filters.Or() })
}

func TestAnd_DefensiveCopiesArguments(t *testing.T) {
	t.Parallel()

	conds := []any{filters.Field("a").IsEmpty()}
	f := filters.And(conds...)

	// Mutating the original slice must not affect the filter.
	conds[0] = filters.Field("z").IsEmpty()

	got := f.MustSerialize()
	assert.Contains(t, got, `"field":"a"`)
	assert.NotContains(t, got, `"field":"z"`)
}

func TestCombine_JoinsTopLevelGroups(t *testing.T) {
	t.Parallel()

	a := filters.And(filters.Field("status").IsEqualTo("open"))
	b := filters.Or(filters.Field("type").IsAnyOf("event"))

	c := filters.Combine(&a.Group, &b.Group)
	got, err := c.Serialize()
	require.NoError(t, err)

	var parsed []map[string]any
	require.NoError(t, json.Unmarshal([]byte(got), &parsed))
	assert.Len(t, parsed, 2)
}

func TestCombine_PanicsOnEmpty(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() { _ = filters.Combine() })
}

func TestCombinedFilters_MustSerializeReturnsOnSuccess(t *testing.T) {
	t.Parallel()

	a := filters.And(filters.Field("status").IsEqualTo("open"))
	c := filters.Combine(&a.Group)
	got := c.MustSerialize()
	assert.NotEmpty(t, got)
	assert.Contains(t, got, `"logic":"and"`)
}

func TestMustSerialize_PanicsOnFailure(t *testing.T) {
	t.Parallel()

	// Inject a non-marshalable value (channel) into a condition. JSON
	// marshal will fail, MustSerialize must panic.
	f := filters.And(filters.Field("x").IsEqualTo(make(chan int)))
	assert.Panics(t, func() { _ = f.MustSerialize() })

	c := filters.Combine(&filters.And(filters.Field("x").IsEqualTo(make(chan int))).Group)
	assert.Panics(t, func() { _ = c.MustSerialize() })
}

func TestSerialize_ReturnsErrorOnNonMarshalable(t *testing.T) {
	t.Parallel()

	f := filters.And(filters.Field("x").IsEqualTo(make(chan int)))
	_, err := f.Serialize()
	assert.Error(t, err)

	c := filters.Combine(&filters.And(filters.Field("x").IsEqualTo(make(chan int))).Group)
	_, err = c.Serialize()
	assert.Error(t, err)
}
