import { createAutoPaginator } from '@/pagination/auto-paginator'

import type {
  BillSubscriptionResult,
  ListSubscriptionsResponse,
  RenewSubscriptionResult,
  Subscription,
  SubscriptionCancelBody,
  SubscriptionCancelImmediatelyBody,
  SubscriptionCreateBody,
  SubscriptionInvoicePreview,
  SubscriptionListParams,
  SubscriptionManageUrl,
  SubscriptionRetrieveParams,
  SubscriptionUpdateBody,
  UpdateSubscriptionResult,
} from './types'
import type { HttpClient } from '@/core/http-client'
import type { RequestOptions } from '@/types/public'

interface DetailEnvelope<T> {
  readonly data: T
}

interface BillEnvelope {
  readonly data: Subscription
  readonly invoice: BillSubscriptionResult['invoice']
}

interface RenewEnvelope {
  readonly data: Subscription
  readonly invoice?: RenewSubscriptionResult['invoice']
}

interface UpdateEnvelope {
  readonly data: Subscription
  readonly invoice?: UpdateSubscriptionResult['invoice']
  readonly proration: UpdateSubscriptionResult['proration']
}

interface PreviewEnvelope {
  readonly data: { readonly invoice: SubscriptionInvoicePreview | null }
}

/**
 * Subscriptions service. Bound as `client.subscriptions` on the main client.
 *
 * @public
 */
export interface SubscriptionsService {
  /**
   * List the authenticated host's subscriptions.
   *
   * @example
   * ```ts
   * const { data, hasMore } = await client.subscriptions.list({ status: 'active', pageSize: 50 })
   * ```
   */
  list(
    params?: SubscriptionListParams,
    options?: RequestOptions,
  ): Promise<ListSubscriptionsResponse>

  /**
   * Retrieve a single subscription by ID.
   *
   * @example
   * ```ts
   * const sub = await client.subscriptions.retrieve('sub_123')
   * ```
   */
  retrieve(
    id: string,
    params?: SubscriptionRetrieveParams,
    options?: RequestOptions,
  ): Promise<Subscription>

  /**
   * Create a new subscription against an active recurring Price. Starts in
   * `trialing` if `trialDays` is set, else `incomplete` (awaiting first
   * payment).
   *
   * @example
   * ```ts
   * const sub = await client.subscriptions.create({
   *   priceId: 'price_42',
   *   contactId: 'cnt_7',
   *   trialDays: 14,
   * })
   * ```
   */
  create(body: SubscriptionCreateBody, options?: RequestOptions): Promise<Subscription>

  /**
   * Apply a mid-cycle price/quantity change with Stripe-style daily
   * proration, or flip forward-looking settings (notes, taxIds, taxRate,
   * autoCharge, dunningEnabled, paymentDueDays).
   *
   * @example
   * ```ts
   * const { subscription, invoice, proration } = await client.subscriptions.update(
   *   'sub_123',
   *   { priceId: 'price_upgrade', quantity: 2 },
   * )
   * ```
   */
  update(
    id: string,
    body: SubscriptionUpdateBody,
    options?: RequestOptions,
  ): Promise<UpdateSubscriptionResult>

  /**
   * Fetch a signed, customer-facing self-service portal URL for the
   * subscription. The link is scoped to this one subscription — share it with
   * the subscriber so they can view, cancel, or resume it.
   *
   * @example
   * ```ts
   * const { url } = await client.subscriptions.retrieveManageUrl('sub_123')
   * ```
   */
  retrieveManageUrl(id: string, options?: RequestOptions): Promise<SubscriptionManageUrl>

  /**
   * Transition an incomplete or trialing subscription to `active`.
   *
   * @example
   * ```ts
   * await client.subscriptions.activate('sub_123')
   * ```
   */
  activate(id: string, options?: RequestOptions): Promise<Subscription>

  /**
   * Schedule cancellation at the end of the current period. Idempotent.
   *
   * @example
   * ```ts
   * await client.subscriptions.cancel('sub_123', { reason: 'Customer churned' })
   * ```
   */
  cancel(id: string, body?: SubscriptionCancelBody, options?: RequestOptions): Promise<Subscription>

  /**
   * Admin override — terminate the subscription immediately (status
   * `canceled`, `endedAt` stamped).
   *
   * @example
   * ```ts
   * await client.subscriptions.cancelImmediately('sub_123', { reason: 'Fraud' })
   * ```
   */
  cancelImmediately(
    id: string,
    body?: SubscriptionCancelImmediatelyBody,
    options?: RequestOptions,
  ): Promise<Subscription>

