import { createAutoPaginator } from '@/pagination/auto-paginator'

import type {
  Entitlement,
  EntitlementConsumeBody,
  EntitlementGrantBody,
  EntitlementListParams,
  EntitlementLookupParams,
  EntitlementRetrieveParams,
  ListEntitlementsResponse,
} from './types'
import type { HttpClient } from '@/core/http-client'
import type { RequestOptions } from '@/types/public'

interface DetailEnvelope<T> {
  readonly data: T
}

/**
 * Entitlements service. Bound as `client.entitlements` on the main client.
 *
 * Wraps `GET /v1/entitlements`, `GET /v1/entitlements/lookup`,
 * `GET /v1/entitlements/{id}`, `POST /v1/entitlements/grants`, and
 * `POST /v1/entitlements/consume`.
 *
 * @public
 */
export interface EntitlementsService {
  /**
   * List the authenticated host's entitlement balance records. Filterable by
   * contact, feature, or minimum balance; sorted by most-recently-updated.
   *
   * @example
   * ```ts
   * const { data, hasMore } = await client.entitlements.list({ featureKey: 'api_calls', minBalance: 1 })
   * ```
   */
  list(params?: EntitlementListParams, options?: RequestOptions): Promise<ListEntitlementsResponse>

  /**
   * Retrieve a single entitlement record by id, including grant history.
   *
   * @example
   * ```ts
   * const entitlement = await client.entitlements.retrieve('ent_123')
   * ```
   */
  retrieve(
    id: string,
    params?: EntitlementRetrieveParams,
    options?: RequestOptions,
  ): Promise<Entitlement>

  /**
   * Look up the unique entitlement record for a `(contactId, featureKey)`
   * pair. Throws `ThreeCommonNotFoundError` if no record exists yet.
   *
   * @example
   * ```ts
   * const entitlement = await client.entitlements.lookup({ contactId: 'cnt_7', featureKey: 'api_calls' })
   * ```
   */
  lookup(params: EntitlementLookupParams, options?: RequestOptions): Promise<Entitlement>

  /**
   * Add a manual entitlement grant for a `(contactId, featureKey)` — useful for
   * admin top-ups, comp credits, or migration. Idempotent on `grantId`.
   *
   * @example
   * ```ts
   * const entitlement = await client.entitlements.grant({
   *   contactId: 'cnt_7',
   *   featureKey: 'api_calls',
   *   amount: 100,
   *   grantId: 'grant_2026_q2_comp',
   * })
   * ```
   */
  grant(body: EntitlementGrantBody, options?: RequestOptions): Promise<Entitlement>

  /**
   * Debit units from a customer's entitlement balance. Throws on insufficient
   * balance. Debits `one_time_addon` grants first, then `manual`, then
   * `subscription_recurring` (FIFO within source).
   *
   * @example
   * ```ts
   * const entitlement = await client.entitlements.consume({
   *   contactId: 'cnt_7',
   *   featureKey: 'api_calls',
   *   amount: 1,
   *   reason: 'POST /generate',
   * })
   * ```
   */
  consume(body: EntitlementConsumeBody, options?: RequestOptions): Promise<Entitlement>

  /**
   * Iterate every entitlement matching the filter, paging automatically.
   *
   * @example
   * ```ts
   * for await (const entitlement of client.entitlements.listAutoPaginate({ featureKey: 'api_calls' })) {
   *   console.log(entitlement.contactId, entitlement.balance)
   * }
   * ```
   */
  listAutoPaginate(
    params?: EntitlementListParams,
    options?: RequestOptions,
  ): AsyncIterableIterator<Entitlement>
}

/**
 * Build an entitlements service bound to an {@link HttpClient}.
 *
 * @internal
 */
export function entitlementsService(http: HttpClient): EntitlementsService {
  return {
    async list(
      params: EntitlementListParams = {},
      options?: RequestOptions,
    ): Promise<ListEntitlementsResponse> {
      return http.request<ListEntitlementsResponse>({
        method: 'GET',
        path: '/entitlements',
        query: paramsToQuery(params),
        options,
      })
    },

    async retrieve(
      id: string,
      params: EntitlementRetrieveParams = {},
      options?: RequestOptions,
    ): Promise<Entitlement> {
      requireId('retrieve', id)
      const response = await http.request<DetailEnvelope<Entitlement>>({
        method: 'GET',
        path: `/entitlements/${encodeURIComponent(id)}`,
        query: paramsToQuery(params),
        options,
      })
      return response.data
    },

    async lookup(params: EntitlementLookupParams, options?: RequestOptions): Promise<Entitlement> {
      requireParam('lookup', 'contactId', params.contactId)
      requireParam('lookup', 'featureKey', params.featureKey)
      const response = await http.request<DetailEnvelope<Entitlement>>({
        method: 'GET',
        path: '/entitlements/lookup',
        query: paramsToQuery(params),
        options,
      })
      return response.data
    },

    async grant(body: EntitlementGrantBody, options?: RequestOptions): Promise<Entitlement> {
      const response = await http.request<DetailEnvelope<Entitlement>>({
        method: 'POST',
        path: '/entitlements/grants',
        body,
        options,
      })
      return response.data
    },

    async consume(body: EntitlementConsumeBody, options?: RequestOptions): Promise<Entitlement> {
      const response = await http.request<DetailEnvelope<Entitlement>>({
        method: 'POST',
        path: '/entitlements/consume',
        body,
        options,
      })
      return response.data
    },

    listAutoPaginate(
      params: EntitlementListParams = {},
      options?: RequestOptions,
    ): AsyncIterableIterator<Entitlement> {
      const fetchPage = async (
        pageParams: EntitlementListParams & { page: number },
      ): Promise<{ data: readonly Entitlement[]; hasMore: boolean }> => {
        const result = await http.request<ListEntitlementsResponse>({
          method: 'GET',
          path: '/entitlements',
          query: paramsToQuery(pageParams),
          options,
        })
        return { data: result.data, hasMore: result.hasMore }
      }
      return createAutoPaginator<Entitlement, EntitlementListParams>(fetchPage, params)
    },
  }
}

function requireId(method: string, id: string): void {
  if (typeof id !== 'string' || id.length === 0) {
    throw new TypeError(`entitlements.${method}: \`id\` must be a non-empty string`)
  }
}

function requireParam(method: string, name: string, value: string): void {
  if (typeof value !== 'string' || value.length === 0) {
    throw new TypeError(`entitlements.${method}: \`${name}\` must be a non-empty string`)
  }
}

function paramsToQuery(
  params: EntitlementListParams | EntitlementRetrieveParams | EntitlementLookupParams,
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
