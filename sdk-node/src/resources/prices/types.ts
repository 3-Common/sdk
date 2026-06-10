/**
 * Public types for the prices resource. Hand-curated friendly aliases over the
 * auto-generated OpenAPI types.
 *
 * @public
 */

import type { paths } from '@/generated/types'

/**
 * One price as returned by the API. Detail responses populate every field the
 * server has for that record; list responses with a `fields` filter only
 * include the requested values.
 */
export type Price = NonNullable<
  paths['/v1/prices/{id}']['get']['responses'][200]['content']['application/json']['data']
>

/**
 * Price cadence.
 *
 * - `recurring` — billed on a fixed cadence (subscription-backed).
 * - `one_time` — single charge, typically an add-on / top-up pack.
 */
export type PriceType = NonNullable<Price['type']>

/** Settlement currency of a price. */
export type PriceCurrency = NonNullable<Price['currency']>

/** Recurring cadence descriptor, present when `type` is `recurring`. */
export type PriceRecurring = NonNullable<Price['recurring']>

/** Cadence unit of a recurring price. */
export type PriceInterval = PriceRecurring['interval']

/**
 * One typed feature grant on a price. A discriminated union on `type`:
 * `boolean`, `quantity`, `enum`, or `duration`.
 */
export type PriceFeature = NonNullable<Price['features']>[number]

/** Successful response shape from `GET /v1/prices`. */
export interface ListPricesResponse {
  readonly data: readonly Price[]
  readonly hasMore: boolean
}

/** Query parameters accepted by `GET /v1/prices`. */
export interface PriceListParams {
  /** Page number, 0-indexed. Default 0. */
  readonly page?: number
  /** Items per page. */
  readonly pageSize?: number
  /** Filter by parent product. */
  readonly productId?: string
  /** Filter by cadence; omit to include all. */
  readonly type?: PriceType
  /** When set, returns only active (`true`) or only archived (`false`) prices. */
  readonly active?: boolean
  /** Comma-separated list of fields to include in the response. Omit for all fields. */
  readonly fields?: string
}

/** Query parameters accepted by `GET /v1/prices/{id}`. */
export interface PriceRetrieveParams {
  /** Comma-separated list of fields to include in the response. */
  readonly fields?: string
}

/** Body accepted by `POST /v1/prices`. */
export type PriceCreateBody =
  paths['/v1/prices/']['post']['requestBody']['content']['application/json']

/**
 * Body accepted by `PATCH /v1/prices/{id}`. Only fields you provide are
 * changed; `features`, `nickname`, and `metadata` accept `null` to clear.
 */
export type PriceUpdateBody =
  paths['/v1/prices/{id}']['patch']['requestBody']['content']['application/json']