  /**
   * Stage a one-time fully-free (100% off) next renewal cycle. The next
   * renewal consumes the comp exactly once, then billing resumes at full
   * price. Rejected on a `canceled` or `unpaid` subscription.
   *
   * @example
   * ```ts
   * await client.subscriptions.compNextCycle('sub_123')
   * ```
   */
  compNextCycle(id: string, options?: RequestOptions): Promise<Subscription>

  /**
   * Remove a staged comp so the next renewal bills at full price again — the
   * inverse of `compNextCycle`. A no-op when no comp is pending, and allowed
   * on a subscription in any state.
   *
   * @example
   * ```ts
   * await client.subscriptions.uncompNextCycle('sub_123')
   * ```
   */
  uncompNextCycle(id: string, options?: RequestOptions): Promise<Subscription>

  /**
   * Admin override — mark a subscription `unpaid` (terminal), bypassing
   * dunning retries.
   *
   * @example
   * ```ts
   * await client.subscriptions.markUnpaid('sub_123')
   * ```
   */
  markUnpaid(id: string, options?: RequestOptions): Promise<Subscription>

  /**
   * Generate a draft invoice for the subscription's current period without
   * advancing the period.
   *
   * @example
   * ```ts
   * const { subscription, invoice } = await client.subscriptions.bill('sub_123')
   * ```
   */
  bill(id: string, options?: RequestOptions): Promise<BillSubscriptionResult>

  /**
   * Advance the subscription to its next billing period and generate an
   * invoice. Transitions to `canceled` instead when `cancelAtPeriodEnd` was
   * set.
   *
   * @example
   * ```ts
   * const { subscription, invoice } = await client.subscriptions.renew('sub_123')
   * ```
   */
  renew(id: string, options?: RequestOptions): Promise<RenewSubscriptionResult>

  /**
   * Return a non-persisted preview of the invoice the next renewal will
   * generate. Returns `null` when the subscription is set to cancel at
   * period end.
   *
   * @example
   * ```ts
   * const preview = await client.subscriptions.previewUpcomingInvoice('sub_123')
   * ```
   */
  previewUpcomingInvoice(
    id: string,
    options?: RequestOptions,
  ): Promise<SubscriptionInvoicePreview | null>

  /**
   * Iterate every subscription matching the filter, paging automatically.
   *
   * @example
   * ```ts
   * for await (const sub of client.subscriptions.listAutoPaginate({ status: 'active' })) {
   *   console.log(sub.id)
   * }
   * ```
   */
  listAutoPaginate(
    params?: SubscriptionListParams,
    options?: RequestOptions,
  ): AsyncIterableIterator<Subscription>
}

/**
 * Build a subscriptions service bound to an {@link HttpClient}.
 *
 * @internal
 */
