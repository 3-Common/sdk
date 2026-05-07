/**
 * Typed builder for the `filters` query parameter. Shared by every resource
 * that accepts `filters`.
 *
 * @example
 * ```ts
 * import { filter } from '@3-common/sdk'
 *
 * const f = filter.and(
 *   filter.field('status').isAnyOf(['open', 'closed']),
 *   filter.or(
 *     filter.field('ticketSum').isGreaterThan(10),
 *     filter.field('revenueCents').isGreaterThan(10000),
 *   ),
 * )
 *
 * await client.events.list({ filters: f.serialize() })
 * ```
 *
 * @public
 */

import { and, combine, field, or } from './builder'

export { and, combine, field, or } from './builder'
export type { FieldRef, SerializableFilter } from './builder'
export type {
  FilterCondition,
  FilterGroup,
  FilterLogic,
  FilterOperator,
  FilterRange,
  FilterValue,
  Filters,
} from './types'

/**
 * Namespace export — `filter.field(...)`, `filter.and(...)`, `filter.or(...)`,
 * `filter.combine(...)`. Equivalent to importing the named functions
 * individually; provided for ergonomic discovery.
 *
 * @public
 */
export const filter = { field, and, or, combine } as const
