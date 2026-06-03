import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'

import { ThreeCommon } from '@/client'
import { ThreeCommonNotFoundError, ThreeCommonValidationError } from '@/errors'

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

const samplePrice = {
  id: 'price_123',
  hostId: 'host_1',
  productId: 'prod_7',
  type: 'recurring' as const,
  currency: 'USD' as const,
  unitAmount: 1500,
  recurring: { interval: 'month' as const, intervalCount: 1 },
  features: [
    { featureKey: 'api_calls', type: 'quantity' as const, quantity: 1000, rolloverEnabled: false },
  ],
  nickname: 'Pro monthly',
  active: true,
  metadata: {},
  createdAt: '2026-05-01T00:00:00.000Z',
  updatedAt: '2026-05-01T00:00:00.000Z',
}

describe('prices.list', () => {
  it('returns data + hasMore', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/prices`, () =>
        HttpResponse.json({ data: [samplePrice], hasMore: false }),
      ),
    )
    const client = buildClient()
    const result = await client.prices.list({ productId: 'prod_7', active: true })
    expect(result.hasMore).toBe(false)
    expect(result.data).toHaveLength(1)
    expect(result.data[0]?.id).toBe('price_123')
    expect(result.data[0]?.recurring?.interval).toBe('month')
  })

  it('forwards query params (including boolean active)', async () => {
    let url = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/prices`, ({ request }) => {
        url = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.prices.list({
      productId: 'prod_7',
      type: 'recurring',
      active: false,
      pageSize: 25,
    })
    expect(url).toContain('productId=prod_7')
    expect(url).toContain('type=recurring')
    expect(url).toContain('active=false')
    expect(url).toContain('pageSize=25')
  })
})

describe('prices.retrieve', () => {
  it('returns the unwrapped price', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/prices/price_123`, () =>
        HttpResponse.json({ data: samplePrice }),
      ),
    )
    const client = buildClient()
    const price = await client.prices.retrieve('price_123', { fields: 'id,unitAmount' })
    expect(price.id).toBe('price_123')
    expect(price.unitAmount).toBe(1500)
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.prices.retrieve('')).rejects.toThrow(TypeError)
  })

  it('throws ThreeCommonNotFoundError on 404', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/prices/price_missing`, () =>
        HttpResponse.json({ error: { code: 'not_found', message: 'gone' } }, { status: 404 }),
      ),
    )
    const client = buildClient()
    await expect(client.prices.retrieve('price_missing')).rejects.toBeInstanceOf(
      ThreeCommonNotFoundError,
    )
  })
})

describe('prices.create', () => {
  it('POSTs the body and returns the unwrapped price (201)', async () => {
    let body: unknown
    server.use(
      http.post(`${TEST_BASE_URL}/v1/prices`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: samplePrice }, { status: 201 })
      }),
    )
    const client = buildClient()
    const price = await client.prices.create({
      productId: 'prod_7',
      type: 'recurring',
      currency: 'USD',
      unitAmount: 1500,
      recurring: { interval: 'month', intervalCount: 1 },
    })
    expect(price.id).toBe('price_123')
    expect(body).toEqual({
      productId: 'prod_7',
      type: 'recurring',
      currency: 'USD',
      unitAmount: 1500,
      recurring: { interval: 'month', intervalCount: 1 },
    })
  })

  it('surfaces 400 validation errors', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/prices`, () =>
        HttpResponse.json(
          { error: { code: 'validation_error', message: 'recurring required' } },
          { status: 400 },
        ),
      ),
    )
    const client = buildClient()
    await expect(
      client.prices.create({
        productId: 'prod_7',
        type: 'recurring',
        currency: 'USD',
        unitAmount: 1500,
      }),
    ).rejects.toBeInstanceOf(ThreeCommonValidationError)
  })
})

describe('prices.update', () => {
  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.prices.update('', { unitAmount: 1 })).rejects.toThrow(TypeError)
  })

  it('PATCHes and returns the unwrapped price', async () => {
    let body: unknown
    server.use(
      http.patch(`${TEST_BASE_URL}/v1/prices/price_123`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: { ...samplePrice, unitAmount: 1200, nickname: null } })
      }),
    )
    const client = buildClient()
    const price = await client.prices.update('price_123', { unitAmount: 1200, nickname: null })
    expect(price.unitAmount).toBe(1200)
    expect(body).toEqual({ unitAmount: 1200, nickname: null })
  })
})

describe('prices.archive / unarchive', () => {
  it('archives and returns the price with active=false', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/prices/price_123/archive`, () =>
        HttpResponse.json({ data: { ...samplePrice, active: false } }),
      ),
    )
    const client = buildClient()
    const price = await client.prices.archive('price_123')
    expect(price.active).toBe(false)
  })

  it('unarchives and returns the price with active=true', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/prices/price_123/unarchive`, () =>
        HttpResponse.json({ data: { ...samplePrice, active: true } }),
      ),
    )
    const client = buildClient()
    const price = await client.prices.unarchive('price_123')
    expect(price.active).toBe(true)
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.prices.archive('')).rejects.toThrow(TypeError)
  })
})

describe('prices.list — paramsToQuery edge cases', () => {
  it('skips explicit undefined params', async () => {
    let captured = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/prices`, ({ request }) => {
        captured = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.prices.list({
      productId: 'prod_7',
      type: undefined,
      pageSize: undefined,
    } as unknown as Parameters<typeof client.prices.list>[0])
    expect(captured).toContain('productId=prod_7')
    expect(captured).not.toContain('type')
    expect(captured).not.toContain('pageSize')
  })

  it('drops non-primitive values silently rather than passing them on the wire', async () => {
    let captured = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/prices`, ({ request }) => {
        captured = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.prices.list({
      productId: 'prod_7',
      forbidden: ['x', 'y'],
    } as unknown as Parameters<typeof client.prices.list>[0])
    expect(captured).toContain('productId=prod_7')
    expect(captured).not.toContain('forbidden')
  })
})

describe('prices.listAutoPaginate', () => {
  it('iterates across pages', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/prices`, ({ request }) => {
        const url = new URL(request.url)
        const page = Number(url.searchParams.get('page') ?? '0')
        if (page === 0) {
          return HttpResponse.json({
            data: [
              { ...samplePrice, id: 'price_a' },
              { ...samplePrice, id: 'price_b' },
            ],
            hasMore: true,
          })
        }
        if (page === 1) {
          return HttpResponse.json({ data: [{ ...samplePrice, id: 'price_c' }], hasMore: false })
        }
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    const ids: string[] = []
    for await (const price of client.prices.listAutoPaginate({ active: true })) {
      if (price.id !== undefined) ids.push(price.id)
    }
    expect(ids).toEqual(['price_a', 'price_b', 'price_c'])
  })
})
