import { resolveConfig } from '@/config'
import { HttpClient } from '@/core/http-client'
import { resolveFetch } from '@/core/platform'
import { Telemetry } from '@/core/telemetry'
import { contactsService, type ContactsService } from '@/resources/contacts'
import { entitlementsService, type EntitlementsService } from '@/resources/entitlements'
import { eventsService, type EventsService } from '@/resources/events'
import { invoicesService, type InvoicesService } from '@/resources/invoices'
import { pricesService, type PricesService } from '@/resources/prices'
import { subscriptionsService, type SubscriptionsService } from '@/resources/subscriptions'

import type { ClientConfig } from '@/types/public'

/**
 * Main entry point. One instance wraps a single API key + base URL and
 * exposes every resource as a property.
 *
 * @example
 * ```ts
 * import { ThreeCommon } from '@3common/sdk'
 *
 * const client = new ThreeCommon({ apiKey: process.env.THREECOMMON_API_KEY })
 * const { data, hasMore } = await client.events.list({ status: 'open' })
 * ```
 *
 * @public
 */
export class ThreeCommon {
  /** Events resource — `GET /v1/events`, `GET /v1/events/{id}`, `PATCH /v1/events/{id}`. */
  public readonly events: EventsService

  /** Invoices resource — list, retrieve, create, update, finalize, void, recordPayment. */
  public readonly invoices: InvoicesService

  /** Subscriptions resource — list, retrieve, create, update, activate, cancel, cancelImmediately, markUnpaid, bill, renew, previewUpcomingInvoice. */
  public readonly subscriptions: SubscriptionsService

  /** Contacts resource — list, count, retrieve, create, update, delete, bulkUpsert, listActivity, plus auto-paginators. */
  public readonly contacts: ContactsService

  /** Entitlements resource — list, retrieve, lookup, grant, consume, plus listAutoPaginate. */
  public readonly entitlements: EntitlementsService

  /** Prices resource — list, retrieve, create, update, archive, unarchive, plus listAutoPaginate. */
  public readonly prices: PricesService

  private readonly httpClient: HttpClient
  private readonly telemetry: Telemetry

  public constructor(config: ClientConfig = {}) {
    const resolved = resolveConfig(config)
    const fetchImpl = resolveFetch(resolved.fetch)
    this.telemetry = new Telemetry(resolved.telemetry)
    this.httpClient = new HttpClient({
      apiKey: resolved.apiKey,
      baseUrl: resolved.baseUrl,
      apiVersion: resolved.apiVersion,
      timeoutMs: resolved.timeoutMs,
      retry: {
        maxRetries: resolved.maxRetries,
        initialDelayMs: resolved.retryDelay.initialMs,
        maxDelayMs: resolved.retryDelay.maxMs,
        jitter: resolved.retryDelay.jitter,
      },
      fetch: fetchImpl,
      telemetry: this.telemetry,
      logger: resolved.logger,
    })

    this.events = eventsService(this.httpClient)
    this.invoices = invoicesService(this.httpClient)
    this.subscriptions = subscriptionsService(this.httpClient)
    this.contacts = contactsService(this.httpClient)
    this.entitlements = entitlementsService(this.httpClient)
    this.prices = pricesService(this.httpClient)
  }

  /**
   * Disable opt-out client telemetry at runtime. The next request and all
   * subsequent ones will omit the `Threecommon-Client-Telemetry` header.
   */
  public disableTelemetry(): void {
    this.telemetry.disable()
  }
}
