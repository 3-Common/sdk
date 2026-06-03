import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'

import { ThreeCommon } from '@/client'
import { ThreeCommonConflictError, ThreeCommonNotFoundError } from '@/errors'

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

const sampleFeature = {
  id: 'feat_123',
  hostId: 'host_1',
  key: 'api_calls',
  name: 'API calls',
  description: 'Monthly API call quota',
  type: 'quantity' as const,
  active: true,
  metadata: {},
  createdAt: '2026-05-01T00:00:00.000Z',
  updatedAt: '2026-05-01T00:00:00.000Z',
}

const sampleResolved = {
  feature: sampleFeature,
  value: { type: 'quantity' as const, quantity: 1000, balance: 850 },
  contributingSubscriptionIds: ['sub_1'],
}

describe('features.list', () => {
  it('returns data + hasMore', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/features`, () =>
        HttpResponse.json({ data: [sampleFeature], hasMore: false }),
      ),
    )
    const client = buildClient()
    const result = await client.features.list({ type: 'quantity', active: true })
    expect(result.hasMore).toBe(false)
    expect(result.data).toHaveLength(1)
    expect(result.data[0]?.key).toBe('api_calls')
  })

  it('forwards query params (including boolean active)', async () => {
    let url = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/features`, ({ request }) => {
        url = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.features.list({ type: 'enum', active: false, pageSize: 25 })
    expect(url).toContain('type=enum')
    expect(url).toContain('active=false')
    expect(url).toContain('pageSize=25')
  })
})

describe('features.resolve', () => {
  it('forwards contactId + featureKey and returns the unwrapped resolution', async () => {
    let url = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/features/resolve`, ({ request }) => {
        url = request.url
        return HttpResponse.json({ data: sampleResolved })
      }),
    )
    const client = buildClient()
    const resolved = await client.features.resolve({ contactId: 'cnt_7', featureKey: 'api_calls' })
    expect(url).toContain('contactId=cnt_7')
    expect(url).toContain('featureKey=api_calls')
    expect(resolved.feature.key).toBe('api_calls')
    expect(resolved.contributingSubscriptionIds).toEqual(['sub_1'])
    expect(resolved.value.type).toBe('quantity')
    if (resolved.value.type === 'quantity') {
      expect(resolved.value.quantity).toBe(1000)
      expect(resolved.value.balance).toBe(850)
    }
  })

  it('throws ThreeCommonNotFoundError for an unknown feature key', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/features/resolve`, () =>
        HttpResponse.json(
          { error: { code: 'not_found', message: 'unknown feature' } },
          { status: 404 },
        ),
      ),
    )
    const client = buildClient()
    await expect(
      client.features.resolve({ contactId: 'cnt_7', featureKey: 'nope' }),
    ).rejects.toBeInstanceOf(ThreeCommonNotFoundError)
  })
})

describe('features.retrieve', () => {
  it('returns the unwrapped feature', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/features/feat_123`, () =>
        HttpResponse.json({ data: sampleFeature }),
      ),
    )
    const client = buildClient()
    const feature = await client.features.retrieve('feat_123', { fields: 'id,key,type' })
    expect(feature.key).toBe('api_calls')
    expect(feature.type).toBe('quantity')
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.features.retrieve('')).rejects.toThrow(TypeError)
  })

  it('throws ThreeCommonNotFoundError on 404', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/features/feat_missing`, () =>
        HttpResponse.json({ error: { code: 'not_found', message: 'gone' } }, { status: 404 }),
      ),
    )
    const client = buildClient()
    await expect(client.features.retrieve('feat_missing')).rejects.toBeInstanceOf(
      ThreeCommonNotFoundError,
    )
  })
})

describe('features.create', () => {
  it('POSTs the body and returns the unwrapped feature (201)', async () => {
    let body: unknown
    server.use(
      http.post(`${TEST_BASE_URL}/v1/features`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: sampleFeature }, { status: 201 })
      }),
    )
    const client = buildClient()
    const feature = await client.features.create({
      key: 'api_calls',
      name: 'API calls',
      type: 'quantity',
    })
    expect(feature.id).toBe('feat_123')
    expect(body).toEqual({ key: 'api_calls', name: 'API calls', type: 'quantity' })
  })

  it('throws ThreeCommonConflictError on duplicate key (409)', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/features`, () =>
        HttpResponse.json(
          { error: { code: 'conflict', message: 'feature key exists' } },
          { status: 409 },
        ),
      ),
    )
    const client = buildClient()
    await expect(
      client.features.create({ key: 'api_calls', name: 'API calls', type: 'quantity' }),
    ).rejects.toBeInstanceOf(ThreeCommonConflictError)
  })
})

describe('features.update', () => {
  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.features.update('', { name: 'x' })).rejects.toThrow(TypeError)
  })

  it('PATCHes and returns the unwrapped feature (null clears)', async () => {
    let body: unknown
    server.use(
      http.patch(`${TEST_BASE_URL}/v1/features/feat_123`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({
          data: { ...sampleFeature, name: 'API requests', description: null },
        })
      }),
    )
    const client = buildClient()
    const feature = await client.features.update('feat_123', {
      name: 'API requests',
      description: null,
    })
    expect(feature.name).toBe('API requests')
    expect(body).toEqual({ name: 'API requests', description: null })
  })
})

describe('features.archive / unarchive', () => {
  it('archives and returns the feature with active=false', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/features/feat_123/archive`, () =>
        HttpResponse.json({ data: { ...sampleFeature, active: false } }),
      ),
    )
    const client = buildClient()
    const feature = await client.features.archive('feat_123')
    expect(feature.active).toBe(false)
  })

  it('unarchives and returns the feature with active=true', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/features/feat_123/unarchive`, () =>
        HttpResponse.json({ data: { ...sampleFeature, active: true } }),
      ),
    )
    const client = buildClient()
    const feature = await client.features.unarchive('feat_123')
    expect(feature.active).toBe(true)
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.features.archive('')).rejects.toThrow(TypeError)
  })
})

describe('features.list — paramsToQuery edge cases', () => {
  it('skips explicit undefined params', async () => {
    let captured = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/features`, ({ request }) => {
        captured = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.features.list({
      type: 'quantity',
      active: undefined,
      pageSize: undefined,
    } as unknown as Parameters<typeof client.features.list>[0])
    expect(captured).toContain('type=quantity')
    expect(captured).not.toContain('active')
    expect(captured).not.toContain('pageSize')
  })

  it('drops non-primitive values silently rather than passing them on the wire', async () => {
    let captured = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/features`, ({ request }) => {
        captured = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.features.list({
      type: 'quantity',
      forbidden: ['x', 'y'],
    } as unknown as Parameters<typeof client.features.list>[0])
    expect(captured).toContain('type=quantity')
    expect(captured).not.toContain('forbidden')
  })
})

describe('features.listAutoPaginate', () => {
  it('iterates across pages', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/features`, ({ request }) => {
        const url = new URL(request.url)
        const page = Number(url.searchParams.get('page') ?? '0')
        if (page === 0) {
          return HttpResponse.json({
            data: [
              { ...sampleFeature, id: 'feat_a' },
              { ...sampleFeature, id: 'feat_b' },
            ],
            hasMore: true,
          })
        }
        if (page === 1) {
          return HttpResponse.json({ data: [{ ...sampleFeature, id: 'feat_c' }], hasMore: false })
        }
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    const ids: string[] = []
    for await (const feature of client.features.listAutoPaginate({ active: true })) {
      if (feature.id !== undefined) ids.push(feature.id)
    }
    expect(ids).toEqual(['feat_a', 'feat_b', 'feat_c'])
  })
})
