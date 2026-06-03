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

const sampleEntitlement = {
  id: 'ent_123',
  hostId: 'host_1',
  contactId: 'cnt_7',
  featureKey: 'api_calls',
  balance: 100,
  grants: [
    {
      id: 'grant_1',
      source: 'manual' as const,
      amount: 100,
      remaining: 100,
      addedAt: '2026-05-01T18:00:00.000Z',
    },
  ],
  totalGranted: 100,
  totalConsumed: 0,
  metadata: {},
  createdAt: '2026-05-01T18:00:00.000Z',
  updatedAt: '2026-05-01T18:00:00.000Z',
}

describe('entitlements.list', () => {
  it('returns data + hasMore', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/entitlements`, () =>
        HttpResponse.json({ data: [sampleEntitlement], hasMore: false }),
      ),
    )
    const client = buildClient()
    const result = await client.entitlements.list({ featureKey: 'api_calls' })
    expect(result.hasMore).toBe(false)
    expect(result.data).toHaveLength(1)
    expect(result.data[0]?.id).toBe('ent_123')
  })

  it('forwards query params', async () => {
    let url = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/entitlements`, ({ request }) => {
        url = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.entitlements.list({
      contactId: 'cnt_7',
      featureKey: 'api_calls',
      minBalance: 1,
      pageSize: 25,
    })
    expect(url).toContain('contactId=cnt_7')
    expect(url).toContain('featureKey=api_calls')
    expect(url).toContain('minBalance=1')
    expect(url).toContain('pageSize=25')
  })
})

describe('entitlements.retrieve', () => {
  it('returns the unwrapped entitlement', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/entitlements/ent_123`, () =>
        HttpResponse.json({ data: sampleEntitlement }),
      ),
    )
    const client = buildClient()
    const entitlement = await client.entitlements.retrieve('ent_123', { fields: 'id,balance' })
    expect(entitlement.id).toBe('ent_123')
    expect(entitlement.balance).toBe(100)
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.entitlements.retrieve('')).rejects.toThrow(TypeError)
  })

  it('throws ThreeCommonNotFoundError on 404', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/entitlements/ent_missing`, () =>
        HttpResponse.json({ error: { code: 'not_found', message: 'gone' } }, { status: 404 }),
      ),
    )
    const client = buildClient()
    await expect(client.entitlements.retrieve('ent_missing')).rejects.toBeInstanceOf(
      ThreeCommonNotFoundError,
    )
  })
})

describe('entitlements.lookup', () => {
  it('forwards contactId + featureKey and returns the unwrapped entitlement', async () => {
    let url = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/entitlements/lookup`, ({ request }) => {
        url = request.url
        return HttpResponse.json({ data: sampleEntitlement })
      }),
    )
    const client = buildClient()
    const entitlement = await client.entitlements.lookup({
      contactId: 'cnt_7',
      featureKey: 'api_calls',
    })
    expect(url).toContain('contactId=cnt_7')
    expect(url).toContain('featureKey=api_calls')
    expect(entitlement.id).toBe('ent_123')
  })

  it('throws ThreeCommonNotFoundError when no record exists', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/entitlements/lookup`, () =>
        HttpResponse.json({ error: { code: 'not_found', message: 'gone' } }, { status: 404 }),
      ),
    )
    const client = buildClient()
    await expect(
      client.entitlements.lookup({ contactId: 'cnt_7', featureKey: 'unknown' }),
    ).rejects.toBeInstanceOf(ThreeCommonNotFoundError)
  })
})

describe('entitlements.grant', () => {
  it('POSTs the body and returns the unwrapped entitlement', async () => {
    let body: unknown
    server.use(
      http.post(`${TEST_BASE_URL}/v1/entitlements/grants`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: { ...sampleEntitlement, balance: 150 } })
      }),
    )
    const client = buildClient()
    const entitlement = await client.entitlements.grant({
      contactId: 'cnt_7',
      featureKey: 'api_calls',
      amount: 50,
      grantId: 'grant_2',
    })
    expect(entitlement.balance).toBe(150)
    expect(body).toEqual({
      contactId: 'cnt_7',
      featureKey: 'api_calls',
      amount: 50,
      grantId: 'grant_2',
    })
  })
})

describe('entitlements.consume', () => {
  it('POSTs the body and returns the unwrapped entitlement', async () => {
    let body: unknown
    server.use(
      http.post(`${TEST_BASE_URL}/v1/entitlements/consume`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: { ...sampleEntitlement, balance: 99 } })
      }),
    )
    const client = buildClient()
    const entitlement = await client.entitlements.consume({
      contactId: 'cnt_7',
      featureKey: 'api_calls',
      amount: 1,
      reason: 'POST /generate',
    })
    expect(entitlement.balance).toBe(99)
    expect(body).toEqual({
      contactId: 'cnt_7',
      featureKey: 'api_calls',
      amount: 1,
      reason: 'POST /generate',
    })
  })

  it('throws ThreeCommonConflictError on insufficient balance (409)', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/entitlements/consume`, () =>
        HttpResponse.json(
          { error: { code: 'conflict', message: 'insufficient balance' } },
          { status: 409 },
        ),
      ),
    )
    const client = buildClient()
    await expect(
      client.entitlements.consume({ contactId: 'cnt_7', featureKey: 'api_calls', amount: 9999 }),
    ).rejects.toBeInstanceOf(ThreeCommonConflictError)
  })
})

describe('entitlements.list — paramsToQuery edge cases', () => {
  it('skips explicit undefined params', async () => {
    let captured = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/entitlements`, ({ request }) => {
        captured = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    // Bypass exactOptionalPropertyTypes to feed explicit undefineds — verifies paramsToQuery
    // correctly skips them.
    await client.entitlements.list({
      featureKey: 'api_calls',
      contactId: undefined,
      minBalance: undefined,
    } as unknown as Parameters<typeof client.entitlements.list>[0])
    expect(captured).toContain('featureKey=api_calls')
    expect(captured).not.toContain('contactId')
    expect(captured).not.toContain('minBalance')
  })

  it('drops non-primitive values silently rather than passing them on the wire', async () => {
    let captured = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/entitlements`, ({ request }) => {
        captured = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    // Bypass the type system to pass an array; the SDK must drop it.
    await client.entitlements.list({
      featureKey: 'api_calls',
      forbidden: ['x', 'y'],
    } as unknown as Parameters<typeof client.entitlements.list>[0])
    expect(captured).toContain('featureKey=api_calls')
    expect(captured).not.toContain('forbidden')
  })
})

describe('entitlements.listAutoPaginate', () => {
  it('iterates across pages', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/entitlements`, ({ request }) => {
        const url = new URL(request.url)
        const page = Number(url.searchParams.get('page') ?? '0')
        if (page === 0) {
          return HttpResponse.json({
            data: [
              { ...sampleEntitlement, id: 'ent_a' },
              { ...sampleEntitlement, id: 'ent_b' },
            ],
            hasMore: true,
          })
        }
        if (page === 1) {
          return HttpResponse.json({
            data: [{ ...sampleEntitlement, id: 'ent_c' }],
            hasMore: false,
          })
        }
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    const ids: string[] = []
    for await (const entitlement of client.entitlements.listAutoPaginate({
      featureKey: 'api_calls',
    })) {
      if (entitlement.id !== undefined) ids.push(entitlement.id)
    }
    expect(ids).toEqual(['ent_a', 'ent_b', 'ent_c'])
  })
})
