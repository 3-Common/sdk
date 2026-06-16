/**
 * Public types for the properties resource. Hand-curated friendly aliases over
 * the auto-generated OpenAPI types.
 *
 * @public
 */

import type { paths } from '@/generated/types'

/**
 * One property as returned by the API. A discriminated union on `type`: the
 * `Select One` and `Select Multiple` variants additionally carry an `options`
 * array; every other variant shares the same base shape.
 */
export type Property = NonNullable<
  paths['/v1/properties/{id}']['get']['responses'][200]['content']['application/json']['data']
>

/**
 * The data type of a property.
 *
 * One of `Text`, `Multi-line Text`, `Select One`, `Yes/No`, `Select Multiple`,
 * `Date`, `File`, `Email`, or `Phone`. Set at creation time and immutable
 * thereafter.
 */
export type PropertyType = Property['type']

/**
 * The kind of object a property is attached to.
 *
 * - `event` - properties on events
 * - `order` - properties on orders (buyer-level)
 * - `ticket` - properties on individual products within an order (tickets, add-ons, etc.)
 * - `contact` - properties on customer contact records
 *
 * Set at creation time and immutable thereafter.
 */
export type PropertyObjectType = Property['objectType']

/**
 * Lifecycle status of a property. `archived` properties are soft-deleted: any
 * existing reference remains valid, but only `active` properties should be used
 * in new workflows, forms, etc.
 */
export type PropertyStatus = Property['status']

/**
 * A single selectable option on a `Select One` or `Select Multiple` property.
 * The `value` is the identity persisted on every instance that selected it;
 * `label` is the display text.
 */
export type PropertyOption = NonNullable<PropertyCreateBody['options']>[number]

/** Successful response shape from `GET /v1/properties`. */
export interface ListPropertiesResponse {
  readonly data: readonly Property[]
  readonly hasMore: boolean
}

/** Query parameters accepted by `GET /v1/properties`. */
export interface PropertyListParams {
  /** Page number, 0-indexed. Default 0. */
  readonly page?: number
  /** Items per page. Default 20, max 100. */
  readonly pageSize?: number
  /** Filter by the type of object this property belongs to. */
  readonly objectType?: PropertyObjectType
  /** Filter by property data type. */
  readonly propertyType?: PropertyType
  /** Filter by property status. */
  readonly status?: PropertyStatus
  /** Field to sort by. Defaults to `name`. */
  readonly sort?: 'name' | 'description' | 'type' | 'objectType' | 'status'
  /** Sort direction. Defaults to `asc`. */
  readonly order?: 'asc' | 'desc'
  /** Searches property names, case-insensitive. */
  readonly search?: string
}

/** Body accepted by `POST /v1/properties`. */
export type PropertyCreateBody =
  paths['/v1/properties/']['post']['requestBody']['content']['application/json']

/**
 * Body accepted by `PATCH /v1/properties/{id}`. Only fields you provide are
 * changed; `description` accepts `null` to clear it. `type` and `objectType`
 * cannot be modified on an existing property.
 */
export type PropertyUpdateBody =
  paths['/v1/properties/{id}']['patch']['requestBody']['content']['application/json']
