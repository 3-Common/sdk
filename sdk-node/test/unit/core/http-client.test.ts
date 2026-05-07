import { http, HttpResponse } from 'msw'
import { describe, expect, it, vi } from 'vitest'

import { API_VERSION } from '@/api-version'
import { HttpClient } from '@/core/http-client'
import { Telemetry } from '@/core/telemetry'
import {
  ThreeCommonConnectionError,
  ThreeCommonNotFoundError,
  ThreeCommonRateLimitError,
  ThreeCommonServerError,
} from '@/errors'

import { setupMockServer, TEST_BASE_URL } from '../../helpers/mock-server'

import type { Logger } from '@/types/public'

const server = setupMockServer()

function buildClient(
  overrides: {
    maxRetries?: number
    jitter?: boolean
    telemetry?: boolean
    logger?: Logger
    fetch?: typeof fetch
  } = {},
): HttpClient {
  return new HttpClient({
    apiKey: '3co_test',
    baseUrl: TEST_BASE_URL,
    apiVersion: API_VERSION,
    timeoutMs: 5_000,
    retry: {
      maxRetries: overrides.maxRetries ?? 3,
      initialDelayMs: 1,
      maxDelayMs: 4,
      jitter: overrides.jitter ?? false,
    },
    fetch: overrides.fetch ?? globalThis.fetch,
    telemetry: new Telemetry(overrides.telemetry ?? true),
    logger: overrides.logger,
  })
}

