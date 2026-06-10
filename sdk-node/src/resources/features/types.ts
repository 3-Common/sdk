/**
 * Public types for the features resource. Hand-curated friendly aliases over
 * the auto-generated OpenAPI types.
 *
 * @public
 */

import type { paths } from '@/generated/types'

/**
 * One feature in the host's catalog. Detail responses populate every field the
 * server has for that record; list responses with a `fields` filter only
 * include the requested values.
 */
export type Feature = NonNullable<
  paths['/v1/features/{id}']['get']['responses'][200]['content']['application/json']['data']
>

/**
 * Feature value shape.
 *
 * - `boolean` — pure on/off.
 * - `quantity` — countable; drives entitlement balance.
 * - `enum` — one of a fixed ordered set of values.
 * - `duration` — number of days (or unlimited).
 */
export type FeatureType = NonNullable<Feature['type']>

/**
 * The resolved state of a feature for a customer, returned by
 * `GET /v1/features/resolve`. Combines the catalog feature, the resolved
 * type-specific value, and the subscriptions that contributed it.
 */
export type ResolvedFeature =
  paths['/v1/features/resolve']['get']['responses'][200]['content']['application/json']['data']

/**
 * The resolved type-specific value of a feature for a customer. A discriminated
 * union on `type`: `boolean`, `quantity`, `enum`, or `duration`.
 */
export type ResolvedFeatureValue = ResolvedFeature['value']

/** Successful response shape from `GET /v1/features`. */
export interface ListFeaturesResponse {
  readonly data: readonly Feature[]
  readonly hasMore: boolean
}

/** Query parameters accepted by `GET /v1/features`. */
export interface FeatureListParams {
  /** Page number, 0-indexed. Default 0. */
  readonly page?: number
  /** Items per page. */
  readonly pageSize?: number
  /** Filter by value shape; omit to include all. */
  readonly type?: FeatureType
  /** When set, returns only active (`true`) or only archived (`false`) features. */
  readonly active?: boolean
  /** Comma-separated list of fields to include in the response. Omit for all fields. */
  readonly fields?: string
}

/** Query parameters accepted by `GET /v1/features/{id}`. */
export interface FeatureRetrieveParams {
  /** Comma-separated list of fields to include in the response. */
  readonly fields?: string
}

/** Query parameters accepted by `GET /v1/features/resolve`. */
export interface FeatureResolveParams {
  /** CRM contact id. */
  readonly contactId: string
  /** Feature catalog key. */
  readonly featureKey: string
}

/** Body accepted by `POST /v1/features`. */
export type FeatureCreateBody =
  paths['/v1/features/']['post']['requestBody']['content']['application/json']

/**
 * Body accepted by `PATCH /v1/features/{id}`. Only fields you provide are
 * changed; `description` and `metadata` accept `null` to clear. `key` and
 * `type` are immutable — archive and create a new feature to change them.
 */
export type FeatureUpdateBody =
  paths['/v1/features/{id}']['patch']['requestBody']['content']['application/json']
