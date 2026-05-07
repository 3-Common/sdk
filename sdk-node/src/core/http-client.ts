import { API_PATH } from '@/api-version'
import { errorFromResponse, ThreeCommonConnectionError, ThreeCommonError } from '@/errors'
import { SDK_VERSION } from '@/version'

import { buildHeaders } from './headers'
import { parseRetryAfter, parseSuccessBody, tryParseJson } from './parse'
import { nowMs, userAgentSuffix } from './platform'
import {
  computeBackoffMs,
  isIdempotent,
  isRetryableStatus,
  type HttpMethod,
  type RetryPolicy,
} from './retry'
import { send } from './send'
import { buildUrl } from './url'

import type { Telemetry } from './telemetry'
import type { Logger, RequestOptions } from '@/types/public'

/**
 * Options for the {@link HttpClient}.
 *
 * @internal
 */
export interface HttpClientOptions {
  readonly apiKey: string
  readonly baseUrl: string
  readonly apiVersion: string
  readonly timeoutMs: number
  readonly retry: RetryPolicy
  readonly fetch: typeof fetch
  readonly telemetry: Telemetry
  readonly logger: Logger | undefined
}

/**
 * Internal request descriptor.
 *
 * @internal
 */
export interface InternalRequest {
  readonly method: HttpMethod
  readonly path: string
  readonly query?: Record<string, string | number | boolean | undefined>
  readonly body?: Record<string, unknown> | undefined
  readonly options?: RequestOptions | undefined
}

/**
 * Request orchestrator. Composes the pure modules in this folder into a
 * complete request lifecycle: build URL → build headers → send → parse →
 * map errors → retry on retryable failures.
 *
 * @internal
 */
export class HttpClient {
  public constructor(private readonly opts: HttpClientOptions) {}

  public async request<T>(req: InternalRequest): Promise<T> {
    const url = buildUrl({
      baseUrl: this.opts.baseUrl,
      apiPath: API_PATH,
      path: req.path,
      query: req.query,
    })
    const maxRetries = req.options?.maxRetries ?? this.opts.retry.maxRetries
    const idempotent = isIdempotent(req.method, req.options?.idempotencyKey !== undefined)

    let attempt = 0

    for (;;) {
      const start = nowMs()
      try {
        const response = await send({
          fetch: this.opts.fetch,
          url,
          method: req.method,
          headers: buildHeaders({
            apiKey: this.opts.apiKey,
            apiVersion: this.opts.apiVersion,
            sdkVersion: SDK_VERSION,
            userAgentSuffix: userAgentSuffix(),
            telemetryHeader: this.opts.telemetry.buildHeaderValue(),
            idempotencyKey: req.options?.idempotencyKey,
          }),
          body: req.body,
          timeoutMs: req.options?.timeoutMs ?? this.opts.timeoutMs,
          signal: req.options?.signal,
        })

        const durationMs = nowMs() - start
        this.opts.telemetry.record({
          method: req.method,
          path: req.path,
          status: response.status,
          durationMs,
          requestId: response.requestId,
        })
        this.opts.logger?.debug?.('threecommon:request', {
          method: req.method,
          path: req.path,
          status: response.status,
          durationMs,
          requestId: response.requestId,
          attempt,
        })

        if (response.status >= 200 && response.status < 300) {
          return parseSuccessBody(response) as T
        }

        const retryAfter = parseRetryAfter(response.headers.get('retry-after'))
        const error = errorFromResponse({
          status: response.status,
          body: tryParseJson(response.bodyText),
          rawResponse: response.bodyText,
          requestId: response.requestId,
          retryAfterSeconds: retryAfter,
        })

        if (idempotent && attempt < maxRetries && isRetryableStatus(response.status)) {
          await this.sleep(
            computeBackoffMs({ attempt, retryAfterSeconds: retryAfter, policy: this.opts.retry }),
          )
          attempt += 1
          continue
        }

        throw error
      } catch (err) {
        if (err instanceof ThreeCommonError) {
          throw err
        }

        // Network-level failure — retry idempotent requests.
        if (idempotent && attempt < maxRetries) {
          await this.sleep(
            computeBackoffMs({
              attempt,
              retryAfterSeconds: undefined,
              policy: this.opts.retry,
            }),
          )
          attempt += 1
          continue
        }

        throw new ThreeCommonConnectionError({
          code: 'connection_error',
          message: errorMessage(err) ?? 'Request failed before reaching the server',
          requestId: undefined,
          details: undefined,
          cause: err,
        })
      }
    }
  }

  private async sleep(ms: number): Promise<void> {
    if (ms <= 0) return
    await new Promise<void>((resolve) => setTimeout(resolve, ms))
  }
}

function errorMessage(err: unknown): string | undefined {
  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err
  return undefined
}
