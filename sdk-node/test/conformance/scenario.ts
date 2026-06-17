/**
 * Shared scenario types for the conformance harness. The runner and each
 * per-resource dispatcher import from here so a new resource only needs a
 * sibling `dispatch-<resource>.ts` file plus a one-line edit to `dispatch`
 * in `runner.test.ts`.
 */

import type * as Errors from '@/errors'

type ExpectedHeaders = Readonly<Record<string, string>>

export interface ExpectedRequest {
  readonly method: 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE'
  readonly path: string
  readonly query?: Record<string, string>
  readonly headers?: ExpectedHeaders
  readonly headersAbsent?: readonly string[]
  readonly body?: Record<string, unknown> | null
}

export interface MockResponse {
  readonly status: number
  readonly headers?: Record<string, string>
  readonly body?: unknown
}

export interface ExpectedError {
  readonly type: keyof typeof Errors
  readonly code: string
  readonly httpStatus?: number
  readonly requestId?: string
  readonly retryAfterSeconds?: number
  readonly details?: Record<string, unknown>
}

export interface ClientOverrides {
  readonly apiVersion?: string
  readonly telemetry?: boolean
  readonly maxRetries?: number
}

export type Resource =
  | 'events'
  | 'invoices'
  | 'subscriptions'
  | 'contacts'
  | 'entitlements'
  | 'prices'
  | 'features'
  | 'forms'
  | 'properties'

export type Method =
  | 'list'
  | 'retrieve'
  | 'update'
  | 'create'
  | 'finalize'
  | 'void'
  | 'recordPayment'
  | 'autoCharge'
  | 'refundPayment'
  | 'deleteDraft'
  | 'activate'
  | 'cancel'
  | 'cancelImmediately'
  | 'markUnpaid'
  | 'bill'
  | 'renew'
  | 'previewUpcomingInvoice'
  | 'count'
  | 'delete'
  | 'bulkUpsert'
  | 'listActivity'
  | 'listActivityAutoPaginate'
  | 'listAutoPaginate'
  | 'lookup'
  | 'grant'
  | 'consume'
  | 'archive'
  | 'unarchive'
  | 'resolve'
  | 'duplicate'
  | 'addElement'
  | 'updateElement'
  | 'deleteElement'
  | 'moveElement'
  | 'enableOtherOption'
  | 'disableOtherOption'
  | 'addLogicRule'
  | 'removeLogicRule'
  | 'retrievePaymentMethod'
  | 'attachPaymentMethod'
  | 'createPaymentMethodSetupIntent'
  | 'removePaymentMethod'

export interface ScenarioCall {
  readonly resource?: Resource
  readonly method: Method
  readonly args: Record<string, unknown>
}

export interface Scenario {
  readonly name: string
  readonly call: ScenarioCall
  readonly client?: ClientOverrides
  readonly expectedRequest?: ExpectedRequest
  readonly mockResponse?: MockResponse
  readonly exchanges?: readonly { request: ExpectedRequest; response: MockResponse }[]
  readonly expectedResult?: unknown
  readonly expectedAutoPaginated?: readonly Record<string, unknown>[]
  readonly expectedError?: ExpectedError
  readonly expectedCallCount?: number
}

export function expectString(v: unknown, method: string): string {
  if (typeof v !== 'string') throw new Error(`${method} requires args.id (string)`)
  return v
}

export function expectBody(v: unknown, method: string): Record<string, unknown> {
  if (v === undefined || v === null || typeof v !== 'object') {
    throw new Error(`${method} requires args.body`)
  }
  return v as Record<string, unknown>
}
