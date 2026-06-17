/**
 * Public types for the subscriptions resource. Hand-curated friendly aliases
 * over the auto-generated OpenAPI types.
 *
 * @public
 */

import type { paths } from '@/generated/types'

/** Lifecycle status of a subscription. */
export type SubscriptionStatus =
  | 'incomplete'
  | 'trialing'
  | 'active'
  | 'past_due'
  | 'canceled'
  | 'unpaid'

/**
 * One subscription as returned by the API. Detail responses populate every
 * field the server has for that record; list responses with a `fields`
 * filter only include the requested values.
 */
export type Subscription = NonNullable<
  paths['/v1/subscriptions/{id}']['get']['responses'][200]['content']['application/json']['data']
>

/** One billed item on a subscription. */
export type SubscriptionItem = NonNullable<Subscription['items']>[number]

/** Host tax-ID snapshot carried onto each renewal invoice. */
export type SubscriptionTaxId = NonNullable<Subscription['taxIds']>[number]

/** Slim invoice reference returned alongside renew/bill/update responses. */
export interface SubscriptionInvoiceRef {
  readonly id: string
  readonly status: string
  readonly total: number
  readonly currency: string
}

/** Proration summary returned by `PATCH /v1/subscriptions/{id}`. */
export type SubscriptionProration =
  paths['/v1/subscriptions/{id}']['patch']['responses'][200]['content']['application/json']['proration']

/**
 * Non-persisted projection of the invoice the next renewal will generate
 * (Stripe-style `invoice.upcoming`).
 */
export type SubscriptionInvoicePreview = NonNullable<
  paths['/v1/subscriptions/{id}/upcoming']['get']['responses'][200]['content']['application/json']['data']['invoice']
>

/** One line item on a subscription invoice preview. */
export type SubscriptionInvoicePreviewLineItem = SubscriptionInvoicePreview['lineItems'][number]

/**
 * Signed, customer-facing self-service portal link for a subscription,
 * returned by `GET /v1/subscriptions/{id}/manage-url`. The link is scoped to
 * the one subscription; share it with the subscriber so they can view, cancel,
 * or resume it.
 */
export type SubscriptionManageUrl =
  paths['/v1/subscriptions/{id}/manage-url']['get']['responses'][200]['content']['application/json']['data']

/** Successful response shape from `GET /v1/subscriptions`. */
export interface ListSubscriptionsResponse {
  readonly data: readonly Subscription[]
  readonly hasMore: boolean
}

/** Successful response shape from `PATCH /v1/subscriptions/{id}`. */
export interface UpdateSubscriptionResult {
  readonly subscription: Subscription
  /** Present only when proration produced a positive amount. */
  readonly invoice?: SubscriptionInvoiceRef
  readonly proration: SubscriptionProration
}

/** Successful response shape from `POST /v1/subscriptions/{id}/bill`. */
export interface BillSubscriptionResult {
  readonly subscription: Subscription
  readonly invoice: SubscriptionInvoiceRef
}

/** Successful response shape from `POST /v1/subscriptions/{id}/renew`. */
export interface RenewSubscriptionResult {
  readonly subscription: Subscription
  /** Generated only when the renewal advanced the period. */
  readonly invoice?: SubscriptionInvoiceRef
}

/** Query parameters accepted by `GET /v1/subscriptions`. */
export interface SubscriptionListParams {
  /** Page number, 0-indexed. Default 0. */
  readonly page?: number
  /** Items per page. */
  readonly pageSize?: number
  /** Filter by lifecycle status; omit to include all. */
  readonly status?: SubscriptionStatus
  /** Filter by recipient contact id. */
  readonly contactId?: string
  /** Filter by Price reference. */
  readonly priceId?: string
  /** Comma-separated list of fields to include in the response. Omit for all fields. */
  readonly fields?: string
}

/** Query parameters accepted by `GET /v1/subscriptions/{id}`. */
export interface SubscriptionRetrieveParams {
  /** Comma-separated list of fields to include in the response. */
  readonly fields?: string
}

/** Body accepted by `POST /v1/subscriptions`. */
export type SubscriptionCreateBody =
  paths['/v1/subscriptions/']['post']['requestBody']['content']['application/json']

/** Body accepted by `PATCH /v1/subscriptions/{id}`. Only fields you provide are changed. */
export type SubscriptionUpdateBody =
  paths['/v1/subscriptions/{id}']['patch']['requestBody']['content']['application/json']

/** Body accepted by `POST /v1/subscriptions/{id}/cancel`. */
export type SubscriptionCancelBody = NonNullable<
  paths['/v1/subscriptions/{id}/cancel']['post']['requestBody']
>['content']['application/json']

/** Body accepted by `POST /v1/subscriptions/{id}/cancel-immediately`. */
export type SubscriptionCancelImmediatelyBody = NonNullable<
  paths['/v1/subscriptions/{id}/cancel-immediately']['post']['requestBody']
>['content']['application/json']
