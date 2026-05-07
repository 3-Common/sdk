import { createAutoPaginator } from '@/pagination/auto-paginator'

import type {
  Event,
  EventListParams,
  EventRetrieveParams,
  EventUpdateBody,
  ListEventsResponse,
} from './types'
import type { HttpClient } from '@/core/http-client'
import type { RequestOptions } from '@/types/public'

interface DetailEnvelope<T> {
  readonly data: T
}

/**
 * Events service. Bound as `client.events` on the main client.
 *
 * @public
 */
export interface EventsService {
  /**
   * List the authenticated host's events.
   *
   * @example
   * ```ts
   * const { data, hasMore } = await client.events.list({ status: 'open', pageSize: 50 })
   * ```
   */
  list(params?: EventListParams, options?: RequestOptions): Promise<ListEventsResponse>

  /**
   * Retrieve a single event by ID.
   *
   * @example
   * ```ts
   * const event = await client.events.retrieve('evt_123')
   * ```
   */
  retrieve(id: string, params?: EventRetrieveParams, options?: RequestOptions): Promise<Event>

  /**
   * Update an event's basic fields. Only the fields you provide change.
   *
   * @example
   * ```ts
   * const updated = await client.events.update('evt_123', { name: 'Renamed' })
   * ```
   */
  update(id: string, body: EventUpdateBody, options?: RequestOptions): Promise<Event>

  /**
   * Iterate every event matching the filter, paging automatically.
   *
   * @example
   * ```ts
   * for await (const event of client.events.listAutoPaginate({ status: 'open' })) {
   *   console.log(event.name)
   * }
   * ```
   */
  listAutoPaginate(params?: EventListParams, options?: RequestOptions): AsyncIterableIterator<Event>
}

/**
 * Build an events service bound to an {@link HttpClient}.
 *
 * @internal
 */
export function eventsService(http: HttpClient): EventsService {
  return {
    async list(
      params: EventListParams = {},
      options?: RequestOptions,
    ): Promise<ListEventsResponse> {
      return http.request<ListEventsResponse>({
        method: 'GET',
        path: '/events',
        query: paramsToQuery(params),
        options,
      })
    },

    async retrieve(
      id: string,
      params: EventRetrieveParams = {},
      options?: RequestOptions,
    ): Promise<Event> {
      if (typeof id !== 'string' || id.length === 0) {
        throw new TypeError('events.retrieve: `id` must be a non-empty string')
      }
      const response = await http.request<DetailEnvelope<Event>>({
        method: 'GET',
        path: `/events/${encodeURIComponent(id)}`,
        query: paramsToQuery(params),
        options,
      })
      return response.data
    },

    async update(id: string, body: EventUpdateBody, options?: RequestOptions): Promise<Event> {
      if (typeof id !== 'string' || id.length === 0) {
        throw new TypeError('events.update: `id` must be a non-empty string')
      }
      const response = await http.request<DetailEnvelope<Event>>({
        method: 'PATCH',
        path: `/events/${encodeURIComponent(id)}`,
        body: body,
        options,
      })
      return response.data
    },

    listAutoPaginate(
      params: EventListParams = {},
      options?: RequestOptions,
    ): AsyncIterableIterator<Event> {
      const fetchPage = async (
        pageParams: EventListParams & { page: number },
      ): Promise<{ data: readonly Event[]; hasMore: boolean }> => {
        const result = await http.request<ListEventsResponse>({
          method: 'GET',
          path: '/events',
          query: paramsToQuery(pageParams),
          options,
        })
        return { data: result.data, hasMore: result.hasMore }
      }
      return createAutoPaginator<Event, EventListParams>(fetchPage, params)
    },
  }
}

function paramsToQuery(
  params: EventListParams | EventRetrieveParams,
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
