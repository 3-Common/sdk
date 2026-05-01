import { describe, expect, it } from 'vitest'

import { API_VERSION } from '@/api-version'
import { Telemetry } from '@/core/telemetry'

describe('Telemetry', () => {
  it('returns undefined header when nothing has been recorded yet', () => {
    const telemetry = new Telemetry(true)
    const value = telemetry.buildHeaderValue()
    expect(value).toBeDefined()
    expect(JSON.parse(value!)).toEqual({
      lang: 'node',
      sdk: expect.any(String),
      api: API_VERSION,
      last: undefined,
    })
  })

  it('records and surfaces the last request', () => {
    const telemetry = new Telemetry(true)
    telemetry.record({
      method: 'GET',
      path: '/events',
      status: 200,
      durationMs: 42.7,
      requestId: 'req-x',
    })
    const value = telemetry.buildHeaderValue() ?? ''
    expect(JSON.parse(value)).toMatchObject({
      lang: 'node',
      api: API_VERSION,
      last: { m: 'GET', p: '/events', s: 200, d: 42.7 },
    })
  })

  it('returns undefined header when disabled', () => {
    const telemetry = new Telemetry(false)
    expect(telemetry.buildHeaderValue()).toBeUndefined()
  })

  it('disable() drops any recorded metric and disables the header', () => {
    const telemetry = new Telemetry(true)
    telemetry.record({ method: 'GET', path: '/events', status: 200, durationMs: 1, requestId: 'r' })
    telemetry.disable()
    expect(telemetry.isEnabled()).toBe(false)
    expect(telemetry.buildHeaderValue()).toBeUndefined()
  })

  it('record() is a no-op when disabled', () => {
    const telemetry = new Telemetry(false)
    telemetry.record({ method: 'GET', path: '/events', status: 200, durationMs: 1, requestId: 'r' })
    expect(telemetry.buildHeaderValue()).toBeUndefined()
  })
})
