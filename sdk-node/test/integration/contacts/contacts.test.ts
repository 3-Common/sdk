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

const sampleContact = {
  id: 'cnt_123',
  firstName: 'Alex',
  lastName: 'Garcia',
  fullName: 'Alex Garcia',
  email: 'alex@example.com',
  phone: '+15555550123',
  vendorId: 'hst_1',
  orderSum: 3,
  grossSum: 15_000,
  firstOrder: 1_700_000_000_000,
  lastOrder: 1_710_000_000_000,
  createdAt: '2026-01-01T00:00:00.000Z',
  status: 'opted-in' as const,
  eventsAttended_IDS: ['evt_a', 'evt_b'],
  itemsPurchased_IDS: [],
  productsPurchased_IDS: [],
}

describe('contacts.list', () => {
  it('returns data + hasMore + page info', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/contacts`, () =>
        HttpResponse.json({
          data: [sampleContact],
          hasMore: false,
          pageNumber: 0,
          pageSize: 20,
        }),
      ),
    )
    const client = buildClient()
    const result = await client.contacts.list({ filter: 'opted-in' })
    expect(result.hasMore).toBe(false)
    expect(result.pageNumber).toBe(0)
    expect(result.pageSize).toBe(20)
    expect(result.data).toHaveLength(1)
    expect(result.data[0]?.email).toBe('alex@example.com')
  })

  it('forwards filters as query params', async () => {
    let url = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/contacts`, ({ request }) => {
        url = request.url
        return HttpResponse.json({ data: [], hasMore: false, pageNumber: 0, pageSize: 20 })
      }),
    )
    const client = buildClient()
    await client.contacts.list({
      filter: 'opted-in',
      pageNumber: 2,
      pageSize: 100,
      sortField: 'grossSum',
      sortDirection: 'desc',
      search: 'garcia',
    })
    expect(url).toContain('filter=opted-in')
    expect(url).toContain('pageNumber=2')
    expect(url).toContain('pageSize=100')
    expect(url).toContain('sortField=grossSum')
    expect(url).toContain('sortDirection=desc')
    expect(url).toContain('search=garcia')
  })

  it('skips undefined params', async () => {
    let captured = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/contacts`, ({ request }) => {
        captured = request.url
        return HttpResponse.json({ data: [], hasMore: false, pageNumber: 0, pageSize: 20 })
      }),
    )
    const client = buildClient()
    await client.contacts.list({
      filter: 'opted-in',
      pageSize: undefined,
      search: undefined,
    } as unknown as Parameters<typeof client.contacts.list>[0])
    expect(captured).toContain('filter=opted-in')
    expect(captured).not.toContain('pageSize')
    expect(captured).not.toContain('search')
  })

  it('drops non-primitive params', async () => {
    let captured = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/contacts`, ({ request }) => {
        captured = request.url
        return HttpResponse.json({ data: [], hasMore: false, pageNumber: 0, pageSize: 20 })
      }),
    )
    const client = buildClient()
    await client.contacts.list({
      filter: 'opted-in',
      forbidden: ['x', 'y'],
    } as unknown as Parameters<typeof client.contacts.list>[0])
    expect(captured).toContain('filter=opted-in')
    expect(captured).not.toContain('forbidden')
  })
})

