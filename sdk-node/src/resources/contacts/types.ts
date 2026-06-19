/**
 * Public types for the contacts resource. Hand-curated friendly aliases over
 * the auto-generated OpenAPI types.
 *
 * @public
 */

import type { paths } from '@/generated/types'

/**
 * Lifecycle status of a contact.
 *
 * - `opted-in` / `unsubscribed`: explicit consent state
 * - `unknown`: never recorded a choice
 * - `imported`: created via CSV / bulk-upsert before consent was captured
 * - `deleted`: soft-deleted
 */
export type ContactStatus = 'deleted' | 'imported' | 'unsubscribed' | 'opted-in' | 'unknown'

/**
 * How to resolve field-level conflicts when merging a second contact into the
 * target during {@link ContactsService.update}.
 *
 * - `safe-merge`: only fill fields that are empty on the target
 * - `overwrite-merge`: target fields are replaced by source fields
 */
export type ContactMergeResolution = 'safe-merge' | 'overwrite-merge'

/**
 * The kind of event recorded against a contact in their activity feed.
 */
export type ContactActivityType =
  | 'checkout_session_completed'
  | 'product_set_checkout_session_completed'
  | 'order_refunded'
  | 'ticket_scanned'
  | 'email_sent'
  | 'invoice_paid'

/**
 * A contact in the compact projection returned by `list`, `retrieve`, and
 * `create`. Custom-property keys (24-char hex ids) may appear as additional
 * top-level fields.
 */
export type Contact = NonNullable<
  paths['/v1/contacts/{id}']['get']['responses'][200]['content']['application/json']['data']
>

/**
 * The richer "order-details" projection returned by `update`. Includes raw
 * `events_attended` / `items_purchased` / `products_purchased` arrays and the
 * `properties` array, in addition to everything on {@link Contact}.
 */
export type ContactWithOrderDetails = NonNullable<
  paths['/v1/contacts/{id}']['patch']['responses'][200]['content']['application/json']['data']
>

/** A single activity record in a contact's activity feed. */
export type ContactActivity = NonNullable<
  paths['/v1/contacts/{id}/activity']['get']['responses'][200]['content']['application/json']['data']
>[number]

/** Successful response shape from `GET /v1/contacts`. */
export interface ListContactsResponse {
  readonly data: readonly Contact[]
  readonly hasMore: boolean
  readonly pageNumber: number
  readonly pageSize: number
}

/** Successful response shape from `GET /v1/contacts/{id}/activity`. */
export interface ListContactActivityResponse {
  readonly data: readonly ContactActivity[]
  readonly hasMore: boolean
  readonly pageNumber: number
  readonly pageSize: number
}

/** Result of `GET /v1/contacts/count`. */
export interface ContactCountResult {
  readonly count: number
}

/** Result of `POST /v1/contacts/bulk`. */
export interface BulkUpsertContactsResult {
  /** Number of contacts inserted or updated. */
  readonly affected: number
}

/** Result of `DELETE /v1/contacts/{id}` — the id of the removed contact. */
export type DeletedContact =
  paths['/v1/contacts/{id}']['delete']['responses'][200]['content']['application/json']['data']

/** Query parameters accepted by `GET /v1/contacts`. */
export interface ContactListParams {
  /** Page number, 0-indexed. Default 0. */
  readonly pageNumber?: number
  /** Items per page. Default 20, max 500. */
  readonly pageSize?: number
  /**
   * Field to sort by. Index-backed values: `mostRecentOrder`, `orderSum`,
   * `grossSum`. Other field names and 24-char hex custom-property ids are
   * accepted and sorted in-memory. Defaults to `mostRecentOrder` desc.
   */
  readonly sortField?: string
  /** Sort direction. Defaults to `desc` when `sortField` is provided. */
  readonly sortDirection?: 'asc' | 'desc'
  /**
   * Quick status filter; case-insensitive. `"all"` or omit for no filter.
   * ANDed with `filters` when both are supplied.
   */
  readonly filter?: 'all' | 'opted-in' | 'unknown' | 'unsubscribed' | 'imported'
  /**
   * JSON-encoded `FilterGroup[]` for advanced filtering. Build with the typed
   * `filter` helper from `@3common/sdk` rather than writing the JSON by hand.
   *
   * @example
   * ```ts
   * import { filter } from '@3common/sdk'
   *
   * await client.contacts.list({
   *   filters: filter.and(
   *     filter.field('status').isAnyOf(['opted-in']),
   *     filter.field('grossSum').isGreaterThan(1000),
   *   ).serialize(),
   * })
   * ```
   */
  readonly filters?: string
  /** Free-text search over email, firstName, lastName, fullName. */
  readonly search?: string
}

/** Query parameters accepted by `GET /v1/contacts/{id}/activity`. */
export interface ContactActivityListParams {
  /** Page number, 0-indexed. Default 0. */
  readonly pageNumber?: number
  /** Items per page. Default 20, max 50. */
  readonly pageSize?: number
  /** Restrict to a single activity type. */
  readonly filter?: ContactActivityType
  /** Default is newest-first; `"oldest"` reverses. */
  readonly sort?: 'oldest'
}

/** Body accepted by `POST /v1/contacts`. */
export type ContactCreateBody =
  paths['/v1/contacts/']['post']['requestBody']['content']['application/json']

/** Body accepted by `PATCH /v1/contacts/{id}`. */
export type ContactUpdateBody =
  paths['/v1/contacts/{id}']['patch']['requestBody']['content']['application/json']

/** Body accepted by `POST /v1/contacts/bulk`. */
export type ContactBulkUpsertBody =
  paths['/v1/contacts/bulk']['post']['requestBody']['content']['application/json']

/**
 * Lifecycle status of a saved payment method.
 *
 * - `active`: usable card on file
 * - `detached`: removed from Stripe and the contact
 * - `expired`: past its expiry date
 */
export type PaymentMethodStatus = 'active' | 'detached' | 'expired'

/**
 * A saved card on file for a contact, returned by
 * {@link ContactsService.retrievePaymentMethod} and
 * {@link ContactsService.attachPaymentMethod}. One card is supported per
 * contact. The raw card number never touches our servers — only the brand,
 * last-4, expiry, and billing metadata Stripe returns are stored.
 */
export type PaymentMethod = NonNullable<
  paths['/v1/contacts/{id}/payment-methods']['get']['responses'][200]['content']['application/json']['data']
>

/** Body accepted by `POST /v1/contacts/{id}/payment-methods`. */
export type AttachPaymentMethodBody =
  paths['/v1/contacts/{id}/payment-methods']['post']['requestBody']['content']['application/json']

/** Result of `POST /v1/contacts/{id}/payment-methods`. */
export interface AttachPaymentMethodResult {
  /** The newly saved payment method. */
  readonly data: PaymentMethod
  /** `true` when this card replaced an existing card on file for the contact. */
  readonly replacedExisting: boolean
}

/**
 * Result of `POST /v1/contacts/{id}/payment-methods/setup-intent` — the Stripe
 * SetupIntent to confirm client-side with Stripe Elements before attaching.
 */
export type PaymentMethodSetupIntent =
  paths['/v1/contacts/{id}/payment-methods/setup-intent']['post']['responses'][200]['content']['application/json']['data']

/** Result of `DELETE /v1/contacts/{id}/payment-methods/{methodId}`. */
export type RemovedPaymentMethod =
  paths['/v1/contacts/{id}/payment-methods/{methodId}']['delete']['responses'][200]['content']['application/json']['data']
