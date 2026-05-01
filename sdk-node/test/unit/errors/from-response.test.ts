import { describe, expect, it } from 'vitest'

import {
  errorFromResponse,
  ThreeCommonAuthError,
  ThreeCommonConflictError,
  ThreeCommonError,
  ThreeCommonNotFoundError,
  ThreeCommonPermissionError,
  ThreeCommonRateLimitError,
  ThreeCommonServerError,
  ThreeCommonValidationError,
} from '@/errors'

describe('errorFromResponse', () => {
  const baseArgs = {
    rawResponse: undefined,
    requestId: 'req-test-001',
    retryAfterSeconds: undefined,
  }

  it('maps 401 to ThreeCommonAuthError', () => {
    const err = errorFromResponse({
      ...baseArgs,
      status: 401,
      body: { error: { code: 'unauthorized', message: 'Invalid API key' } },
    })
    expect(err).toBeInstanceOf(ThreeCommonAuthError)
    expect(err).toBeInstanceOf(ThreeCommonError)
    expect(err.code).toBe('unauthorized')
    expect(err.httpStatus).toBe(401)
    expect(err.requestId).toBe('req-test-001')
  })

  it('maps 403 to ThreeCommonPermissionError', () => {
    const err = errorFromResponse({
      ...baseArgs,
      status: 403,
      body: { error: { code: 'forbidden', message: 'Missing scope' } },
    })
    expect(err).toBeInstanceOf(ThreeCommonPermissionError)
  })

  it('maps 404 to ThreeCommonNotFoundError', () => {
    const err = errorFromResponse({
      ...baseArgs,
      status: 404,
      body: { error: { code: 'not_found', message: 'Event not found' } },
    })
    expect(err).toBeInstanceOf(ThreeCommonNotFoundError)
  })

  it('maps 400 and 422 to ThreeCommonValidationError', () => {
    const e400 = errorFromResponse({
      ...baseArgs,
      status: 400,
      body: { error: { code: 'bad_request', message: 'x' } },
    })
    const e422 = errorFromResponse({
      ...baseArgs,
      status: 422,
      body: { error: { code: 'validation_error', message: 'y' } },
    })
    expect(e400).toBeInstanceOf(ThreeCommonValidationError)
    expect(e422).toBeInstanceOf(ThreeCommonValidationError)
  })

  it('maps 409 to ThreeCommonConflictError', () => {
    const err = errorFromResponse({
      ...baseArgs,
      status: 409,
      body: { error: { code: 'conflict', message: 'x' } },
    })
    expect(err).toBeInstanceOf(ThreeCommonConflictError)
  })

  it('maps 429 to ThreeCommonRateLimitError and carries retryAfterSeconds', () => {
    const err = errorFromResponse({
      ...baseArgs,
      status: 429,
      body: { error: { code: 'rate_limit_exceeded', message: 'slow down' } },
      retryAfterSeconds: 30,
    })
    expect(err).toBeInstanceOf(ThreeCommonRateLimitError)
    expect((err as ThreeCommonRateLimitError).retryAfterSeconds).toBe(30)
  })

  it('maps 5xx to ThreeCommonServerError', () => {
    const err = errorFromResponse({ ...baseArgs, status: 503, body: undefined })
    expect(err).toBeInstanceOf(ThreeCommonServerError)
    expect(err.httpStatus).toBe(503)
  })

  it('falls back to default code/message when body is malformed', () => {
    const err = errorFromResponse({ ...baseArgs, status: 401, body: 'plain text' })
    expect(err).toBeInstanceOf(ThreeCommonAuthError)
    expect(err.code).toBe('unauthorized')
    expect(err.message).toBe('Request failed with status 401')
  })

  it('preserves error.details when present', () => {
    const err = errorFromResponse({
      ...baseArgs,
      status: 403,
      body: {
        error: { code: 'forbidden', message: 'x', details: { requiredScopes: ['events:write'] } },
      },
    })
    expect(err.details).toEqual({ requiredScopes: ['events:write'] })
  })
})

describe('errorFromResponse — edge cases', () => {
  const baseArgs = {
    rawResponse: undefined,
    requestId: undefined,
    retryAfterSeconds: undefined,
  }

  it('uses generic default code for unmapped 4xx (e.g. 418)', () => {
    const err = errorFromResponse({ ...baseArgs, status: 418, body: undefined })
    expect(err.code).toBe('request_failed')
    expect(err).toBeInstanceOf(ThreeCommonValidationError)
  })

  it('treats body where error.code is not a string as malformed', () => {
    const err = errorFromResponse({
      ...baseArgs,
      status: 401,
      body: { error: { code: 12345, message: 'x' } },
    })
    expect(err.code).toBe('unauthorized')
    expect(err.message).toBe('Request failed with status 401')
  })

  it('treats body where error is not an object as malformed', () => {
    const err = errorFromResponse({
      ...baseArgs,
      status: 401,
      body: { error: 'string instead of object' },
    })
    expect(err.code).toBe('unauthorized')
  })

  it.each([
    [401, 'unauthorized'],
    [403, 'forbidden'],
    [404, 'not_found'],
    [409, 'conflict'],
    [429, 'rate_limit_exceeded'],
    [503, 'internal_error'],
  ])('uses default code "%s" for status %d when body has no error', (status, expectedCode) => {
    const err = errorFromResponse({ ...baseArgs, status, body: undefined })
    expect(err.code).toBe(expectedCode)
  })
})
