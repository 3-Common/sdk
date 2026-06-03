import { createAutoPaginator } from '@/pagination/auto-paginator'

import type {
  ListPricesResponse,
  Price,
  PriceCreateBody,
  PriceListParams,
  PriceRetrieveParams,
  PriceUpdateBody,
} from './types'
import type { HttpClient } from '@/core/http-client'
import type { RequestOptions } from '@/types/public'

interface DetailEnvelope<T> {
  readonly data: T
}

/**
 * Prices service. Bound as `client.prices` on the main client.
 *
 * Wraps `GET /v1/prices`, `POST /v1/prices`, `GET /v1/prices/{id}`,
 * `PATCH /v1/prices/{id}`, `POST /v1/prices/{id}/archive`, and
 * `POST /v1/prices/{id}/unarchive`.
 *
 * @public
 */
export interface PricesService {
  /**
   * List the authenticated host's prices.
   *
   * @example
   * ```ts
   * const { data, hasMore } = await client.prices.list({ productId: 'prod_7', active: true })
   * ```
   */
  list(params?: PriceListParams, options?: RequestOptions): Promise<ListPricesResponse>

  /**
   * Retrieve a single price by id.
   *
   * @example
   * ```ts
   * const price = await client.prices.retrieve('price_123')
   * ```
   */
  retrieve(id: string, params?: PriceRetrieveParams, options?: RequestOptions): Promise<Price>

  /**
   * Create a price for a product. Defines cadence (`one_time` or `recurring`),
   * per-unit cost, and an optional array of typed feature grants.
   *
   * @example
   * ```ts
   * const price = await client.prices.create({
   *   productId: 'prod_7',
   *   type: 'recurring',
   *   currency: 'USD',
   *   unitAmount: 1500,
   *   recurring: { interval: 'month', intervalCount: 1 },
   * })
   * ```
   */
  create(body: PriceCreateBody, options?: RequestOptions): Promise<Price>

  /**
   * Apply a partial update to a price. Mutable fields: `unitAmount`,
   * `recurring`, `features`, `nickname`, `metadata`. To switch
   * type/currency/product, archive and create a new price.
   *
   * @example
   * ```ts
   * const price = await client.prices.update('price_123', { unitAmount: 1200 })
   * ```
   */
  update(id: string, body: PriceUpdateBody, options?: RequestOptions): Promise<Price>

  /**
   * Soft-archive a price. Idempotent. Existing subscriptions are unaffected;
   * new subscriptions cannot select this price until unarchived.
   *
   * @example
   * ```ts
   * await client.prices.archive('price_123')
   * ```
   */
  archive(id: string, options?: RequestOptions): Promise<Price>

  /**
   * Reactivate a previously archived price. Idempotent.
   *
   * @example
   * ```ts
   * await client.prices.unarchive('price_123')
   * ```
   */
  unarchive(id: string, options?: RequestOptions): Promise<Price>

  /**
   * Iterate every price matching the filter, paging automatically.
   *
   * @example
   * ```ts
   * for await (const price of client.prices.listAutoPaginate({ active: true })) {
   *   console.log(price.id, price.unitAmount)
   * }
   * ```
   */
  listAutoPaginate(params?: PriceListParams, options?: RequestOptions): AsyncIterableIterator<Price>
}

/**
 * Build a prices service bound to an {@link HttpClient}.
 *
 * @internal
 */
export function pricesService(http: HttpClient): PricesService {
  return {
    async list(
      params: PriceListParams = {},
      options?: RequestOptions,
    ): Promise<ListPricesResponse> {
      return http.request<ListPricesResponse>({
        method: 'GET',
        path: '/prices',
        query: listParamsToQuery(params),
        options,
      })
    },

    async retrieve(
      id: string,
      params: PriceRetrieveParams = {},
      options?: RequestOptions,
    ): Promise<Price> {
      requireId('retrieve', id)
      const response = await http.request<DetailEnvelope<Price>>({
        method: 'GET',
        path: `/prices/${encodeURIComponent(id)}`,
        query: retrieveParamsToQuery(params),
        options,
      })
      return response.data
    },

    async create(body: PriceCreateBody, options?: RequestOptions): Promise<Price> {
      const response = await http.request<DetailEnvelope<Price>>({
        method: 'POST',
        path: '/prices',
        body,
        options,
      })
      return response.data
    },

    async update(id: string, body: PriceUpdateBody, options?: RequestOptions): Promise<Price> {
      requireId('update', id)
      const response = await http.request<DetailEnvelope<Price>>({
        method: 'PATCH',
        path: `/prices/${encodeURIComponent(id)}`,
        body,
        options,
      })
      return response.data
    },

    async archive(id: string, options?: RequestOptions): Promise<Price> {
      requireId('archive', id)
      const response = await http.request<DetailEnvelope<Price>>({
        method: 'POST',
        path: `/prices/${encodeURIComponent(id)}/archive`,
        options,
      })
      return response.data
    },

    async unarchive(id: string, options?: RequestOptions): Promise<Price> {
      requireId('unarchive', id)
      const response = await http.request<DetailEnvelope<Price>>({
        method: 'POST',
        path: `/prices/${encodeURIComponent(id)}/unarchive`,
        options,
      })
      return response.data
    },

    listAutoPaginate(
      params: PriceListParams = {},
      options?: RequestOptions,
    ): AsyncIterableIterator<Price> {
      const fetchPage = async (
        pageParams: PriceListParams & { page: number },
      ): Promise<{ data: readonly Price[]; hasMore: boolean }> => {
        const result = await http.request<ListPricesResponse>({
          method: 'GET',
          path: '/prices',
          query: listParamsToQuery(pageParams),
          options,
        })
        return { data: result.data, hasMore: result.hasMore }
      }
      return createAutoPaginator<Price, PriceListParams>(fetchPage, params)
    },
  }
}

function requireId(method: string, id: string): void {
  if (typeof id !== 'string' || id.length === 0) {
    throw new TypeError(`prices.${method}: \`id\` must be a non-empty string`)
  }
}

function listParamsToQuery(
  params: PriceListParams,
): Record<string, string | number | boolean | undefined> {
  const query: Record<string, string | number | boolean | undefined> = {}
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined) continue
    if (typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') {
      query[key] = value
    }
  }
  return query
}

function retrieveParamsToQuery(params: PriceRetrieveParams): Record<string, string | undefined> {
  if (params.fields === undefined) return {}
  return { fields: params.fields }
}
