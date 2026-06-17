import { createAutoPaginator } from '@/pagination/auto-paginator'

import type {
  AttachPaymentMethodBody,
  AttachPaymentMethodResult,
  BulkUpsertContactsResult,
  Contact,
  ContactActivity,
  ContactActivityListParams,
  ContactBulkUpsertBody,
  ContactCountResult,
  ContactCreateBody,
  ContactListParams,
  ContactUpdateBody,
  ContactWithOrderDetails,
  DeletedContact,
  ListContactActivityResponse,
  ListContactsResponse,
  PaymentMethod,
  PaymentMethodSetupIntent,
  RemovedPaymentMethod,
} from './types'
import type { HttpClient } from '@/core/http-client'
import type { RequestOptions } from '@/types/public'

interface DetailEnvelope<T> {
  readonly data: T
}

/**
 * Contacts service. Bound as `client.contacts` on the main client.
 *
 * Wraps `GET /v1/contacts`, `GET /v1/contacts/count`, `POST /v1/contacts`,
 * `POST /v1/contacts/bulk`, `GET /v1/contacts/{id}`,
 * `PATCH /v1/contacts/{id}`, `DELETE /v1/contacts/{id}`,
 * `GET /v1/contacts/{id}/activity`,
 * `GET /v1/contacts/{id}/payment-methods`,
 * `POST /v1/contacts/{id}/payment-methods`,
 * `POST /v1/contacts/{id}/payment-methods/setup-intent`, and
 * `DELETE /v1/contacts/{id}/payment-methods/{methodId}`.
 *
 * @public
 */
export interface ContactsService {
  /**
   * List the authenticated host's contacts.
   *
   * @example
   * ```ts
   * const { data, hasMore } = await client.contacts.list({ filter: 'opted-in', pageSize: 50 })
   * ```
   */
  list(params?: ContactListParams, options?: RequestOptions): Promise<ListContactsResponse>

  /**
   * Return the total contact count for the host.
   *
   * @example
   * ```ts
   * const { count } = await client.contacts.count()
   * ```
   */
  count(options?: RequestOptions): Promise<ContactCountResult>

  /**
   * Retrieve a single contact by id.
   *
   * @example
   * ```ts
   * const contact = await client.contacts.retrieve('cnt_123')
   * ```
   */
  retrieve(id: string, options?: RequestOptions): Promise<Contact>

  /**
   * Create a new contact. Fails with `409 conflict` if a contact with the same
   * email already exists for this host.
   *
   * @example
   * ```ts
   * const contact = await client.contacts.create({
   *   email: 'guest@example.com',
   *   firstName: 'Alex',
   *   lastName: 'Garcia',
   * })
   * ```
   */
  create(body: ContactCreateBody, options?: RequestOptions): Promise<Contact>

  /**
   * Update a contact. Optionally merges a second contact (identified by
   * `mergeWith`) into the target — required `resolution` chooses how
   * field-level conflicts are resolved.
   *
   * Returns the richer order-details projection
   * ({@link ContactWithOrderDetails}), not the compact `Contact` shape.
   *
   * @example
   * ```ts
   * const updated = await client.contacts.update('cnt_123', {
   *   contact: { firstName: 'Alex', lastName: 'Garcia', email: 'a.garcia@example.com', status: 'opted-in' },
   * })
   * ```
   */
  update(
    id: string,
    body: ContactUpdateBody,
    options?: RequestOptions,
  ): Promise<ContactWithOrderDetails>

  /**
   * Permanently remove a contact. Returns `{ id }` echoing the removed
   * contact's id.
   *
   * @example
   * ```ts
   * await client.contacts.delete('cnt_123')
   * ```
   */
  delete(id: string, options?: RequestOptions): Promise<DeletedContact>

  /**
   * Bulk-upsert up to 10,000 contacts in one round-trip. Deduplicated
   * server-side by email; existing rows are updated rather than rejected.
   *
   * @example
   * ```ts
   * const { affected } = await client.contacts.bulkUpsert({
   *   contacts: [
   *     { email: 'a@example.com', firstName: 'Ada' },
   *     { email: 'b@example.com', firstName: 'Beatrix' },
   *   ],
   * })
   * ```
   */
  bulkUpsert(
    body: ContactBulkUpsertBody,
    options?: RequestOptions,
  ): Promise<BulkUpsertContactsResult>

  /**
   * Paginated activity log for a contact (checkouts, refunds, scans, emails,
   * invoice payments).
   *
   * @example
   * ```ts
   * const { data, hasMore } = await client.contacts.listActivity('cnt_123', { filter: 'checkout_session_completed' })
   * ```
   */
  listActivity(
    id: string,
    params?: ContactActivityListParams,
    options?: RequestOptions,
  ): Promise<ListContactActivityResponse>

