import { describe, expect, it } from 'vitest'

import { nowMs, resolveFetch, userAgentSuffix } from '@/core/platform'

describe('resolveFetch', () => {
  it('returns globalThis.fetch when no override is provided', () => {
    expect(resolveFetch(undefined)).toBe(globalThis.fetch)
  })

  it('accepts a fetch override', () => {
    const fakeFetch: typeof fetch = () => Promise.resolve(new Response())
    expect(resolveFetch(fakeFetch)).toBe(fakeFetch)
  })

  it('throws TypeError when the resolved fetch is not a function', () => {
    expect(() => resolveFetch(42 as unknown as typeof fetch)).toThrow(TypeError)
  })
})

describe('nowMs', () => {
  it('returns a positive monotonic timestamp', () => {
    const a = nowMs()
    const b = nowMs()
    expect(a).toBeGreaterThan(0)
    expect(b).toBeGreaterThanOrEqual(a)
  })
})

describe('userAgentSuffix', () => {
  it('reports the Node version and platform', () => {
    expect(userAgentSuffix()).toMatch(/^Node\/v.+; .+-.+-.+$/u)
  })
})
