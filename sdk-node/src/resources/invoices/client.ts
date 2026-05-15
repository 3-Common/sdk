import { createAutoPaginator } from '@/pagination/auto-paginator'

import type {
  Invoice,
  InvoiceCreateBody,
  InvoiceListParams,
  InvoicePaymentBody,
  InvoiceRetrieveParams,
  InvoiceUpdateBody,
  InvoiceVoidBody,
  ListInvoicesResponse,
} from './types'
import type { HttpClient } from '@/core/http-client'
import type { RequestOptions } from '@/types/public'

interface DetailEnvelope<T> {
  readonly data: T
}

/**
 * Invoices service. Bound as `client.invoices` on the main client.
 *
 * @public
 */
export interface InvoicesService {
  /**
   * List the authenticated host's invoices.
   *
   * @example
   * ```ts
   * const { data, hasMore } = await client.invoices.list({ status: 'open', pageSize: 50 })
   * ```
   */
  list(params?: InvoiceListParams, options?: RequestOptions): Promise<ListInvoicesResponse>

  /**
   * Retrieve a single invoice by ID.
   *
   * @example
   * ```ts
   * const invoice = await client.invoices.retrieve('inv_123')
   * ```
   */
  retrieve(id: string, params?: InvoiceRetrieveParams, options?: RequestOptions): Promise<Invoice>

  /**
   * Create a draft invoice. Totals are computed server-side from line items.
   *
   * @example
   * ```ts
   * const draft = await client.invoices.create({
   *   customerId: 'cnt_42',
   *   currency: 'USD',
   *   lineItems: [{ description: 'Consulting', quantity: 1, unitAmount: 50_000 }],
   * })
   * ```
   */
  create(body: InvoiceCreateBody, options?: RequestOptions): Promise<Invoice>

  /**
   * Revise a draft invoice. Only legal while in draft.
   *
   * @example
   * ```ts
   * const updated = await client.invoices.update('inv_123', { notes: 'Net 30' })
   * ```
   */
  update(id: string, body: InvoiceUpdateBody, options?: RequestOptions): Promise<Invoice>

  /**
   * Finalize a draft invoice: assigns a sequential number, stamps `issuedAt`,
   * and transitions the status to `open`.
   *
   * @example
   * ```ts
   * const issued = await client.invoices.finalize('inv_123')
   * ```
   */
  finalize(id: string, options?: RequestOptions): Promise<Invoice>

  /**
   * Void an invoice. Permitted from `draft` or `open`; paid invoices cannot be voided.
   *
   * @example
   * ```ts
   * await client.invoices.void('inv_123', { reason: 'Sent in error' })
   * ```
   */
  void(id: string, body?: InvoiceVoidBody, options?: RequestOptions): Promise<Invoice>

  /**
   * Record a manual payment against an open invoice. Cumulative payments
   * meeting the total transition the invoice to `paid`.
   *
   * @example
   * ```ts
   * await client.invoices.recordPayment('inv_123', { payment: 50_000, idempotencyKey: 'pmt-2026-05-11' })
   * ```
   */
  recordPayment(id: string, body: InvoicePaymentBody, options?: RequestOptions): Promise<Invoice>

  /**
   * Iterate every invoice matching the filter, paging automatically.
   *
   * @example
   * ```ts
   * for await (const inv of client.invoices.listAutoPaginate({ status: 'open' })) {
   *   console.log(inv.id)
   * }
   * ```
   */
  listAutoPaginate(
    params?: InvoiceListParams,
    options?: RequestOptions,
  ): AsyncIterableIterator<Invoice>
}

/**
 * Build an invoices service bound to an {@link HttpClient}.
 *
 * @internal
 */
export function invoicesService(http: HttpClient): InvoicesService {
  return {
    async list(
      params: InvoiceListParams = {},
      options?: RequestOptions,
    ): Promise<ListInvoicesResponse> {
      return http.request<ListInvoicesResponse>({
        method: 'GET',
        path: '/invoices',
        query: listParamsToQuery(params),
        options,
      })
    },

    async retrieve(
      id: string,
      params: InvoiceRetrieveParams = {},
      options?: RequestOptions,
    ): Promise<Invoice> {
      requireId('retrieve', id)
      const response = await http.request<DetailEnvelope<Invoice>>({
        method: 'GET',
        path: `/invoices/${encodeURIComponent(id)}`,
        query: retrieveParamsToQuery(params),
        options,
      })
      return response.data
    },

    async create(body: InvoiceCreateBody, options?: RequestOptions): Promise<Invoice> {
      const response = await http.request<DetailEnvelope<Invoice>>({
        method: 'POST',
        path: '/invoices',
        body,
        options,
      })
      return response.data
    },

    async update(id: string, body: InvoiceUpdateBody, options?: RequestOptions): Promise<Invoice> {
      requireId('update', id)
      const response = await http.request<DetailEnvelope<Invoice>>({
        method: 'PATCH',
        path: `/invoices/${encodeURIComponent(id)}`,
        body,
        options,
      })
      return response.data
    },

    async finalize(id: string, options?: RequestOptions): Promise<Invoice> {
      requireId('finalize', id)
      const response = await http.request<DetailEnvelope<Invoice>>({
        method: 'POST',
        path: `/invoices/${encodeURIComponent(id)}/finalize`,
        options,
      })
      return response.data
    },

    async void(id: string, body?: InvoiceVoidBody, options?: RequestOptions): Promise<Invoice> {
      requireId('void', id)
      const response = await http.request<DetailEnvelope<Invoice>>({
        method: 'POST',
        path: `/invoices/${encodeURIComponent(id)}/void`,
        body: body ?? {},
        options,
      })
      return response.data
    },

    async recordPayment(
      id: string,
      body: InvoicePaymentBody,
      options?: RequestOptions,
    ): Promise<Invoice> {
      requireId('recordPayment', id)
      const response = await http.request<DetailEnvelope<Invoice>>({
        method: 'POST',
        path: `/invoices/${encodeURIComponent(id)}/payments`,
        body,
        options,
      })
      return response.data
    },

    listAutoPaginate(
      params: InvoiceListParams = {},
      options?: RequestOptions,
    ): AsyncIterableIterator<Invoice> {
      const fetchPage = async (
        pageParams: InvoiceListParams & { page: number },
      ): Promise<{ data: readonly Invoice[]; hasMore: boolean }> => {
        const result = await http.request<ListInvoicesResponse>({
          method: 'GET',
          path: '/invoices',
          query: listParamsToQuery(pageParams),
          options,
        })
        return { data: result.data, hasMore: result.hasMore }
      }
      return createAutoPaginator<Invoice, InvoiceListParams>(fetchPage, params)
    },
  }
}

function requireId(method: string, id: string): void {
  if (typeof id !== 'string' || id.length === 0) {
    throw new TypeError(`invoices.${method}: \`id\` must be a non-empty string`)
  }
}

function listParamsToQuery(
  params: InvoiceListParams,
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

function retrieveParamsToQuery(params: InvoiceRetrieveParams): Record<string, string | undefined> {
  if (params.fields === undefined) return {}
  return { fields: params.fields }
}