  /**
   * Iterate every contact matching the filter, paging automatically.
   *
   * @example
   * ```ts
   * for await (const contact of client.contacts.listAutoPaginate({ filter: 'opted-in' })) {
   *   console.log(contact.email)
   * }
   * ```
   */
  listAutoPaginate(
    params?: ContactListParams,
    options?: RequestOptions,
  ): AsyncIterableIterator<Contact>

  /**
   * Iterate every activity record for a contact, paging automatically.
   *
   * @example
   * ```ts
   * for await (const event of client.contacts.listActivityAutoPaginate('cnt_123')) {
   *   console.log(event.type)
   * }
   * ```
   */
  listActivityAutoPaginate(
    id: string,
    params?: ContactActivityListParams,
    options?: RequestOptions,
  ): AsyncIterableIterator<ContactActivity>

  /**
   * Retrieve the saved card on file for a contact, or `null` when none is
   * saved. One card is supported per contact.
   *
   * @example
   * ```ts
   * const card = await client.contacts.retrievePaymentMethod('cnt_123')
   * if (card) console.log(`${card.card.brand} ••••${card.card.last4}`)
   * ```
   */
  retrievePaymentMethod(id: string, options?: RequestOptions): Promise<PaymentMethod | null>

  /**
   * Persist the card from a confirmed Stripe SetupIntent against the contact.
   * The SetupIntent is re-verified server-side before the card is saved, and
   * any existing card on file is replaced (`replacedExisting` reports whether
   * that happened).
   *
   * @example
   * ```ts
   * const { data, replacedExisting } = await client.contacts.attachPaymentMethod('cnt_123', {
   *   setupIntentId: 'seti_123',
   * })
   * ```
   */
  attachPaymentMethod(
    id: string,
    body: AttachPaymentMethodBody,
    options?: RequestOptions,
  ): Promise<AttachPaymentMethodResult>

  /**
   * Begin saving a card for a contact. Returns a Stripe SetupIntent
   * `clientSecret` to confirm client-side with Stripe Elements; once confirmed,
   * call {@link ContactsService.attachPaymentMethod} with the returned
   * `setupIntentId` to persist the card.
   *
   * @example
   * ```ts
   * const { clientSecret } = await client.contacts.createPaymentMethodSetupIntent('cnt_123')
   * ```
   */
  createPaymentMethodSetupIntent(
    id: string,
    options?: RequestOptions,
  ): Promise<PaymentMethodSetupIntent>

  /**
   * Detach the saved card from Stripe and remove it from the contact. Returns
   * `{ removed: true }` on success.
   *
   * @example
   * ```ts
   * await client.contacts.removePaymentMethod('cnt_123', 'pm_456')
   * ```
   */
  removePaymentMethod(
    id: string,
    methodId: string,
    options?: RequestOptions,
  ): Promise<RemovedPaymentMethod>
}

/**
 * Build a contacts service bound to an {@link HttpClient}.
 *
 * @internal
 */
