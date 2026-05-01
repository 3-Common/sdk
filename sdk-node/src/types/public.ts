/**
 * Public configuration and option types. Re-exported from the package root.
 *
 * @public
 */

/**
 * Minimal logger interface — compatible with `console`, `pino`, `winston`,
 * `bunyan`, and most other Node loggers.
 *
 * @public
 */
export interface Logger {
  debug?: (message: string, meta?: Record<string, unknown>) => void
  info?: (message: string, meta?: Record<string, unknown>) => void
  warn?: (message: string, meta?: Record<string, unknown>) => void
  error?: (message: string, meta?: Record<string, unknown>) => void
}

/**
 * Public client configuration.
 *
 * @public
 */
export interface ClientConfig {
  /**
   * 3Common API key. Required unless the `THREECOMMON_API_KEY` environment
   * variable is set. Generate keys in the 3Common organizer dashboard
   * (Settings → API Keys).
   */
  apiKey?: string

  /**
   * Base URL for the API. Defaults to `https://api.3common.com`. The SDK
   * appends `/v1/...` to every request path.
   */
  baseUrl?: string

  /**
   * Date-stamped version of the API surface to target. Sent as the
   * `Threecommon-Version` header so server-side behavior changes don't
   * silently break SDKs pinned to a previous date.
   */
  apiVersion?: string

  /** Per-request timeout in milliseconds. Default 30_000. */
  timeoutMs?: number

  /**
   * Maximum number of automatic retries for retriable failures (network
   * errors, 408/425/429, 5xx on idempotent methods). Default 3.
   */
  maxRetries?: number

  /** Retry-delay parameters. Exponential backoff with optional full jitter. */
  retryDelay?: {
    initialMs?: number
    maxMs?: number
    jitter?: boolean
  }

  /** Override the `fetch` implementation. Useful for proxies and tests. */
  fetch?: typeof fetch

  /** Optional logger. The SDK calls `debug` on every request when set. */
  logger?: Logger

  /**
   * Disable opt-out client telemetry. Default `true` (telemetry enabled).
   */
  telemetry?: boolean
}

/**
 * Per-request override options.
 *
 * @public
 */
export interface RequestOptions {
  /** Override the configured timeout for this call only. */
  readonly timeoutMs?: number
  /** Override the configured maxRetries for this call only. */
  readonly maxRetries?: number
  /** Caller-provided AbortSignal. The SDK still applies its own timeout. */
  readonly signal?: AbortSignal
  /** Idempotency key (forward-compat — no v1 endpoints require it). */
  readonly idempotencyKey?: string
}