describe('contacts.count', () => {
  it('unwraps the count from the envelope', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/contacts/count`, () =>
        HttpResponse.json({ data: { count: 4823 } }),
      ),
    )
    const client = buildClient()
    const result = await client.contacts.count()
    expect(result.count).toBe(4823)
  })
})

describe('contacts.retrieve', () => {
  it('returns the unwrapped contact', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/contacts/cnt_123`, () =>
        HttpResponse.json({ data: sampleContact }),
      ),
    )
    const client = buildClient()
    const contact = await client.contacts.retrieve('cnt_123')
    expect(contact.id).toBe('cnt_123')
    expect(contact.email).toBe('alex@example.com')
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.contacts.retrieve('')).rejects.toThrow(TypeError)
  })

  it('throws ThreeCommonNotFoundError on 404', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/contacts/cnt_missing`, () =>
        HttpResponse.json({ error: { code: 'not_found', message: 'gone' } }, { status: 404 }),
      ),
    )
    const client = buildClient()
    await expect(client.contacts.retrieve('cnt_missing')).rejects.toBeInstanceOf(
      ThreeCommonNotFoundError,
    )
  })
})

describe('contacts.create', () => {
  it('POSTs the body and returns the unwrapped contact', async () => {
    let body: unknown
    server.use(
      http.post(`${TEST_BASE_URL}/v1/contacts`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: sampleContact })
      }),
    )
    const client = buildClient()
    const created = await client.contacts.create({
      email: 'alex@example.com',
      firstName: 'Alex',
      lastName: 'Garcia',
    })
    expect(created.id).toBe('cnt_123')
    expect(body).toEqual({
      email: 'alex@example.com',
      firstName: 'Alex',
      lastName: 'Garcia',
    })
  })

  it('throws ThreeCommonConflictError on duplicate email', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/contacts`, () =>
        HttpResponse.json(
          { error: { code: 'conflict', message: 'duplicate email' } },
          { status: 409 },
        ),
      ),
    )
    const client = buildClient()
    await expect(client.contacts.create({ email: 'alex@example.com' })).rejects.toBeInstanceOf(
      ThreeCommonConflictError,
    )
  })
})

describe('contacts.update', () => {
  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(
      client.contacts.update('', {
        contact: {
          firstName: 'A',
          lastName: 'B',
          email: 'x@example.com',
          status: 'opted-in',
        },
      }),
    ).rejects.toThrow(TypeError)
  })

  it('PATCHes and returns the order-details projection', async () => {
    let body: unknown
    server.use(
      http.patch(`${TEST_BASE_URL}/v1/contacts/cnt_123`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({
          data: {
            _id: 'cnt_123',
            email: 'a.garcia@example.com',
            vendorId: 'hst_1',
            firstName: 'Alex',
            lastName: 'Garcia',
            fullName: 'Alex Garcia',
            status: 'opted-in' as const,
            grossSum: 15_000,
            orderSum: 3,
            events_attended: [],
            items_purchased: [],
            products_purchased: [],
          },
        })
      }),
    )
    const client = buildClient()
    const updated = await client.contacts.update('cnt_123', {
      contact: {
        firstName: 'Alex',
        lastName: 'Garcia',
        email: 'a.garcia@example.com',
        status: 'opted-in',
      },
    })
    expect(updated._id).toBe('cnt_123')
    expect(updated.email).toBe('a.garcia@example.com')
    expect(body).toEqual({
      contact: {
        firstName: 'Alex',
        lastName: 'Garcia',
        email: 'a.garcia@example.com',
        status: 'opted-in',
      },
    })
  })

  it('sends mergeWith and resolution when provided', async () => {
    let body: unknown
    server.use(
      http.patch(`${TEST_BASE_URL}/v1/contacts/cnt_123`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({
          data: {
            _id: 'cnt_123',
            email: 'a@example.com',
            vendorId: 'hst_1',
            firstName: 'A',
            lastName: 'G',
            fullName: 'A G',
            status: 'opted-in' as const,
            grossSum: 0,
            orderSum: 0,
            events_attended: [],
            items_purchased: [],
            products_purchased: [],
          },
        })
      }),
    )
    const client = buildClient()
    await client.contacts.update('cnt_123', {
      contact: { firstName: 'A', lastName: 'G', email: 'a@example.com', status: 'opted-in' },
      mergeWith: 'cnt_456',
      resolution: 'safe-merge',
    })
    expect(body).toMatchObject({ mergeWith: 'cnt_456', resolution: 'safe-merge' })
  })
})

