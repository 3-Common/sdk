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

const sampleSubscription = {
  id: 'sub_123',
  hostId: 'hst_1',
  contactId: 'cnt_42',
  priceId: 'price_7',
  quantity: 1,
  status: 'active' as const,
  currentPeriodStart: '2026-05-01T00:00:00.000Z',
  currentPeriodEnd: '2026-06-01T00:00:00.000Z',
  cancelAtPeriodEnd: false,
  dunningEnabled: true,
  autoCharge: true,
  createdAt: '2026-04-01T00:00:00.000Z',
  updatedAt: '2026-05-01T00:00:00.000Z',
}

const sampleInvoiceRef = {
  id: 'inv_500',
  status: 'open',
  total: 50_000,
  currency: 'USD',
}

describe('subscriptions.list', () => {
  it('returns data + hasMore', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/subscriptions`, () =>
        HttpResponse.json({ data: [sampleSubscription], hasMore: false }),
      ),
    )
    const client = buildClient()
    const result = await client.subscriptions.list({ status: 'active' })
    expect(result.hasMore).toBe(false)
    expect(result.data).toHaveLength(1)
    expect(result.data[0]?.id).toBe('sub_123')
  })

  it('forwards filters as query params', async () => {
    let url = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/subscriptions`, ({ request }) => {
        url = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.subscriptions.list({
      status: 'past_due',
      pageSize: 25,
      contactId: 'cnt_42',
      priceId: 'price_7',
    })
    expect(url).toContain('status=past_due')
    expect(url).toContain('pageSize=25')
    expect(url).toContain('contactId=cnt_42')
    expect(url).toContain('priceId=price_7')
  })
})

describe('subscriptions.retrieve', () => {
  it('returns the unwrapped subscription', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/subscriptions/sub_123`, () =>
        HttpResponse.json({ data: sampleSubscription }),
      ),
    )
    const client = buildClient()
    const sub = await client.subscriptions.retrieve('sub_123', { fields: 'id,status' })
    expect(sub.id).toBe('sub_123')
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.subscriptions.retrieve('')).rejects.toThrow(TypeError)
  })

  it('throws ThreeCommonNotFoundError on 404', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/subscriptions/sub_missing`, () =>
        HttpResponse.json({ error: { code: 'not_found', message: 'gone' } }, { status: 404 }),
      ),
    )
    const client = buildClient()
    await expect(client.subscriptions.retrieve('sub_missing')).rejects.toBeInstanceOf(
      ThreeCommonNotFoundError,
    )
  })
})

describe('subscriptions.create', () => {
  it('POSTs the body and returns the unwrapped subscription', async () => {
    let body: unknown
    server.use(
      http.post(`${TEST_BASE_URL}/v1/subscriptions`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: sampleSubscription })
      }),
    )
    const client = buildClient()
    const created = await client.subscriptions.create({
      priceId: 'price_7',
      contactId: 'cnt_42',
      trialDays: 14,
    })
    expect(created.id).toBe('sub_123')
    expect(body).toEqual({ priceId: 'price_7', contactId: 'cnt_42', trialDays: 14 })
  })
})

describe('subscriptions.update', () => {
  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.subscriptions.update('', { quantity: 2 })).rejects.toThrow(TypeError)
  })

  it('PATCHes and unwraps subscription, invoice, proration', async () => {
    server.use(
      http.patch(`${TEST_BASE_URL}/v1/subscriptions/sub_123`, () =>
        HttpResponse.json({
          data: { ...sampleSubscription, quantity: 2 },
          invoice: sampleInvoiceRef,
          proration: { netAmountMinor: 1234, daysRemaining: 10, daysInCycle: 30 },
        }),
      ),
    )
    const client = buildClient()
    const result = await client.subscriptions.update('sub_123', { quantity: 2 })
    expect(result.subscription.quantity).toBe(2)
    expect(result.invoice?.id).toBe('inv_500')
    expect(result.proration.netAmountMinor).toBe(1234)
  })

  it('omits invoice when server returns none (downgrade)', async () => {
    server.use(
      http.patch(`${TEST_BASE_URL}/v1/subscriptions/sub_123`, () =>
        HttpResponse.json({
          data: sampleSubscription,
          proration: { netAmountMinor: 0, daysRemaining: 10, daysInCycle: 30 },
        }),
      ),
    )
    const client = buildClient()
    const result = await client.subscriptions.update('sub_123', { quantity: 1 })
    expect(result.invoice).toBeUndefined()
    expect(result.proration.netAmountMinor).toBe(0)
  })
})

