/**
 * Public types for the invoices resource. Hand-curated friendly aliases over
 * the auto-generated OpenAPI types.
 *
 * @public
 */

import type { paths } from '@/generated/types'

/** Lifecycle status of an invoice. */
export type InvoiceStatus = 'draft' | 'open' | 'paid' | 'void'

/** Currency code used on an invoice; all line amounts must match. */
export type InvoiceCurrency = 'USD' | 'CAD'

/**
 * One invoice as returned by the API. Detail responses populate every field
 * the server has for that record; list responses with a `fields` filter only
 * include the requested values.
 */
export type Invoice = NonNullable<
  paths['/v1/invoices/{id}']['get']['responses'][200]['content']['application/json']['data']
>

/** One line item on an invoice. */
export type InvoiceLineItem = NonNullable<Invoice['lineItems']>[number]

/** One recorded payment against an invoice. */
export type InvoicePayment = NonNullable<Invoice['payments']>[number]

/** Successful response shape from `GET /v1/invoices`. */
export interface ListInvoicesResponse {
  readonly data: readonly Invoice[]
  readonly hasMore: boolean
}

/** Query parameters accepted by `GET /v1/invoices`. */
export interface InvoiceListParams {
  /** Page number, 0-indexed. Default 0. */
  readonly page?: number
  /** Items per page. Default 20, max 50. */
  readonly pageSize?: number
  /** Filter by lifecycle status; omit to include all. */
  readonly status?: InvoiceStatus
  /** Filter by recipient contact id. */
  readonly customerId?: string
  /** ISO 8601; only invoices issued on or after this date. */
  readonly issuedAfter?: string
  /** ISO 8601; only invoices issued on or before this date. */
  readonly issuedBefore?: string
  /** Comma-separated list of fields to include in the response. Omit for all fields. */
  readonly fields?: string
}

/** Query parameters accepted by `GET /v1/invoices/{id}`. */
export interface InvoiceRetrieveParams {
  /** Comma-separated list of fields to include in the response. */
  readonly fields?: string
}

/** Body accepted by `POST /v1/invoices`. */
export type InvoiceCreateBody =
  paths['/v1/invoices/']['post']['requestBody']['content']['application/json']

/** Body accepted by `PATCH /v1/invoices/{id}`. Only fields you provide are changed. */
export type InvoiceUpdateBody =
  paths['/v1/invoices/{id}']['patch']['requestBody']['content']['application/json']

/** Body accepted by `POST /v1/invoices/{id}/void`. */
export type InvoiceVoidBody = NonNullable<
  paths['/v1/invoices/{id}/void']['post']['requestBody']
>['content']['application/json']

/** Body accepted by `POST /v1/invoices/{id}/payments`. */
export type InvoicePaymentBody =
  paths['/v1/invoices/{id}/payments']['post']['requestBody']['content']['application/json']
