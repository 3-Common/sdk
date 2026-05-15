import type { FilterCondition, FilterGroup, FilterRange, Filters } from './types'

/**
 * A {@link FilterGroup} produced by {@link and} / {@link or}, augmented with
 * convenience methods to serialize directly to the wire format. Can be nested
 * inside another `and()` / `or()` because it implements the {@link FilterGroup}
 * shape structurally.
 *
 * @public
 */
export interface SerializableFilter extends FilterGroup {
  /** JSON-string form, ready to assign to the `filters` query param. */
  serialize(): string
  /** Plain JS array form (single-element wrapping this group). */
  toFilters(): Filters
}

/**
 * Reference to a single field on an event (or any resource). Returned by
 * {@link field}; produces a {@link FilterCondition} when an operator method
 * is invoked.
 *
 * @public
 */
export interface FieldRef {
  // Existence
  isEmpty(): FilterCondition
  isNotEmpty(): FilterCondition

  // Equality
  isEqualTo(value: string | number | boolean): FilterCondition
  isNotEqualTo(value: string | number | boolean): FilterCondition

  // Set membership
  isEqualToAnyOf(values: readonly (string | number)[]): FilterCondition
  isNotEqualToAnyOf(values: readonly (string | number)[]): FilterCondition
  isAnyOf(values: readonly (string | number)[]): FilterCondition
  isNoneOf(values: readonly (string | number)[]): FilterCondition

  // Substring
  contains(value: string): FilterCondition
  containsExactly(value: string): FilterCondition

  // Date
  isBefore(value: string): FilterCondition
  isAfter(value: string): FilterCondition

  // Numeric comparison
  isGreaterThan(value: number): FilterCondition
  isGreaterThanOrEqualTo(value: number): FilterCondition
  isLessThan(value: number): FilterCondition
  isLessThanOrEqualTo(value: number): FilterCondition

  // Range
  isBetween(value: FilterRange<string> | FilterRange<number>): FilterCondition
}

class SerializableFilterImpl implements SerializableFilter {
  public readonly logic: 'and' | 'or'
  public readonly conditions: readonly (FilterCondition | FilterGroup)[]

  public constructor(logic: 'and' | 'or', conditions: readonly (FilterCondition | FilterGroup)[]) {
    this.logic = logic
    this.conditions = conditions
  }

  public toFilters(): Filters {
    return [{ logic: this.logic, conditions: this.conditions }]
  }

  public serialize(): string {
    return JSON.stringify(this.toFilters())
  }
}

class CombinedFilters {
  public constructor(private readonly groups: readonly FilterGroup[]) {}

  public toFilters(): Filters {
    return this.groups
  }

  public serialize(): string {
    return JSON.stringify(this.groups)
  }
}

/**
 * Reference a field by name. Operator methods on the result produce a
 * {@link FilterCondition} that can be passed to {@link and} / {@link or} or
 * nested inside other groups.
 *
 * @example
 * ```ts
 * import { filter } from '@3common/sdk'
 *
 * const f = filter.and(
 *   filter.field('status').isAnyOf(['open']),
 *   filter.field('ticketSum').isGreaterThan(10),
 * )
 *
 * await client.events.list({ filters: f.serialize() })
 * ```
 *
 * @public
 */
export function field(name: string): FieldRef {
  if (typeof name !== 'string' || name.length === 0) {
    throw new TypeError('filter.field: name must be a non-empty string')
  }
  const make = (
    operator: FilterCondition['operator'],
    value?: FilterCondition['value'],
  ): FilterCondition =>
    value === undefined ? { field: name, operator } : { field: name, operator, value }
  return {
    isEmpty: () => make('is_empty'),
    isNotEmpty: () => make('is_not_empty'),
    isEqualTo: (v) => make('is_equal_to', v),
    isNotEqualTo: (v) => make('is_not_equal_to', v),
    isEqualToAnyOf: (vs) => make('is_equal_to_any_of', vs),
    isNotEqualToAnyOf: (vs) => make('is_not_equal_to_any_of', vs),
    isAnyOf: (vs) => make('is_any_of', vs),
    isNoneOf: (vs) => make('is_none_of', vs),
    contains: (v) => make('contains', v),
    containsExactly: (v) => make('contains_exactly', v),
    isBefore: (v) => make('is_before', v),
    isAfter: (v) => make('is_after', v),
    isGreaterThan: (v) => make('is_greater_than', v),
    isGreaterThanOrEqualTo: (v) => make('is_greater_than_or_equal_to', v),
    isLessThan: (v) => make('is_less_than', v),
    isLessThanOrEqualTo: (v) => make('is_less_than_or_equal_to', v),
    isBetween: (v) => make('is_between', v),
  }
}

/**
 * Combine conditions / groups with AND logic. Returns a {@link SerializableFilter}
 * that can either be nested in another group or serialized directly.
 *
 * @public
 */
export function and(...conditions: readonly (FilterCondition | FilterGroup)[]): SerializableFilter {
  if (conditions.length === 0) {
    throw new TypeError('filter.and: at least one condition is required')
  }
  return new SerializableFilterImpl('and', conditions)
}

/**
 * Combine conditions / groups with OR logic. Returns a {@link SerializableFilter}
 * that can either be nested in another group or serialized directly.
 *
 * @public
 */
export function or(...conditions: readonly (FilterCondition | FilterGroup)[]): SerializableFilter {
  if (conditions.length === 0) {
    throw new TypeError('filter.or: at least one condition is required')
  }
  return new SerializableFilterImpl('or', conditions)
}

/**
 * Combine multiple top-level groups. The API treats multiple top-level groups
 * as ANDed together. Most callers will only need {@link and} / {@link or} —
 * use this only when you need an explicit array of groups.
 *
 * @public
 */
export function combine(...groups: readonly FilterGroup[]): {
  serialize(): string
  toFilters(): Filters
} {
  if (groups.length === 0) {
    throw new TypeError('filter.combine: at least one group is required')
  }
  return new CombinedFilters(groups)
}
