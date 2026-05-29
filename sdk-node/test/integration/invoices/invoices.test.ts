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

const sampleInvoice = {
  id: 'inv_123',
  hostId: 'hst_1',
  customerId: 'cnt_42',
  number: null,
  currency: 'USD' as const,
  lineItems: [{ description: 'Consulting', quantity: 1, unitAmount: 50_000 }],
  payments: [],
  subtotal: 50_000,
  taxTotal: 0,
  total: 50_000,
  amountPaid: 0,
  amountDue: 50_000,
  status: 'draft' as const,
  createdAt: '2026-05-11T00:00:00.000Z',
  updatedAt: '2026-05-11T00:00:00.000Z',
}

describe('invoices.list', () => {
  it('returns data + hasMore', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/invoices`, () =>
        HttpResponse.json({ data: [sampleInvoice], hasMore: false }),
      ),
    )
    const client = buildClient()
    const result = await client.invoices.list({ status: 'open' })
    expect(result.hasMore).toBe(false)
    expect(result.data).toHaveLength(1)
    expect(result.data[0]?.id).toBe('inv_123')
  })

  it('forwards filters as query params', async () => {
    let url = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/invoices`, ({ request }) => {
        url = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.invoices.list({
      status: 'open',
      pageSize: 25,
      customerId: 'cnt_42',
      subscriptionId: 'sub_99',
      issuedAfter: '2026-01-01T00:00:00.000Z',
    })
    expect(url).toContain('status=open')
    expect(url).toContain('pageSize=25')
    expect(url).toContain('customerId=cnt_42')
    expect(url).toContain('subscriptionId=sub_99')
    expect(url).toContain('issuedAfter=2026-01-01T00%3A00%3A00.000Z')
  })
})

describe('invoices.retrieve', () => {
  it('returns the unwrapped invoice', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/invoices/inv_123`, () =>
        HttpResponse.json({ data: sampleInvoice }),
      ),
    )
    const client = buildClient()
    const invoice = await client.invoices.retrieve('inv_123', { fields: 'id,status' })
    expect(invoice.id).toBe('inv_123')
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.invoices.retrieve('')).rejects.toThrow(TypeError)
  })

  it('throws ThreeCommonNotFoundError on 404', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/invoices/inv_missing`, () =>
        HttpResponse.json({ error: { code: 'not_found', message: 'gone' } }, { status: 404 }),
      ),
    )
    const client = buildClient()
    await expect(client.invoices.retrieve('inv_missing')).rejects.toBeInstanceOf(
      ThreeCommonNotFoundError,
    )
  })
})

describe('invoices.create', () => {
  it('POSTs the body and returns the unwrapped invoice', async () => {
    let body: unknown
    server.use(
      http.post(`${TEST_BASE_URL}/v1/invoices`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: sampleInvoice })
      }),
    )
    const client = buildClient()
    const created = await client.invoices.create({
      customerId: 'cnt_42',
      currency: 'USD',
      lineItems: [{ description: 'Consulting', quantity: 1, unitAmount: 50_000 }],
    })
    expect(created.id).toBe('inv_123')
    expect(body).toEqual({
      customerId: 'cnt_42',
      currency: 'USD',
      lineItems: [{ description: 'Consulting', quantity: 1, unitAmount: 50_000 }],
    })
  })
})

describe('invoices.update', () => {
  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.invoices.update('', { notes: 'x' })).rejects.toThrow(TypeError)
  })

  it('PATCHes and returns the unwrapped invoice', async () => {
    server.use(
      http.patch(`${TEST_BASE_URL}/v1/invoices/inv_123`, () =>
        HttpResponse.json({ data: { ...sampleInvoice, notes: 'Net 30' } }),
      ),
    )
    const client = buildClient()
    const updated = await client.invoices.update('inv_123', { notes: 'Net 30' })
    expect(updated.notes).toBe('Net 30')
  })
})

