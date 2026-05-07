/**
 * Wire-level types for the API's `filters` query parameter. Shared across
 * resources — every endpoint that accepts `filters` consumes this same shape.
 *
 * @public
 */

export type FilterLogic = 'and' | 'or'

export type FilterOperator =
  // Common
  | 'is_empty'
  | 'is_not_empty'
  // Numeric / text equality
  | 'is_equal_to'
  | 'is_not_equal_to'
  // Set membership (text)
  | 'is_equal_to_any_of'
  | 'is_not_equal_to_any_of'
  | 'is_any_of'
  | 'is_none_of'
  // Substring
  | 'contains'
  | 'contains_exactly'
  // Date
  | 'is_before'
  | 'is_after'
  // Numeric comparison
  | 'is_greater_than'
  | 'is_greater_than_or_equal_to'
  | 'is_less_than'
  | 'is_less_than_or_equal_to'
  // Range (date or numeric)
  | 'is_between'

/**
 * Range envelope used by `is_between`. The value's components must agree on
 * type (both string ISO dates or both numbers).
 */
export interface FilterRange<T extends string | number> {
  readonly start: T
  readonly end: T
}

/**
 * Every value the wire format accepts. The SDK doesn't enforce per-operator
 * value shape at the type level — that's the API server's job — but the
 * fluent builder constrains values for each operator at compile time.
 */
export type FilterValue =
  | string
  | number
  | boolean
  | readonly (string | number)[]
  | FilterRange<string>
  | FilterRange<number>

/** A single condition: `field operator value?`. */
export interface FilterCondition {
  readonly field: string
  readonly operator: FilterOperator
  readonly value?: FilterValue
}

/** A logical group of conditions or nested groups. */
export interface FilterGroup {
  readonly logic: FilterLogic
  readonly conditions: readonly (FilterCondition | FilterGroup)[]
}

/**
 * Anything the API accepts as the `filters` parameter — i.e. an array of
 * top-level groups.
 */
export type Filters = readonly FilterGroup[]
