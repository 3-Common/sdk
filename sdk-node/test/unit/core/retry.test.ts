import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import {
  computeBackoffMs,
  isIdempotent,
  isRetryableStatus,
  RETRYABLE_STATUS_CODES,
} from '@/core/retry'

describe('isIdempotent', () => {
  it.each(['GET', 'PATCH', 'PUT'] as const)('treats %s as idempotent without a key', (method) => {
    expect(isIdempotent(method, false)).toBe(true)
  })

  it.each(['POST', 'DELETE'] as const)('treats %s as non-idempotent without a key', (method) => {
    expect(isIdempotent(method, false)).toBe(false)
  })

  it('treats POST as idempotent when an idempotency key is supplied', () => {
    expect(isIdempotent('POST', true)).toBe(true)
  })
})

describe('isRetryableStatus', () => {
  it.each([...RETRYABLE_STATUS_CODES])('returns true for %s', (status) => {
    expect(isRetryableStatus(status)).toBe(true)
  })

  it.each([200, 201, 400, 401, 404, 422])('returns false for %s', (status) => {
    expect(isRetryableStatus(status)).toBe(false)
  })
})

describe('computeBackoffMs', () => {
  const policy = { maxRetries: 5, initialDelayMs: 100, maxDelayMs: 8000, jitter: false }

  it('returns exponential backoff without jitter', () => {
    expect(computeBackoffMs({ attempt: 0, retryAfterSeconds: undefined, policy })).toBe(100)
    expect(computeBackoffMs({ attempt: 1, retryAfterSeconds: undefined, policy })).toBe(200)
    expect(computeBackoffMs({ attempt: 2, retryAfterSeconds: undefined, policy })).toBe(400)
  })

  it('caps at maxDelayMs', () => {
    expect(computeBackoffMs({ attempt: 20, retryAfterSeconds: undefined, policy })).toBe(8000)
  })

  describe('with jitter', () => {
    const jitterPolicy = { ...policy, jitter: true }

    beforeEach(() => {
      vi.spyOn(Math, 'random').mockReturnValue(0.5)
    })
    afterEach(() => {
      vi.restoreAllMocks()
    })

    it('returns a jittered value bounded by the capped exponential', () => {
      // attempt=2 → exponential 400; jittered = floor(0.5 * 400) = 200
      expect(
        computeBackoffMs({ attempt: 2, retryAfterSeconds: undefined, policy: jitterPolicy }),
      ).toBe(200)
    })

    it('respects maxDelayMs after jittering', () => {
      // attempt=20 → capped 8000; jittered = floor(0.5 * 8000) = 4000
      expect(
        computeBackoffMs({ attempt: 20, retryAfterSeconds: undefined, policy: jitterPolicy }),
      ).toBe(4000)
    })
  })

  it('honors retryAfterSeconds, converting to ms and capping', () => {
    expect(computeBackoffMs({ attempt: 0, retryAfterSeconds: 2, policy })).toBe(2000)
    // Larger than maxDelayMs (8000); should clamp
    expect(computeBackoffMs({ attempt: 0, retryAfterSeconds: 30, policy })).toBe(8000)
  })

  it('ignores non-finite retryAfterSeconds and falls back to backoff', () => {
    expect(
      computeBackoffMs({ attempt: 0, retryAfterSeconds: Number.POSITIVE_INFINITY, policy }),
    ).toBe(100)
  })
})
