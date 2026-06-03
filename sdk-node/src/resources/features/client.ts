import { createAutoPaginator } from '@/pagination/auto-paginator'

import type {
  Feature,
  FeatureCreateBody,
  FeatureListParams,
  FeatureResolveParams,
  FeatureRetrieveParams,
  FeatureUpdateBody,
  ListFeaturesResponse,
  ResolvedFeature,
} from './types'
import type { HttpClient } from '@/core/http-client'
import type { RequestOptions } from '@/types/public'

interface DetailEnvelope<T> {
  readonly data: T
}

/**
 * Features service. Bound as `client.features` on the main client.
 *
 * Wraps `GET /v1/features`, `GET /v1/features/resolve`,
 * `GET /v1/features/{id}`, `POST /v1/features`, `PATCH /v1/features/{id}`,
 * `POST /v1/features/{id}/archive`, and `POST /v1/features/{id}/unarchive`.
 *
 * @public
 */
export interface FeaturesService {
  /**
   * List the authenticated host's feature catalog.
   *
   * @example
   * ```ts
   * const { data, hasMore } = await client.features.list({ type: 'quantity', active: true })
   * ```
   */
  list(params?: FeatureListParams, options?: RequestOptions): Promise<ListFeaturesResponse>

  /**
   * Resolve the current value of a feature for a customer by walking active
   * subscriptions → prices → feature grants. Returns the type-specific value
   * plus the contributing subscription ids (and, for quantity features, the
   * live entitlement balance). Throws `ThreeCommonNotFoundError` if the feature
   * key is unknown.
   *
   * @example
   * ```ts
   * const resolved = await client.features.resolve({ contactId: 'cnt_7', featureKey: 'api_calls' })
   * ```
   */
  resolve(params: FeatureResolveParams, options?: RequestOptions): Promise<ResolvedFeature>

  /**
   * Retrieve a single feature by id.
   *
   * @example
   * ```ts
   * const feature = await client.features.retrieve('feat_123')
   * ```
   */
  retrieve(id: string, params?: FeatureRetrieveParams, options?: RequestOptions): Promise<Feature>

  /**
   * Create a feature in the catalog. The `key` is the stable
   * application-facing identifier; `type` decides how prices grant the feature
   * and how it resolves.
   *
   * @example
   * ```ts
   * const feature = await client.features.create({
   *   key: 'api_calls',
   *   name: 'API calls',
   *   type: 'quantity',
   * })
   * ```
   */
  create(body: FeatureCreateBody, options?: RequestOptions): Promise<Feature>

  /**
   * Apply a partial update to a feature. Mutable fields: `name`,
   * `description`, `enumValues`, `metadata`. `key` and `type` are immutable —
   * archive and create a new feature to change them.
   *
   * @example
   * ```ts
   * const feature = await client.features.update('feat_123', { name: 'API requests' })
   * ```
   */
  update(id: string, body: FeatureUpdateBody, options?: RequestOptions): Promise<Feature>

  /**
   * Soft-archive a feature. Idempotent.
   *
   * @example
   * ```ts
   * await client.features.archive('feat_123')
   * ```
   */
  archive(id: string, options?: RequestOptions): Promise<Feature>

  /**
   * Reactivate a previously archived feature. Idempotent.
   *
   * @example
   * ```ts
   * await client.features.unarchive('feat_123')
   * ```
   */
  unarchive(id: string, options?: RequestOptions): Promise<Feature>

  /**
   * Iterate every feature matching the filter, paging automatically.
   *
   * @example
   * ```ts
   * for await (const feature of client.features.listAutoPaginate({ active: true })) {
   *   console.log(feature.key, feature.type)
   * }
   * ```
   */
  listAutoPaginate(
    params?: FeatureListParams,
    options?: RequestOptions,
  ): AsyncIterableIterator<Feature>
}

/**
 * Build a features service bound to an {@link HttpClient}.
 *
 * @internal
 */
export function featuresService(http: HttpClient): FeaturesService {
  return {
    async list(
      params: FeatureListParams = {},
      options?: RequestOptions,
    ): Promise<ListFeaturesResponse> {
      return http.request<ListFeaturesResponse>({
        method: 'GET',
        path: '/features',
        query: listParamsToQuery(params),
        options,
      })
    },

    async resolve(
      params: FeatureResolveParams,
      options?: RequestOptions,
    ): Promise<ResolvedFeature> {
      const response = await http.request<DetailEnvelope<ResolvedFeature>>({
        method: 'GET',
        path: '/features/resolve',
        query: resolveParamsToQuery(params),
        options,
      })
      return response.data
    },

    async retrieve(
      id: string,
      params: FeatureRetrieveParams = {},
      options?: RequestOptions,
    ): Promise<Feature> {
      requireId('retrieve', id)
      const response = await http.request<DetailEnvelope<Feature>>({
        method: 'GET',
        path: `/features/${encodeURIComponent(id)}`,
        query: retrieveParamsToQuery(params),
        options,
      })
      return response.data
    },

    async create(body: FeatureCreateBody, options?: RequestOptions): Promise<Feature> {
      const response = await http.request<DetailEnvelope<Feature>>({
        method: 'POST',
        path: '/features',
        body,
        options,
      })
      return response.data
    },

    async update(id: string, body: FeatureUpdateBody, options?: RequestOptions): Promise<Feature> {
      requireId('update', id)
      const response = await http.request<DetailEnvelope<Feature>>({
        method: 'PATCH',
        path: `/features/${encodeURIComponent(id)}`,
        body,
        options,
      })
      return response.data
    },

    async archive(id: string, options?: RequestOptions): Promise<Feature> {
      requireId('archive', id)
      const response = await http.request<DetailEnvelope<Feature>>({
        method: 'POST',
        path: `/features/${encodeURIComponent(id)}/archive`,
        options,
      })
      return response.data
    },

    async unarchive(id: string, options?: RequestOptions): Promise<Feature> {
      requireId('unarchive', id)
      const response = await http.request<DetailEnvelope<Feature>>({
        method: 'POST',
        path: `/features/${encodeURIComponent(id)}/unarchive`,
        options,
      })
      return response.data
    },

    listAutoPaginate(
      params: FeatureListParams = {},
      options?: RequestOptions,
    ): AsyncIterableIterator<Feature> {
      const fetchPage = async (
        pageParams: FeatureListParams & { page: number },
      ): Promise<{ data: readonly Feature[]; hasMore: boolean }> => {
        const result = await http.request<ListFeaturesResponse>({
          method: 'GET',
          path: '/features',
          query: listParamsToQuery(pageParams),
          options,
        })
        return { data: result.data, hasMore: result.hasMore }
      }
      return createAutoPaginator<Feature, FeatureListParams>(fetchPage, params)
    },
  }
}

function requireId(method: string, id: string): void {
  if (typeof id !== 'string' || id.length === 0) {
    throw new TypeError(`features.${method}: \`id\` must be a non-empty string`)
  }
}

function listParamsToQuery(
  params: FeatureListParams,
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

function resolveParamsToQuery(params: FeatureResolveParams): Record<string, string | undefined> {
  return { contactId: params.contactId, featureKey: params.featureKey }
}

function retrieveParamsToQuery(params: FeatureRetrieveParams): Record<string, string | undefined> {
  if (params.fields === undefined) return {}
  return { fields: params.fields }
}
