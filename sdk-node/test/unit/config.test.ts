import { afterEach, beforeEach, describe, expect, it } from 'vitest'

import { resolveConfig } from '@/config'
import { ThreeCommonValidationError } from '@/errors'

describe('resolveConfig', () => {
  const original = process.env['THREECOMMON_API_KEY']

  beforeEach(() => {
    delete process.env['THREECOMMON_API_KEY']
  })
  afterEach(() => {
    if (original === undefined) delete process.env['THREECOMMON_API_KEY']
    else process.env['THREECOMMON_API_KEY'] = original
  })

  it('reads apiKey from THREECOMMON_API_KEY when not provided', () => {
    process.env['THREECOMMON_API_KEY'] = '3co_from_env'
    const cfg = resolveConfig({})
    expect(cfg.apiKey).toBe('3co_from_env')
  })

  it('prefers explicit apiKey over the env var', () => {
    process.env['THREECOMMON_API_KEY'] = '3co_from_env'
    const cfg = resolveConfig({ apiKey: '3co_explicit' })
    expect(cfg.apiKey).toBe('3co_explicit')
  })

  it('throws ThreeCommonValidationError when no apiKey is supplied', () => {
    expect(() => resolveConfig({})).toThrow(ThreeCommonValidationError)
  })

  it('applies defaults', () => {
    const cfg = resolveConfig({ apiKey: '3co_x' })
    expect(cfg.baseUrl).toBe('https://api.3common.com')
    expect(cfg.timeoutMs).toBe(30_000)
    expect(cfg.maxRetries).toBe(3)
    expect(cfg.retryDelay).toEqual({ initialMs: 500, maxMs: 8000, jitter: true })
    expect(cfg.telemetry).toBe(true)
  })

  it('strips trailing slashes from baseUrl', () => {
    const cfg = resolveConfig({ apiKey: '3co_x', baseUrl: 'https://api.3common.com///' })
    expect(cfg.baseUrl).toBe('https://api.3common.com')
  })

  it('rejects baseUrl without http(s) scheme', () => {
    expect(() => resolveConfig({ apiKey: '3co_x', baseUrl: 'api.3common.com' })).toThrow(
      ThreeCommonValidationError,
    )
  })

  it.each([0, -1, Number.NaN, Number.POSITIVE_INFINITY])(
    'rejects invalid timeoutMs (%s)',
    (timeoutMs) => {
      expect(() => resolveConfig({ apiKey: '3co_x', timeoutMs })).toThrow(
        ThreeCommonValidationError,
      )
    },
  )

  it.each([-1, 1.5, Number.NaN])('rejects invalid maxRetries (%s)', (maxRetries) => {
    expect(() => resolveConfig({ apiKey: '3co_x', maxRetries })).toThrow(ThreeCommonValidationError)
  })

  it('honors telemetry: false', () => {
    const cfg = resolveConfig({ apiKey: '3co_x', telemetry: false })
    expect(cfg.telemetry).toBe(false)
  })

  it('merges partial retryDelay with defaults', () => {
    const cfg = resolveConfig({ apiKey: '3co_x', retryDelay: { initialMs: 100 } })
    expect(cfg.retryDelay).toEqual({ initialMs: 100, maxMs: 8000, jitter: true })
  })

  it('passes through fetch and logger when provided', () => {
    const fakeFetch: typeof fetch = () => Promise.resolve(new Response())
    const logger = { debug: () => undefined }
    const cfg = resolveConfig({ apiKey: '3co_x', fetch: fakeFetch, logger })
    expect(cfg.fetch).toBe(fakeFetch)
    expect(cfg.logger).toBe(logger)
  })
})