describe('HttpClient.request', () => {
  it('builds Authorization, Threecommon-Version, User-Agent, and telemetry headers', async () => {
    let capturedAuth = ''
    let capturedVersion = ''
    let capturedUA = ''
    let capturedTelemetry: string | null = null

    server.use(
      http.get(`${TEST_BASE_URL}/v1/events`, ({ request }) => {
        capturedAuth = request.headers.get('Authorization') ?? ''
        capturedVersion = request.headers.get('Threecommon-Version') ?? ''
        capturedUA = request.headers.get('User-Agent') ?? ''
        capturedTelemetry = request.headers.get('Threecommon-Client-Telemetry')
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )

    const client = buildClient()
    await client.request({ method: 'GET', path: '/events' })

    expect(capturedAuth).toBe('Bearer 3co_test')
    expect(capturedVersion).toBe(API_VERSION)
    expect(capturedUA).toMatch(/^ThreeCommonNode\/.+ \(Node\/v.+\)$/u)
    expect(capturedTelemetry).toBeTypeOf('string')
  })

  it('omits telemetry header when disabled', async () => {
    let header: string | null = null
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events`, ({ request }) => {
        header = request.headers.get('Threecommon-Client-Telemetry')
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )

    const client = buildClient({ telemetry: false })
    await client.request({ method: 'GET', path: '/events' })
    expect(header).toBeNull()
  })

  it('serializes query params and skips undefined values', async () => {
    let capturedUrl = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events`, ({ request }) => {
        capturedUrl = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.request({
      method: 'GET',
      path: '/events',
      query: { status: 'open', pageSize: 50, missing: undefined },
    })
    expect(capturedUrl).toContain('status=open')
    expect(capturedUrl).toContain('pageSize=50')
    expect(capturedUrl).not.toContain('missing')
  })

  it('throws a typed error on 404', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events/evt_missing`, () =>
        HttpResponse.json(
          { error: { code: 'not_found', message: 'Event not found' } },
          { status: 404, headers: { 'X-Request-ID': 'req-test' } },
        ),
      ),
    )

    const client = buildClient({ maxRetries: 0 })
    await expect(
      client.request({ method: 'GET', path: '/events/evt_missing' }),
    ).rejects.toBeInstanceOf(ThreeCommonNotFoundError)
  })

  it('retries idempotent requests on 5xx and eventually surfaces the error', async () => {
    let count = 0
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events`, () => {
        count += 1
        return HttpResponse.json(
          { error: { code: 'internal_error', message: 'boom' } },
          { status: 503 },
        )
      }),
    )
    const client = buildClient({ maxRetries: 2 })
    await expect(client.request({ method: 'GET', path: '/events' })).rejects.toBeInstanceOf(
      ThreeCommonServerError,
    )
    expect(count).toBe(3) // 1 initial + 2 retries
  })

  it('honors Retry-After on 429 and surfaces a typed rate-limit error', async () => {
    let count = 0
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events`, () => {
        count += 1
        return HttpResponse.json(
          { error: { code: 'rate_limit_exceeded', message: 'slow down' } },
          { status: 429, headers: { 'Retry-After': '0' } },
        )
      }),
    )
    const client = buildClient({ maxRetries: 1 })
    await expect(client.request({ method: 'GET', path: '/events' })).rejects.toBeInstanceOf(
      ThreeCommonRateLimitError,
    )
    expect(count).toBe(2)
  })

  it('does not retry non-idempotent methods (POST) on 5xx', async () => {
    let count = 0
    server.use(
      http.post(`${TEST_BASE_URL}/v1/events`, () => {
        count += 1
        return HttpResponse.json(
          { error: { code: 'internal_error', message: 'boom' } },
          { status: 503 },
        )
      }),
    )
    const client = buildClient({ maxRetries: 3 })
    await expect(
      client.request({ method: 'POST', path: '/events', body: { name: 'x' } }),
    ).rejects.toBeInstanceOf(ThreeCommonServerError)
    expect(count).toBe(1)
  })

  it('attaches X-Request-ID from the response onto the thrown error', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events/evt_x`, () =>
        HttpResponse.json(
          { error: { code: 'not_found', message: 'gone' } },
          { status: 404, headers: { 'X-Request-ID': 'req-dfx-001' } },
        ),
      ),
    )
    const client = buildClient({ maxRetries: 0 })
    try {
      await client.request({ method: 'GET', path: '/events/evt_x' })
      expect.fail('should throw')
    } catch (err) {
      expect(err).toBeInstanceOf(ThreeCommonNotFoundError)
      expect((err as ThreeCommonNotFoundError).requestId).toBe('req-dfx-001')
    }
  })

  it('forwards Idempotency-Key header when supplied via options', async () => {
    let captured = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events`, ({ request }) => {
        captured = request.headers.get('Idempotency-Key') ?? ''
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.request({ method: 'GET', path: '/events', options: { idempotencyKey: 'abc-123' } })
    expect(captured).toBe('abc-123')
  })

  it('aborts the in-flight request when the caller-provided signal aborts', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events`, async () => {
        await new Promise((resolve) => setTimeout(resolve, 5_000))
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient({ maxRetries: 0 })
    const controller = new AbortController()
    setTimeout(() => {
      controller.abort()
    }, 10)
    await expect(
      client.request({ method: 'GET', path: '/events', options: { signal: controller.signal } }),
    ).rejects.toThrow()
  })

  it('rejects with a connection error when fetch throws', async () => {
    server.use(http.get(`${TEST_BASE_URL}/v1/events`, () => HttpResponse.error()))
    const client = buildClient({ maxRetries: 0 })
    await expect(client.request({ method: 'GET', path: '/events' })).rejects.toMatchObject({
      code: 'connection_error',
    })
  })

  it('retries network errors on idempotent methods up to maxRetries', async () => {
    let attempts = 0
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events`, () => {
        attempts += 1
        if (attempts < 3) return HttpResponse.error()
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient({ maxRetries: 3 })
    const result = await client.request<{ data: unknown[]; hasMore: boolean }>({
      method: 'GET',
      path: '/events',
    })
    expect(result.hasMore).toBe(false)
    expect(attempts).toBe(3)
  })

  it('returns undefined for 2xx responses with non-JSON bodies', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events`, () => new HttpResponse('not-json', { status: 200 })),
    )
    const client = buildClient()
    const result = await client.request<unknown>({ method: 'GET', path: '/events' })
    expect(result).toBeUndefined()
  })

  it('parses 4xx responses with non-JSON bodies into a typed error using defaults', async () => {
    server.use(
      http.get(
        `${TEST_BASE_URL}/v1/events`,
        () => new HttpResponse('<html>nope</html>', { status: 404 }),
      ),
    )
    const client = buildClient({ maxRetries: 0 })
    await expect(client.request({ method: 'GET', path: '/events' })).rejects.toMatchObject({
      code: 'not_found',
      httpStatus: 404,
    })
  })

  it('returns the response body for 200 with empty body', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events`, () => new HttpResponse(null, { status: 200 })),
    )
    const client = buildClient()
    const result = await client.request<unknown>({ method: 'GET', path: '/events' })
    expect(result).toBeUndefined()
  })

  it('honors a Retry-After header containing an HTTP-date', async () => {
    let attempts = 0
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events`, () => {
        attempts += 1
        if (attempts === 1) {
          return HttpResponse.json(
            { error: { code: 'rate_limit_exceeded', message: 'wait' } },
            { status: 429, headers: { 'Retry-After': new Date(Date.now() - 1000).toUTCString() } },
          )
        }
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient({ maxRetries: 1 })
    await client.request({ method: 'GET', path: '/events' })
    expect(attempts).toBe(2)
  })

  it('invokes logger.debug for every request when a logger is supplied', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events`, () =>
        HttpResponse.json({ data: [], hasMore: false }, { headers: { 'X-Request-ID': 'req-log' } }),
      ),
    )
    const debug = vi.fn()
    const client = buildClient({ logger: { debug } })
    await client.request({ method: 'GET', path: '/events' })
    expect(debug).toHaveBeenCalledWith(
      'threecommon:request',
      expect.objectContaining({
        method: 'GET',
        path: '/events',
        status: 200,
        requestId: 'req-log',
        attempt: 0,
      }),
    )
  })

  it('wraps non-Error fetch throws in ThreeCommonConnectionError', async () => {
    // eslint-disable-next-line @typescript-eslint/prefer-promise-reject-errors -- intentional: simulate buggy upstream that rejects with a string
    const evilFetch: typeof fetch = () => Promise.reject('totally not an Error')
    const client = buildClient({ maxRetries: 0, fetch: evilFetch })
    await expect(client.request({ method: 'GET', path: '/events' })).rejects.toBeInstanceOf(
      ThreeCommonConnectionError,
    )
  })

  it('wraps non-string non-Error fetch throws with the default message', async () => {
    // eslint-disable-next-line @typescript-eslint/prefer-promise-reject-errors -- intentional: simulate buggy upstream that rejects with a number
    const evilFetch: typeof fetch = () => Promise.reject(42)
    const client = buildClient({ maxRetries: 0, fetch: evilFetch })
    await expect(client.request({ method: 'GET', path: '/events' })).rejects.toMatchObject({
      code: 'connection_error',
      message: 'Request failed before reaching the server',
    })
  })
})
