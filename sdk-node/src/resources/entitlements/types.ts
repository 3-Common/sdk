/**
 * Public types for the entitlements resource. Hand-curated friendly aliases
 * over the auto-generated OpenAPI types.
 *
 * @public
 */

import type { paths } from '@/generated/types'

/**
 * One entitlement balance record as returned by the API. Detail responses
 * (`retrieve`, `lookup`, `grant`, `consume`) populate every field the server
 * has for that record; list responses with a `fields` filter only include the
 * requested values.
 */
export type Entitlement = NonNullable<
  paths['/v1/entitlements/{id}']['get']['responses'][200]['content']['application/json']['data']
>

/** One grant in an entitlement's grant history. */
export type EntitlementGrant = NonNullable<Entitlement['grants']>[number]

/**
 * Source of an entitlement grant.
 *
 * - `subscription_recurring` — cycle grant from a subscription renewal.
 * - `one_time_addon` — top-up purchase (consumed first by `consume`).
 * - `manual` — admin-applied grant.
 */
export type EntitlementGrantSource = EntitlementGrant['source']

/** Successful response shape from `GET /v1/entitlements`. */
export interface ListEntitlementsResponse {
  readonly data: readonly Entitlement[]
  readonly hasMore: boolean
}

/** Query parameters accepted by `GET /v1/entitlements`. */
export interface EntitlementListParams {
  /** Page number, 0-indexed. Default 0. */
  readonly page?: number
  /** Items per page. */
  readonly pageSize?: number
  /** Filter by recipient contact id. */
  readonly contactId?: string
  /** Filter by feature key. */
  readonly featureKey?: string
  /** Only include entitlements whose `balance` is greater than or equal to this. */
  readonly minBalance?: number
  /** Comma-separated list of fields to include in the response. Omit for all fields. */
  readonly fields?: string
}

/** Query parameters accepted by `GET /v1/entitlements/{id}`. */
export interface EntitlementRetrieveParams {
  /** Comma-separated list of fields to include in the response. */
  readonly fields?: string
}

/** Query parameters accepted by `GET /v1/entitlements/lookup`. */
export interface EntitlementLookupParams {
  /** CRM contact id. */
  readonly contactId: string
  /** Feature key. */
  readonly featureKey: string
  /** Comma-separated list of fields to include in the response. Omit for all fields. */
  readonly fields?: string
}

/** Body accepted by `POST /v1/entitlements/grants`. */
export type EntitlementGrantBody =
  paths['/v1/entitlements/grants']['post']['requestBody']['content']['application/json']

/** Body accepted by `POST /v1/entitlements/consume`. */
export type EntitlementConsumeBody =
  paths['/v1/entitlements/consume']['post']['requestBody']['content']['application/json']
