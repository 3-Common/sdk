/**
 * Public types for the events resource. Hand-curated friendly aliases over
 * the auto-generated OpenAPI types.
 *
 * @public
 */

import type { components, paths } from '@/generated/types'

export type { components }

export type EventStatus =
  | 'draft'
  | 'open'
  | 'closed'
  | 'unpublished'
  | 'cancelled'
  | 'postponed'
  | 'schedule'

/**
 * One event as returned by the API. The server marks every field as optional
 * in list responses (clients can request specific fields via the `fields`
 * param), so the shape is partial. Detail responses populate everything the
 * server has for that record.
 */
export type Event = NonNullable<
  paths['/v1/events/{id}']['get']['responses'][200]['content']['application/json']['data']
>

/** Successful response shape from `GET /v1/events`. */
export interface ListEventsResponse {
  readonly data: readonly Event[]
  readonly hasMore: boolean
}

/** Query parameters accepted by `GET /v1/events`. */
export interface EventListParams {
  /** Page number, 0-indexed. Default 0. */
  readonly page?: number
  /** Items per page. Default 20, max 50. */
  readonly pageSize?: number
  /** Filter by status; omit to include all. */
  readonly status?: EventStatus
  /** Search by name or address (case-insensitive partial match). */
  readonly search?: string
  /** ISO 8601; only events starting on or before this date. */
  readonly startBefore?: string
  /** ISO 8601; only events starting on or after this date. */
  readonly startAfter?: string
  /** Field to sort by. Default `start`. */
  readonly sortField?: string
  /** Sort direction. Default `desc`. */
  readonly sortDirection?: 'asc' | 'desc'
  /**
   * JSON-encoded `FilterGroup[]` for advanced filtering. Use the typed
   * `filter` builder from `@3common/sdk` to construct this — never write the
   * JSON by hand.
   *
   * @example
   * ```ts
   * import { filter } from '@3common/sdk'
   *
   * await client.events.list({
   *   filters: filter.and(
   *     filter.field('status').isAnyOf(['open']),
   *     filter.field('ticketSum').isGreaterThan(10),
   *   ).serialize(),
   * })
   * ```
   */
  readonly filters?: string
  /** Comma-separated list of fields to include in the response. Omit for all fields. */
  readonly fields?: string
}

/** Query parameters accepted by `GET /v1/events/{id}`. */
export interface EventRetrieveParams {
  /** Comma-separated list of fields to include in the response. */
  readonly fields?: string
}

/** Body shape accepted by `PATCH /v1/events/{id}`. Only fields you provide are changed. */
export type EventUpdateBody =
  paths['/v1/events/{id}']['patch']['requestBody']['content']['application/json']
