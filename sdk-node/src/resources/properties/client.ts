import { createAutoPaginator } from '@/pagination/auto-paginator'

import type {
  ListPropertiesResponse,
  Property,
  PropertyCreateBody,
  PropertyListParams,
  PropertyUpdateBody,
} from './types'
import type { HttpClient } from '@/core/http-client'
import type { RequestOptions } from '@/types/public'

interface DetailEnvelope<T> {
  readonly data: T
}

/**
 * Properties service. Bound as `client.properties` on the main client.
 *
 * Wraps `GET /v1/properties`, `POST /v1/properties`, `GET /v1/properties/{id}`,
 * and `PATCH /v1/properties/{id}`.
 *
 * @public
 */
export interface PropertiesService {
  /**
   * List the authenticated host's properties.
   *
   * @example
   * ```ts
   * const { data, hasMore } = await client.properties.list({ objectType: 'contact', status: 'active' })
   * ```
   */
  list(params?: PropertyListParams, options?: RequestOptions): Promise<ListPropertiesResponse>

  /**
   * Retrieve a single property by id.
   *
   * @example
   * ```ts
   * const property = await client.properties.retrieve('prop_123')
   * ```
   */
  retrieve(id: string, options?: RequestOptions): Promise<Property>

  /**
   * Create a new property. `type` and `objectType` can only be set here and
   * cannot be modified afterwards. For `Select One` and `Select Multiple`
   * types, `options` is required and must have at least one entry.
   *
   * @example
   * ```ts
   * const property = await client.properties.create({
   *   type: 'Text',
   *   name: 'Dietary notes',
   *   status: 'active',
   *   objectType: 'contact',
   * })
   * ```
   */
  create(body: PropertyCreateBody, options?: RequestOptions): Promise<Property>

  /**
   * Apply a partial update to a property. Only fields you provide are changed;
   * `description` accepts `null` to clear it. To retire a property, set
   * `status` to `archived` (properties cannot be fully deleted). `type` and
   * `objectType` are immutable.
   *
   * @example
   * ```ts
   * const property = await client.properties.update('prop_123', { name: 'Allergies' })
   * ```
   */
  update(id: string, body: PropertyUpdateBody, options?: RequestOptions): Promise<Property>

  /**
   * Iterate every property matching the filter, paging automatically.
   *
   * @example
   * ```ts
   * for await (const property of client.properties.listAutoPaginate({ objectType: 'contact' })) {
   *   console.log(property.id, property.name)
   * }
   * ```
   */
  listAutoPaginate(
    params?: PropertyListParams,
    options?: RequestOptions,
  ): AsyncIterableIterator<Property>
}

/**
 * Build a properties service bound to an {@link HttpClient}.
 *
 * @internal
 */
export function propertiesService(http: HttpClient): PropertiesService {
  return {
    async list(
      params: PropertyListParams = {},
      options?: RequestOptions,
    ): Promise<ListPropertiesResponse> {
      return http.request<ListPropertiesResponse>({
        method: 'GET',
        path: '/properties',
        query: listParamsToQuery(params),
        options,
      })
    },

    async retrieve(id: string, options?: RequestOptions): Promise<Property> {
      requireId('retrieve', id)
      const response = await http.request<DetailEnvelope<Property>>({
        method: 'GET',
        path: `/properties/${encodeURIComponent(id)}`,
        options,
      })
      return response.data
    },

    async create(body: PropertyCreateBody, options?: RequestOptions): Promise<Property> {
      const response = await http.request<DetailEnvelope<Property>>({
        method: 'POST',
        path: '/properties',
        body,
        options,
      })
      return response.data
    },

    async update(
      id: string,
      body: PropertyUpdateBody,
      options?: RequestOptions,
    ): Promise<Property> {
      requireId('update', id)
      const response = await http.request<DetailEnvelope<Property>>({
        method: 'PATCH',
        path: `/properties/${encodeURIComponent(id)}`,
        body,
        options,
      })
      return response.data
    },

    listAutoPaginate(
      params: PropertyListParams = {},
      options?: RequestOptions,
    ): AsyncIterableIterator<Property> {
      const fetchPage = async (
        pageParams: PropertyListParams & { page: number },
      ): Promise<{ data: readonly Property[]; hasMore: boolean }> => {
        const result = await http.request<ListPropertiesResponse>({
          method: 'GET',
          path: '/properties',
          query: listParamsToQuery(pageParams),
          options,
        })
        return { data: result.data, hasMore: result.hasMore }
      }
      return createAutoPaginator<Property, PropertyListParams>(fetchPage, params)
    },
  }
}

function requireId(method: string, id: string): void {
  if (typeof id !== 'string' || id.length === 0) {
    throw new TypeError(`properties.${method}: \`id\` must be a non-empty string`)
  }
}

function listParamsToQuery(
  params: PropertyListParams,
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