describe('subscriptions.activate', () => {
  it('POSTs /activate and returns the activated sub', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/subscriptions/sub_123/activate`, () =>
        HttpResponse.json({ data: { ...sampleSubscription, status: 'active' } }),
      ),
    )
    const client = buildClient()
    const sub = await client.subscriptions.activate('sub_123')
    expect(sub.status).toBe('active')
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.subscriptions.activate('')).rejects.toThrow(TypeError)
  })
})

describe('subscriptions.cancel', () => {
  it('POSTs the reason to /cancel', async () => {
    let body: unknown
    server.use(
      http.post(`${TEST_BASE_URL}/v1/subscriptions/sub_123/cancel`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({
          data: { ...sampleSubscription, cancelAtPeriodEnd: true },
        })
      }),
    )
    const client = buildClient()
    const sub = await client.subscriptions.cancel('sub_123', { reason: 'Churn' })
    expect(sub.cancelAtPeriodEnd).toBe(true)
    expect(body).toEqual({ reason: 'Churn' })
  })

  it('accepts a missing body', async () => {
    let body: unknown
    server.use(
      http.post(`${TEST_BASE_URL}/v1/subscriptions/sub_123/cancel`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: sampleSubscription })
      }),
    )
    const client = buildClient()
    await client.subscriptions.cancel('sub_123')
    expect(body).toEqual({})
  })
})

describe('subscriptions.cancelImmediately', () => {
  it('POSTs to /cancel-immediately', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/subscriptions/sub_123/cancel-immediately`, () =>
        HttpResponse.json({
          data: { ...sampleSubscription, status: 'canceled', endedAt: '2026-05-25T00:00:00.000Z' },
        }),
      ),
    )
    const client = buildClient()
    const sub = await client.subscriptions.cancelImmediately('sub_123', { reason: 'Fraud' })
    expect(sub.status).toBe('canceled')
    expect(sub.endedAt).toBe('2026-05-25T00:00:00.000Z')
  })
})

describe('subscriptions.markUnpaid', () => {
  it('POSTs to /mark-unpaid', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/subscriptions/sub_123/mark-unpaid`, () =>
        HttpResponse.json({ data: { ...sampleSubscription, status: 'unpaid' } }),
      ),
    )
    const client = buildClient()
    const sub = await client.subscriptions.markUnpaid('sub_123')
    expect(sub.status).toBe('unpaid')
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.subscriptions.markUnpaid('')).rejects.toThrow(TypeError)
  })
})

describe('subscriptions.bill', () => {
  it('POSTs to /bill and unwraps subscription + invoice', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/subscriptions/sub_123/bill`, () =>
        HttpResponse.json({ data: sampleSubscription, invoice: sampleInvoiceRef }),
      ),
    )
    const client = buildClient()
    const result = await client.subscriptions.bill('sub_123')
    expect(result.subscription.id).toBe('sub_123')
    expect(result.invoice.id).toBe('inv_500')
  })
})

describe('subscriptions.renew', () => {
  it('POSTs to /renew and unwraps subscription + invoice when present', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/subscriptions/sub_123/renew`, () =>
        HttpResponse.json({ data: sampleSubscription, invoice: sampleInvoiceRef }),
      ),
    )
    const client = buildClient()
    const result = await client.subscriptions.renew('sub_123')
    expect(result.invoice?.id).toBe('inv_500')
  })

  it('omits invoice when the renewal cancels the sub', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/subscriptions/sub_123/renew`, () =>
        HttpResponse.json({ data: { ...sampleSubscription, status: 'canceled' } }),
      ),
    )
    const client = buildClient()
    const result = await client.subscriptions.renew('sub_123')
    expect(result.invoice).toBeUndefined()
    expect(result.subscription.status).toBe('canceled')
  })
})

describe('subscriptions.previewUpcomingInvoice', () => {
  it('returns the unwrapped preview', async () => {
    const preview = {
      customerId: 'cnt_42',
      subscriptionId: 'sub_123',
      currency: 'USD',
      lineItems: [{ description: 'Pro plan', quantity: 1, unitAmount: 50_000 }],
      subtotal: 50_000,
      total: 50_000,
      periodStart: '2026-06-01T00:00:00.000Z',
      periodEnd: '2026-07-01T00:00:00.000Z',
    }
    server.use(
      http.get(`${TEST_BASE_URL}/v1/subscriptions/sub_123/upcoming`, () =>
        HttpResponse.json({ data: { invoice: preview } }),
      ),
    )
    const client = buildClient()
    const result = await client.subscriptions.previewUpcomingInvoice('sub_123')
    expect(result?.total).toBe(50_000)
    expect(result?.lineItems).toHaveLength(1)
  })

  it('returns null when cancel-at-period-end', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/subscriptions/sub_123/upcoming`, () =>
        HttpResponse.json({ data: { invoice: null } }),
      ),
    )
    const client = buildClient()
    const result = await client.subscriptions.previewUpcomingInvoice('sub_123')
    expect(result).toBeNull()
  })
})

describe('subscriptions.listAutoPaginate', () => {
  it('iterates across pages', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/subscriptions`, ({ request }) => {
        const url = new URL(request.url)
        const page = Number(url.searchParams.get('page') ?? '0')
        if (page === 0) {
          return HttpResponse.json({
            data: [
              { ...sampleSubscription, id: 'sub_a' },
              { ...sampleSubscription, id: 'sub_b' },
            ],
            hasMore: true,
          })
        }
        return HttpResponse.json({
          data: [{ ...sampleSubscription, id: 'sub_c' }],
          hasMore: false,
        })
      }),
    )
    const client = buildClient()
    const ids: string[] = []
    for await (const sub of client.subscriptions.listAutoPaginate({ status: 'active' })) {
      if (sub.id !== undefined) ids.push(sub.id)
    }
    expect(ids).toEqual(['sub_a', 'sub_b', 'sub_c'])
  })
})