export function contactsService(http: HttpClient): ContactsService {
  return {
    async list(
      params: ContactListParams = {},
      options?: RequestOptions,
    ): Promise<ListContactsResponse> {
      return http.request<ListContactsResponse>({
        method: 'GET',
        path: '/contacts',
        query: listParamsToQuery(params),
        options,
      })
    },

    async count(options?: RequestOptions): Promise<ContactCountResult> {
      const response = await http.request<DetailEnvelope<ContactCountResult>>({
        method: 'GET',
        path: '/contacts/count',
        options,
      })
      return response.data
    },

    async retrieve(id: string, options?: RequestOptions): Promise<Contact> {
      requireId('retrieve', id)
      const response = await http.request<DetailEnvelope<Contact>>({
        method: 'GET',
        path: `/contacts/${encodeURIComponent(id)}`,
        options,
      })
      return response.data
    },

    async create(body: ContactCreateBody, options?: RequestOptions): Promise<Contact> {
      const response = await http.request<DetailEnvelope<Contact>>({
        method: 'POST',
        path: '/contacts',
        body,
        options,
      })
      return response.data
    },

    async update(
      id: string,
      body: ContactUpdateBody,
      options?: RequestOptions,
    ): Promise<ContactWithOrderDetails> {
      requireId('update', id)
      const response = await http.request<DetailEnvelope<ContactWithOrderDetails>>({
        method: 'PATCH',
        path: `/contacts/${encodeURIComponent(id)}`,
        body,
        options,
      })
      return response.data
    },

    async delete(id: string, options?: RequestOptions): Promise<DeletedContact> {
      requireId('delete', id)
      const response = await http.request<DetailEnvelope<DeletedContact>>({
        method: 'DELETE',
        path: `/contacts/${encodeURIComponent(id)}`,
        options,
      })
      return response.data
    },

    async bulkUpsert(
      body: ContactBulkUpsertBody,
      options?: RequestOptions,
    ): Promise<BulkUpsertContactsResult> {
      const response = await http.request<DetailEnvelope<BulkUpsertContactsResult>>({
        method: 'POST',
        path: '/contacts/bulk',
        body,
        options,
      })
      return response.data
    },

    async listActivity(
      id: string,
      params: ContactActivityListParams = {},
      options?: RequestOptions,
    ): Promise<ListContactActivityResponse> {
      requireId('listActivity', id)
      return http.request<ListContactActivityResponse>({
        method: 'GET',
        path: `/contacts/${encodeURIComponent(id)}/activity`,
        query: activityParamsToQuery(params),
        options,
      })
    },

    listAutoPaginate(
      params: ContactListParams = {},
      options?: RequestOptions,
    ): AsyncIterableIterator<Contact> {
      // Contacts paginates on `pageNumber`, not the generic auto-paginator's
      // `page` cursor — translate at both ends.
      const { pageNumber, ...rest } = params
      const cursorInit: typeof rest & { page?: number } =
        pageNumber === undefined ? { ...rest } : { ...rest, page: pageNumber }
      const fetchPage = async (
        pageParams: typeof rest & { page: number },
      ): Promise<{ data: readonly Contact[]; hasMore: boolean }> => {
        const { page, ...remaining } = pageParams
        const result = await http.request<ListContactsResponse>({
          method: 'GET',
          path: '/contacts',
          query: listParamsToQuery({ ...remaining, pageNumber: page }),
          options,
        })
        return { data: result.data, hasMore: result.hasMore }
      }
      return createAutoPaginator(fetchPage, cursorInit)
    },

    listActivityAutoPaginate(
      id: string,
      params: ContactActivityListParams = {},
      options?: RequestOptions,
    ): AsyncIterableIterator<ContactActivity> {
      requireId('listActivityAutoPaginate', id)
      const { pageNumber, ...rest } = params
      const cursorInit: typeof rest & { page?: number } =
        pageNumber === undefined ? { ...rest } : { ...rest, page: pageNumber }
      const fetchPage = async (
        pageParams: typeof rest & { page: number },
      ): Promise<{ data: readonly ContactActivity[]; hasMore: boolean }> => {
        const { page, ...remaining } = pageParams
        const result = await http.request<ListContactActivityResponse>({
          method: 'GET',
          path: `/contacts/${encodeURIComponent(id)}/activity`,
          query: activityParamsToQuery({ ...remaining, pageNumber: page }),
          options,
        })
        return { data: result.data, hasMore: result.hasMore }
      }
      return createAutoPaginator(fetchPage, cursorInit)
    },

    async retrievePaymentMethod(
      id: string,
      options?: RequestOptions,
    ): Promise<PaymentMethod | null> {
      requireId('retrievePaymentMethod', id)
      const response = await http.request<DetailEnvelope<PaymentMethod | null>>({
        method: 'GET',
        path: `/contacts/${encodeURIComponent(id)}/payment-methods`,
        options,
      })
      return response.data
    },

    async attachPaymentMethod(
      id: string,
      body: AttachPaymentMethodBody,
      options?: RequestOptions,
    ): Promise<AttachPaymentMethodResult> {
      requireId('attachPaymentMethod', id)
      return http.request<AttachPaymentMethodResult>({
        method: 'POST',
        path: `/contacts/${encodeURIComponent(id)}/payment-methods`,
        body,
        options,
      })
    },

    async createPaymentMethodSetupIntent(
      id: string,
      options?: RequestOptions,
    ): Promise<PaymentMethodSetupIntent> {
      requireId('createPaymentMethodSetupIntent', id)
      const response = await http.request<DetailEnvelope<PaymentMethodSetupIntent>>({
        method: 'POST',
        path: `/contacts/${encodeURIComponent(id)}/payment-methods/setup-intent`,
        options,
      })
      return response.data
    },

    async removePaymentMethod(
      id: string,
      methodId: string,
      options?: RequestOptions,
    ): Promise<RemovedPaymentMethod> {
      requireId('removePaymentMethod', id)
      requireString('removePaymentMethod', 'methodId', methodId)
      const response = await http.request<DetailEnvelope<RemovedPaymentMethod>>({
        method: 'DELETE',
        path: `/contacts/${encodeURIComponent(id)}/payment-methods/${encodeURIComponent(methodId)}`,
        options,
      })
      return response.data
    },
  }
}

function requireId(method: string, id: string): void {
  if (typeof id !== 'string' || id.length === 0) {
    throw new TypeError(`contacts.${method}: \`id\` must be a non-empty string`)
  }
}

function requireString(method: string, name: string, value: string): void {
  if (typeof value !== 'string' || value.length === 0) {
    throw new TypeError(`contacts.${method}: \`${name}\` must be a non-empty string`)
  }
}

function listParamsToQuery(
  params: ContactListParams,
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

function activityParamsToQuery(
  params: ContactActivityListParams,
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