describe('contacts.delete', () => {
  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.contacts.delete('')).rejects.toThrow(TypeError)
  })

  it('DELETEs and returns the id', async () => {
    server.use(
      http.delete(`${TEST_BASE_URL}/v1/contacts/cnt_123`, () =>
        HttpResponse.json({ data: { id: 'cnt_123' } }),
      ),
    )
    const client = buildClient()
    const result = await client.contacts.delete('cnt_123')
    expect(result.id).toBe('cnt_123')
  })

  it('surfaces 404 on missing contact', async () => {
    server.use(
      http.delete(`${TEST_BASE_URL}/v1/contacts/cnt_missing`, () =>
        HttpResponse.json({ error: { code: 'not_found', message: 'gone' } }, { status: 404 }),
      ),
    )
    const client = buildClient()
    await expect(client.contacts.delete('cnt_missing')).rejects.toBeInstanceOf(
      ThreeCommonNotFoundError,
    )
  })
})

describe('contacts.bulkUpsert', () => {
  it('POSTs the array and returns the affected count', async () => {
    let body: unknown
    server.use(
      http.post(`${TEST_BASE_URL}/v1/contacts/bulk`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: { affected: 2 } })
      }),
    )
    const client = buildClient()
    const result = await client.contacts.bulkUpsert({
      contacts: [
        { email: 'a@example.com', firstName: 'Ada' },
        { email: 'b@example.com', firstName: 'Beatrix' },
      ],
    })
    expect(result.affected).toBe(2)
    expect(body).toEqual({
      contacts: [
        { email: 'a@example.com', firstName: 'Ada' },
        { email: 'b@example.com', firstName: 'Beatrix' },
      ],
    })
  })
})

describe('contacts.listActivity', () => {
  it('returns the activity feed with paging info', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/contacts/cnt_123/activity`, () =>
        HttpResponse.json({
          data: [
            {
              _id: 'act_1',
              vendor_id: 'hst_1',
              email: 'alex@example.com',
              contact_id: 'cnt_123',
              type: 'checkout_session_completed' as const,
              data: { orderId: 'ord_1' },
              createdAt: '2026-05-01T00:00:00.000Z',
              updatedAt: '2026-05-01T00:00:00.000Z',
            },
          ],
          hasMore: false,
          pageNumber: 0,
          pageSize: 20,
        }),
      ),
    )
    const client = buildClient()
    const result = await client.contacts.listActivity('cnt_123', {
      filter: 'checkout_session_completed',
    })
    expect(result.data).toHaveLength(1)
    expect(result.data[0]?.type).toBe('checkout_session_completed')
    expect(result.hasMore).toBe(false)
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.contacts.listActivity('')).rejects.toThrow(TypeError)
  })

  it('surfaces 404 on missing contact', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/contacts/cnt_missing/activity`, () =>
        HttpResponse.json({ error: { code: 'not_found', message: 'gone' } }, { status: 404 }),
      ),
    )
    const client = buildClient()
    await expect(client.contacts.listActivity('cnt_missing')).rejects.toBeInstanceOf(
      ThreeCommonNotFoundError,
    )
  })
})

describe('contacts.listAutoPaginate', () => {
  it('walks pageNumber → pageNumber+1 until hasMore is false', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/contacts`, ({ request }) => {
        const url = new URL(request.url)
        const page = Number(url.searchParams.get('pageNumber') ?? '0')
        if (page === 0) {
          return HttpResponse.json({
            data: [
              { ...sampleContact, id: 'cnt_a' },
              { ...sampleContact, id: 'cnt_b' },
            ],
            hasMore: true,
            pageNumber: 0,
            pageSize: 20,
          })
        }
        return HttpResponse.json({
          data: [{ ...sampleContact, id: 'cnt_c' }],
          hasMore: false,
          pageNumber: 1,
          pageSize: 20,
        })
      }),
    )
    const client = buildClient()
    const ids: string[] = []
    for await (const contact of client.contacts.listAutoPaginate({ filter: 'opted-in' })) {
      ids.push(contact.id)
    }
    expect(ids).toEqual(['cnt_a', 'cnt_b', 'cnt_c'])
  })
})

describe('contacts.listAutoPaginate — explicit pageNumber start', () => {
  it('honors a non-zero starting pageNumber', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/contacts`, ({ request }) => {
        const url = new URL(request.url)
        const page = Number(url.searchParams.get('pageNumber') ?? '0')
        expect(page).toBe(5)
        return HttpResponse.json({
          data: [{ ...sampleContact, id: 'cnt_p5' }],
          hasMore: false,
          pageNumber: 5,
          pageSize: 20,
        })
      }),
    )
    const client = buildClient()
    const ids: string[] = []
    for await (const c of client.contacts.listAutoPaginate({ pageNumber: 5 })) {
      ids.push(c.id)
    }
    expect(ids).toEqual(['cnt_p5'])
  })
})