export function subscriptionsService(http: HttpClient): SubscriptionsService {
  return {
    async list(
      params: SubscriptionListParams = {},
      options?: RequestOptions,
    ): Promise<ListSubscriptionsResponse> {
      return http.request<ListSubscriptionsResponse>({
        method: 'GET',
        path: '/subscriptions',
        query: listParamsToQuery(params),
        options,
      })
    },

    async retrieve(
      id: string,
      params: SubscriptionRetrieveParams = {},
      options?: RequestOptions,
    ): Promise<Subscription> {
      requireId('retrieve', id)
      const response = await http.request<DetailEnvelope<Subscription>>({
        method: 'GET',
        path: `/subscriptions/${encodeURIComponent(id)}`,
        query: retrieveParamsToQuery(params),
        options,
      })
      return response.data
    },

    async create(body: SubscriptionCreateBody, options?: RequestOptions): Promise<Subscription> {
      const response = await http.request<DetailEnvelope<Subscription>>({
        method: 'POST',
        path: '/subscriptions',
        body,
        options,
      })
      return response.data
    },

    async update(
      id: string,
      body: SubscriptionUpdateBody,
      options?: RequestOptions,
    ): Promise<UpdateSubscriptionResult> {
      requireId('update', id)
      const response = await http.request<UpdateEnvelope>({
        method: 'PATCH',
        path: `/subscriptions/${encodeURIComponent(id)}`,
        body,
        options,
      })
      return response.invoice === undefined
        ? { subscription: response.data, proration: response.proration }
        : { subscription: response.data, invoice: response.invoice, proration: response.proration }
    },

    async retrieveManageUrl(id: string, options?: RequestOptions): Promise<SubscriptionManageUrl> {
      requireId('retrieveManageUrl', id)
      const response = await http.request<DetailEnvelope<SubscriptionManageUrl>>({
        method: 'GET',
        path: `/subscriptions/${encodeURIComponent(id)}/manage-url`,
        options,
      })
      return response.data
    },

    async activate(id: string, options?: RequestOptions): Promise<Subscription> {
      requireId('activate', id)
      const response = await http.request<DetailEnvelope<Subscription>>({
        method: 'POST',
        path: `/subscriptions/${encodeURIComponent(id)}/activate`,
        options,
      })
      return response.data
    },

    async cancel(
      id: string,
      body?: SubscriptionCancelBody,
      options?: RequestOptions,
    ): Promise<Subscription> {
      requireId('cancel', id)
      const response = await http.request<DetailEnvelope<Subscription>>({
        method: 'POST',
        path: `/subscriptions/${encodeURIComponent(id)}/cancel`,
        body: body ?? {},
        options,
      })
      return response.data
    },

    async cancelImmediately(
      id: string,
      body?: SubscriptionCancelImmediatelyBody,
      options?: RequestOptions,
    ): Promise<Subscription> {
      requireId('cancelImmediately', id)
      const response = await http.request<DetailEnvelope<Subscription>>({
        method: 'POST',
        path: `/subscriptions/${encodeURIComponent(id)}/cancel-immediately`,
        body: body ?? {},
        options,
      })
      return response.data
    },

    async compNextCycle(id: string, options?: RequestOptions): Promise<Subscription> {
      requireId('compNextCycle', id)
      const response = await http.request<DetailEnvelope<Subscription>>({
        method: 'POST',
        path: `/subscriptions/${encodeURIComponent(id)}/comp-next-cycle`,
        options,
      })
      return response.data
    },

    async uncompNextCycle(id: string, options?: RequestOptions): Promise<Subscription> {
      requireId('uncompNextCycle', id)
      const response = await http.request<DetailEnvelope<Subscription>>({
        method: 'POST',
        path: `/subscriptions/${encodeURIComponent(id)}/uncomp-next-cycle`,
        options,
      })
      return response.data
    },

    async markUnpaid(id: string, options?: RequestOptions): Promise<Subscription> {
      requireId('markUnpaid', id)
      const response = await http.request<DetailEnvelope<Subscription>>({
        method: 'POST',
        path: `/subscriptions/${encodeURIComponent(id)}/mark-unpaid`,
        options,
      })
      return response.data
    },

    async bill(id: string, options?: RequestOptions): Promise<BillSubscriptionResult> {
      requireId('bill', id)
      const response = await http.request<BillEnvelope>({
        method: 'POST',
        path: `/subscriptions/${encodeURIComponent(id)}/bill`,
        options,
      })
      return { subscription: response.data, invoice: response.invoice }
    },

    async renew(id: string, options?: RequestOptions): Promise<RenewSubscriptionResult> {
      requireId('renew', id)
      const response = await http.request<RenewEnvelope>({
        method: 'POST',
        path: `/subscriptions/${encodeURIComponent(id)}/renew`,
        options,
      })
      return response.invoice === undefined
        ? { subscription: response.data }
        : { subscription: response.data, invoice: response.invoice }
    },

    async previewUpcomingInvoice(
      id: string,
      options?: RequestOptions,
    ): Promise<SubscriptionInvoicePreview | null> {
      requireId('previewUpcomingInvoice', id)
      const response = await http.request<PreviewEnvelope>({
        method: 'GET',
        path: `/subscriptions/${encodeURIComponent(id)}/upcoming`,
        options,
      })
      return response.data.invoice
    },

    listAutoPaginate(
      params: SubscriptionListParams = {},
      options?: RequestOptions,
    ): AsyncIterableIterator<Subscription> {
      const fetchPage = async (
        pageParams: SubscriptionListParams & { page: number },
      ): Promise<{ data: readonly Subscription[]; hasMore: boolean }> => {
        const result = await http.request<ListSubscriptionsResponse>({
          method: 'GET',
          path: '/subscriptions',
          query: listParamsToQuery(pageParams),
          options,
        })
        return { data: result.data, hasMore: result.hasMore }
      }
      return createAutoPaginator<Subscription, SubscriptionListParams>(fetchPage, params)
    },
  }
}

function requireId(method: string, id: string): void {
  if (typeof id !== 'string' || id.length === 0) {
    throw new TypeError(`subscriptions.${method}: \`id\` must be a non-empty string`)
  }
}

function listParamsToQuery(
  params: SubscriptionListParams,
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

function retrieveParamsToQuery(
  params: SubscriptionRetrieveParams,
): Record<string, string | undefined> {
  if (params.fields === undefined) return {}
  return { fields: params.fields }
}
