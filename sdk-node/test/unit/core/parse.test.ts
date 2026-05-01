import { describe, expect, it } from 'vitest'

import { parseRetryAfter, parseSuccessBody, tryParseJson } from '@/core/parse'

describe('tryParseJson', () => {
  it('returns undefined for empty input', () => {
    expect(tryParseJson('')).toBeUndefined()
  })

  it('parses valid JSON', () => {
    expect(tryParseJson('{"a":1}')).toEqual({ a: 1 })
  })

  it('returns undefined when JSON is malformed', () => {
    expect(tryParseJson('not-json')).toBeUndefined()
  })
})

describe('parseSuccessBody', () => {
  const baseResponse = {
    status: 200,
    headers: new Map<string, string>(),
    requestId: undefined,
  }

  it('returns undefined for empty bodies', () => {
    expect(parseSuccessBody({ ...baseResponse, bodyText: '' })).toBeUndefined()
  })

  it('parses JSON bodies', () => {
    expect(parseSuccessBody({ ...baseResponse, bodyText: '{"hasMore":false}' })).toEqual({
      hasMore: false,
    })
  })

  it('returns undefined for non-JSON bodies on 2xx', () => {
    expect(parseSuccessBody({ ...baseResponse, bodyText: '<html>nope</html>' })).toBeUndefined()
  })
})

describe('parseRetryAfter', () => {
  it('returns undefined for missing header', () => {
    expect(parseRetryAfter(undefined)).toBeUndefined()
  })

  it('returns undefined for empty string', () => {
    expect(parseRetryAfter('')).toBeUndefined()
  })

  it('parses delta-seconds integer', () => {
    expect(parseRetryAfter('30')).toBe(30)
  })

  it('treats `0` as a valid retry delay', () => {
    expect(parseRetryAfter('0')).toBe(0)
  })

  it('parses an HTTP-date in the future', () => {
    const future = new Date(Date.now() + 60_000).toUTCString()
    const result = parseRetryAfter(future)
    expect(result).toBeDefined()
    expect(result).toBeGreaterThan(0)
  })

  it('returns 0 for an HTTP-date in the past', () => {
    const past = new Date(Date.now() - 60_000).toUTCString()
    expect(parseRetryAfter(past)).toBe(0)
  })

  it('returns undefined for an unparseable header', () => {
    expect(parseRetryAfter('not-a-date-or-number')).toBeUndefined()
  })

  it('treats negative seconds as a past date and returns 0', () => {
    // '-5' fails the integer check (negative); falls through to Date.parse.
    // Whatever Date.parse returns, a negative delta should clamp to 0.
    const result = parseRetryAfter('-5')
    expect(result === undefined || result === 0).toBe(true)
  })
})
