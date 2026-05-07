import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'

import { ThreeCommon } from '@/client'
import { ThreeCommonNotFoundError } from '@/errors'

import { setupMockServer, TEST_BASE_URL } from '../../helpers/mock-server'

const server = setupMockServer()

function buildClient(): ThreeCommon {
  return new ThreeCommon({
    apiKey: '3co_test',
    baseUrl: TEST_BASE_URL,
    maxRetries: 0,
    telemetry: false,
  })
}

const sampleEvent = {
  id: 'evt_123',
  name: 'Test event',
  type: 'event',
  schedule: 'Single date' as const,
  start: '2026-05-01T18:00:00.000Z',
  status: 'open' as const,
  itemsSold: 0,
  revenueCents: 0,
  minPriceCents: null,
  maxPriceCents: null,
  currency: 'USD',
  isPublic: true,
  isVirtual: false,
}

describe('events.list', () => {
  it('returns data + hasMore', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events`, () =>
        HttpResponse.json({ data: [sampleEvent], hasMore: false }),
      ),
    )
    const client = buildClient()
    const result = await client.events.list({ status: 'open' })
    expect(result.hasMore).toBe(false)
    expect(result.data).toHaveLength(1)
    expect(result.data[0]?.id).toBe('evt_123')
  })

  it('forwards query params', async () => {
    let url = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events`, ({ request }) => {
        url = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.events.list({ status: 'open', pageSize: 25, search: 'concert' })
    expect(url).toContain('status=open')
    expect(url).toContain('pageSize=25')
    expect(url).toContain('search=concert')
  })
})

describe('events.retrieve', () => {
  it('returns the unwrapped event', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events/evt_123`, () =>
        HttpResponse.json({ data: sampleEvent }),
      ),
    )
    const client = buildClient()
    const event = await client.events.retrieve('evt_123', { fields: 'id,name' })
    expect(event.id).toBe('evt_123')
    expect(event.name).toBe('Test event')
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.events.retrieve('')).rejects.toThrow(TypeError)
  })

  it('throws ThreeCommonNotFoundError on 404', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events/evt_missing`, () =>
        HttpResponse.json({ error: { code: 'not_found', message: 'gone' } }, { status: 404 }),
      ),
    )
    const client = buildClient()
    await expect(client.events.retrieve('evt_missing')).rejects.toBeInstanceOf(
      ThreeCommonNotFoundError,
    )
  })
})

describe('events.update', () => {
  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.events.update('', { name: 'x' })).rejects.toThrow(TypeError)
  })

  it('PATCHes and returns the unwrapped event', async () => {
    let body: unknown
    server.use(
      http.patch(`${TEST_BASE_URL}/v1/events/evt_123`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: { ...sampleEvent, name: 'Renamed' } })
      }),
    )
    const client = buildClient()
    const updated = await client.events.update('evt_123', { name: 'Renamed' })
    expect(updated.name).toBe('Renamed')
    expect(body).toEqual({ name: 'Renamed' })
  })
})

describe('events.list — paramsToQuery edge cases', () => {
  it('skips explicit undefined params', async () => {
    let captured = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events`, ({ request }) => {
        captured = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    // Bypass exactOptionalPropertyTypes to feed explicit undefineds — verifies paramsToQuery
    // correctly skips them.
    await client.events.list({
      status: 'open',
      pageSize: undefined,
      search: undefined,
    } as unknown as Parameters<typeof client.events.list>[0])
    expect(captured).toContain('status=open')
    expect(captured).not.toContain('pageSize')
    expect(captured).not.toContain('search')
  })

  it('drops non-primitive values silently rather than passing them on the wire', async () => {
    let captured = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events`, ({ request }) => {
        captured = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    // Bypass the type system to pass an array; the SDK must drop it.
    await client.events.list({
      status: 'open',
      forbidden: ['x', 'y'],
    } as unknown as Parameters<typeof client.events.list>[0])
    expect(captured).toContain('status=open')
    expect(captured).not.toContain('forbidden')
  })
})

describe('events.listAutoPaginate', () => {
  it('iterates across pages', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/events`, ({ request }) => {
        const url = new URL(request.url)
        const page = Number(url.searchParams.get('page') ?? '0')
        if (page === 0) {
          return HttpResponse.json({
            data: [
              { ...sampleEvent, id: 'evt_a' },
              { ...sampleEvent, id: 'evt_b' },
            ],
            hasMore: true,
          })
        }
        if (page === 1) {
          return HttpResponse.json({ data: [{ ...sampleEvent, id: 'evt_c' }], hasMore: false })
        }
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    const ids: string[] = []
    for await (const event of client.events.listAutoPaginate({ status: 'open' })) {
      if (event.id !== undefined) ids.push(event.id)
    }
    expect(ids).toEqual(['evt_a', 'evt_b', 'evt_c'])
  })
})