describe('invoices.finalize', () => {
  it('POSTs to /finalize and returns the issued invoice', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/invoices/inv_123/finalize`, () =>
        HttpResponse.json({
          data: { ...sampleInvoice, status: 'open', number: 'INV-0001' },
        }),
      ),
    )
    const client = buildClient()
    const issued = await client.invoices.finalize('inv_123')
    expect(issued.status).toBe('open')
    expect(issued.number).toBe('INV-0001')
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.invoices.finalize('')).rejects.toThrow(TypeError)
  })
})

describe('invoices.void', () => {
  it('POSTs the reason to /void', async () => {
    let body: unknown
    server.use(
      http.post(`${TEST_BASE_URL}/v1/invoices/inv_123/void`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: { ...sampleInvoice, status: 'void' } })
      }),
    )
    const client = buildClient()
    const voided = await client.invoices.void('inv_123', { reason: 'Sent in error' })
    expect(voided.status).toBe('void')
    expect(body).toEqual({ reason: 'Sent in error' })
  })

  it('accepts a missing body', async () => {
    let body: unknown
    server.use(
      http.post(`${TEST_BASE_URL}/v1/invoices/inv_123/void`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: { ...sampleInvoice, status: 'void' } })
      }),
    )
    const client = buildClient()
    await client.invoices.void('inv_123')
    expect(body).toEqual({})
  })
})

describe('invoices.recordPayment', () => {
  it('POSTs the payment body', async () => {
    let body: unknown
    server.use(
      http.post(`${TEST_BASE_URL}/v1/invoices/inv_123/payments`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({
          data: {
            ...sampleInvoice,
            amountPaid: 50_000,
            amountDue: 0,
            status: 'paid',
          },
        })
      }),
    )
    const client = buildClient()
    const paid = await client.invoices.recordPayment('inv_123', {
      payment: 50_000,
      idempotencyKey: 'pmt-4310',
    })
    expect(paid.status).toBe('paid')
    expect(paid.amountDue).toBe(0)
    expect(body).toEqual({ payment: 50_000, idempotencyKey: 'pmt-4310' })
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.invoices.recordPayment('', { payment: 1 })).rejects.toThrow(TypeError)
  })
})

describe('invoices.autoCharge', () => {
  it('POSTs to /auto_charge and returns { invoice, outcome } when the charge clears', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/invoices/inv_123/auto_charge`, () =>
        HttpResponse.json({
          data: { ...sampleInvoice, status: 'paid', amountPaid: 50_000, amountDue: 0 },
          outcome: 'paid',
        }),
      ),
    )
    const client = buildClient()
    const result = await client.invoices.autoCharge('inv_123')
    expect(result.outcome).toBe('paid')
    expect(result.invoice.status).toBe('paid')
    expect(result.invoice.amountDue).toBe(0)
    // No failureCode key on the happy path, not `failureCode: undefined`.
    expect('failureCode' in result).toBe(false)
  })

  it('surfaces a decline as outcome:failed + failureCode rather than throwing', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/invoices/inv_123/auto_charge`, () =>
        HttpResponse.json({
          data: { ...sampleInvoice, status: 'payment_failed' },
          outcome: 'failed',
          failureCode: 'card_declined',
        }),
      ),
    )
    const client = buildClient()
    const result = await client.invoices.autoCharge('inv_123')
    expect(result.outcome).toBe('failed')
    expect(result.invoice.status).toBe('payment_failed')
    expect(result.failureCode).toBe('card_declined')
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.invoices.autoCharge('')).rejects.toThrow(TypeError)
  })
})

describe('invoices.refundPayment', () => {
  it('POSTs the body to /payments/:paymentId/refunds and returns the unwrapped invoice', async () => {
    let body: unknown
    let capturedPath = ''
    server.use(
      http.post(
        `${TEST_BASE_URL}/v1/invoices/inv_123/payments/pay_456/refunds`,
        async ({ request }) => {
          body = await request.json()
          capturedPath = new URL(request.url).pathname
          return HttpResponse.json({ data: { ...sampleInvoice, status: 'paid' } })
        },
      ),
    )
    const client = buildClient()
    const refunded = await client.invoices.refundPayment('inv_123', 'pay_456', {
      amount: 25_000,
      reason: 'requested_by_customer',
      idempotencyKey: 'rfnd-1',
    })
    expect(refunded.id).toBe('inv_123')
    expect(capturedPath).toBe('/v1/invoices/inv_123/payments/pay_456/refunds')
    expect(body).toEqual({
      amount: 25_000,
      reason: 'requested_by_customer',
      idempotencyKey: 'rfnd-1',
    })
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.invoices.refundPayment('', 'pay_456', { amount: 1 })).rejects.toThrow(
      TypeError,
    )
  })

  it('rejects empty paymentId', async () => {
    const client = buildClient()
    await expect(client.invoices.refundPayment('inv_123', '', { amount: 1 })).rejects.toThrow(
      TypeError,
    )
  })
})

describe('invoices.deleteDraft', () => {
  it('DELETEs /invoices/:id and returns the deleted id', async () => {
    let method = ''
    server.use(
      http.delete(`${TEST_BASE_URL}/v1/invoices/inv_123`, ({ request }) => {
        method = request.method
        return HttpResponse.json({ data: { id: 'inv_123' } })
      }),
    )
    const client = buildClient()
    const result = await client.invoices.deleteDraft('inv_123')
    expect(method).toBe('DELETE')
    expect(result.id).toBe('inv_123')
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.invoices.deleteDraft('')).rejects.toThrow(TypeError)
  })
})

describe('invoices.list — paramsToQuery edge cases', () => {
  it('skips explicit undefined params', async () => {
    let captured = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/invoices`, ({ request }) => {
        captured = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.invoices.list({
      status: 'open',
      pageSize: undefined,
      customerId: undefined,
    } as unknown as Parameters<typeof client.invoices.list>[0])
    expect(captured).toContain('status=open')
    expect(captured).not.toContain('pageSize')
    expect(captured).not.toContain('customerId')
  })

  it('drops non-primitive values silently', async () => {
    let captured = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/invoices`, ({ request }) => {
        captured = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.invoices.list({
      status: 'open',
      forbidden: ['x', 'y'],
    } as unknown as Parameters<typeof client.invoices.list>[0])
    expect(captured).toContain('status=open')
    expect(captured).not.toContain('forbidden')
  })
})

describe('invoices.listAutoPaginate', () => {
  it('iterates across pages', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/invoices`, ({ request }) => {
        const url = new URL(request.url)
        const page = Number(url.searchParams.get('page') ?? '0')
        if (page === 0) {
          return HttpResponse.json({
            data: [
              { ...sampleInvoice, id: 'inv_a' },
              { ...sampleInvoice, id: 'inv_b' },
            ],
            hasMore: true,
          })
        }
        return HttpResponse.json({
          data: [{ ...sampleInvoice, id: 'inv_c' }],
          hasMore: false,
        })
      }),
    )
    const client = buildClient()
    const ids: string[] = []
    for await (const inv of client.invoices.listAutoPaginate({ status: 'open' })) {
      if (inv.id !== undefined) ids.push(inv.id)
    }
    expect(ids).toEqual(['inv_a', 'inv_b', 'inv_c'])
  })
})
