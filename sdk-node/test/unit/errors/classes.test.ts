import { describe, expect, it } from 'vitest'

import { errorFromResponse } from '@/errors'

describe('ThreeCommonError.toString', () => {
  it('formats as "[code] message (request_id=...)"', () => {
    const err = errorFromResponse({
      status: 404,
      body: { error: { code: 'not_found', message: 'Event evt_42 not found' } },
      rawResponse: undefined,
      requestId: 'req-abc-123',
      retryAfterSeconds: undefined,
    })
    expect(err.toString()).toBe('[not_found] Event evt_42 not found (request_id=req-abc-123)')
  })

  it('omits the request_id segment when not set', () => {
    const err = errorFromResponse({
      status: 500,
      body: undefined,
      rawResponse: undefined,
      requestId: undefined,
      retryAfterSeconds: undefined,
    })
    expect(err.toString()).toBe('[internal_error] Request failed with status 500')
  })
})

describe('error name', () => {
  it('matches the class name for catch-by-name patterns', () => {
    const err = errorFromResponse({
      status: 401,
      body: { error: { code: 'unauthorized', message: 'x' } },
      rawResponse: undefined,
      requestId: undefined,
      retryAfterSeconds: undefined,
    })
    expect(err.name).toBe('ThreeCommonAuthError')
  })
})
