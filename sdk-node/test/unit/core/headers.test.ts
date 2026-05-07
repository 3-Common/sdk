import { describe, expect, it } from 'vitest'

import { buildHeaders } from '@/core/headers'

describe('buildHeaders', () => {
  const base = {
    apiKey: '3co_test',
    apiVersion: '2026-04-29',
    sdkVersion: '0.0.0-test',
    userAgentSuffix: 'Node/v22.0.0; darwin-arm64-24.0.0',
    telemetryHeader: undefined,
    idempotencyKey: undefined,
  }

  it('sets the standard request headers', () => {
    const headers = buildHeaders(base)
    expect(headers.get('Authorization')).toBe('Bearer 3co_test')
    expect(headers.get('Threecommon-Version')).toBe('2026-04-29')
    expect(headers.get('User-Agent')).toBe(
      'ThreeCommonNode/0.0.0-test (Node/v22.0.0; darwin-arm64-24.0.0)',
    )
    expect(headers.get('Accept')).toBe('application/json')
    expect(headers.get('Content-Type')).toBe('application/json')
  })

  it('omits the telemetry header when value is undefined', () => {
    const headers = buildHeaders(base)
    expect(headers.has('Threecommon-Client-Telemetry')).toBe(false)
  })

  it('attaches the telemetry header when supplied', () => {
    const headers = buildHeaders({ ...base, telemetryHeader: '{"lang":"node"}' })
    expect(headers.get('Threecommon-Client-Telemetry')).toBe('{"lang":"node"}')
  })

  it('omits Idempotency-Key when value is undefined', () => {
    const headers = buildHeaders(base)
    expect(headers.has('Idempotency-Key')).toBe(false)
  })

  it('attaches Idempotency-Key when supplied', () => {
    const headers = buildHeaders({ ...base, idempotencyKey: 'abc-123' })
    expect(headers.get('Idempotency-Key')).toBe('abc-123')
  })
})
