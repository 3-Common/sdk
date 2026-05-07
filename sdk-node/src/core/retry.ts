/**
 * Retry policy and backoff math. Pure module — no I/O, no timing primitives.
 *
 * @internal
 */

export type HttpMethod = 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE'

export interface RetryPolicy {
  readonly maxRetries: number
  readonly initialDelayMs: number
  readonly maxDelayMs: number
  readonly jitter: boolean
}

export const RETRYABLE_STATUS_CODES: ReadonlySet<number> = new Set([
  408, 425, 429, 500, 502, 503, 504,
])

const IDEMPOTENT_METHODS: ReadonlySet<HttpMethod> = new Set(['GET', 'PATCH', 'PUT'])

/**
 * `true` if the SDK may safely retry this call. `PATCH`/`PUT` are treated as
 * idempotent because the API server marks events.update with `idempotentHint`.
 * `POST` becomes idempotent only when the caller supplies an idempotency key.
 */
export function isIdempotent(method: HttpMethod, hasIdempotencyKey: boolean): boolean {
  return IDEMPOTENT_METHODS.has(method) || hasIdempotencyKey
}

/** `true` if the HTTP status is one we should retry on (alongside idempotency). */
export function isRetryableStatus(status: number): boolean {
  return RETRYABLE_STATUS_CODES.has(status)
}

/**
 * Compute the next backoff delay in milliseconds.
 *
 * - When `retryAfterSeconds` is supplied (e.g. from a `Retry-After` header),
 *   honor it but cap at the policy's `maxDelayMs`.
 * - Otherwise: exponential backoff with optional full jitter.
 */
export function computeBackoffMs(args: {
  readonly attempt: number
  readonly retryAfterSeconds: number | undefined
  readonly policy: RetryPolicy
}): number {
  if (args.retryAfterSeconds !== undefined && Number.isFinite(args.retryAfterSeconds)) {
    return Math.min(args.retryAfterSeconds * 1000, args.policy.maxDelayMs)
  }
  const exponential = args.policy.initialDelayMs * 2 ** args.attempt
  const capped = Math.min(exponential, args.policy.maxDelayMs)
  if (!args.policy.jitter) return capped
  return Math.floor(Math.random() * capped)
}