describe('contacts.listActivity — paramsToQuery edge cases', () => {
  it('skips undefined and drops non-primitive params', async () => {
    let captured = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/contacts/cnt_123/activity`, ({ request }) => {
        captured = request.url
        return HttpResponse.json({ data: [], hasMore: false, pageNumber: 0, pageSize: 20 })
      }),
    )
    const client = buildClient()
    await client.contacts.listActivity('cnt_123', {
      filter: 'email_sent',
      pageSize: undefined,
      forbidden: [1, 2],
    } as unknown as Parameters<typeof client.contacts.listActivity>[1])
    expect(captured).toContain('filter=email_sent')
    expect(captured).not.toContain('pageSize')
    expect(captured).not.toContain('forbidden')
  })
})

describe('contacts.listActivityAutoPaginate — explicit pageNumber start', () => {
  it('honors a non-zero starting pageNumber', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/contacts/cnt_123/activity`, ({ request }) => {
        const url = new URL(request.url)
        expect(url.searchParams.get('pageNumber')).toBe('3')
        return HttpResponse.json({
          data: [
            {
              _id: 'act_p3',
              vendor_id: 'hst_1',
              email: 'a@example.com',
              type: 'email_sent' as const,
              data: {},
              createdAt: '2026-05-01T00:00:00.000Z',
              updatedAt: '2026-05-01T00:00:00.000Z',
            },
          ],
          hasMore: false,
          pageNumber: 3,
          pageSize: 20,
        })
      }),
    )
    const client = buildClient()
    const ids: string[] = []
    for await (const evt of client.contacts.listActivityAutoPaginate('cnt_123', {
      pageNumber: 3,
    })) {
      ids.push(evt._id)
    }
    expect(ids).toEqual(['act_p3'])
  })
})

describe('contacts.listActivityAutoPaginate', () => {
  it('walks activity pages and rejects empty id', async () => {
    const client = buildClient()
    expect(() => client.contacts.listActivityAutoPaginate('')).toThrow(TypeError)

    server.use(
      http.get(`${TEST_BASE_URL}/v1/contacts/cnt_123/activity`, ({ request }) => {
        const url = new URL(request.url)
        const page = Number(url.searchParams.get('pageNumber') ?? '0')
        if (page === 0) {
          return HttpResponse.json({
            data: [
              {
                _id: 'act_1',
                vendor_id: 'hst_1',
                email: 'alex@example.com',
                type: 'email_sent' as const,
                data: {},
                createdAt: '2026-05-01T00:00:00.000Z',
                updatedAt: '2026-05-01T00:00:00.000Z',
              },
            ],
            hasMore: true,
            pageNumber: 0,
            pageSize: 20,
          })
        }
        return HttpResponse.json({
          data: [
            {
              _id: 'act_2',
              vendor_id: 'hst_1',
              email: 'alex@example.com',
              type: 'ticket_scanned' as const,
              data: {},
              createdAt: '2026-05-02T00:00:00.000Z',
              updatedAt: '2026-05-02T00:00:00.000Z',
            },
          ],
          hasMore: false,
          pageNumber: 1,
          pageSize: 20,
        })
      }),
    )
    const ids: string[] = []
    for await (const evt of client.contacts.listActivityAutoPaginate('cnt_123')) {
      ids.push(evt._id)
    }
    expect(ids).toEqual(['act_1', 'act_2'])
  })
})
