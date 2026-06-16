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

const sampleProperty = {
  type: 'Select One' as const,
  id: 'prop_123',
  name: 'T-shirt size',
  description: 'Preferred shirt size.',
  status: 'active' as const,
  objectType: 'contact' as const,
  options: [
    { value: 's', label: 'Small' },
    { value: 'm', label: 'Medium' },
  ],
}

describe('properties.list', () => {
  it('returns data + hasMore', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/properties`, () =>
        HttpResponse.json({ data: [sampleProperty], hasMore: false }),
      ),
    )
    const client = buildClient()
    const result = await client.properties.list({ objectType: 'contact', status: 'active' })
    expect(result.hasMore).toBe(false)
    expect(result.data).toHaveLength(1)
    expect(result.data[0]?.id).toBe('prop_123')
    expect(result.data[0]?.type).toBe('Select One')
  })

  it('forwards query params', async () => {
    let url = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/properties`, ({ request }) => {
        url = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.properties.list({
      objectType: 'contact',
      propertyType: 'Select One',
      status: 'archived',
      sort: 'name',
      order: 'desc',
      search: 'size',
      pageSize: 25,
    })
    expect(url).toContain('objectType=contact')
    expect(url).toContain('propertyType=Select+One')
    expect(url).toContain('status=archived')
    expect(url).toContain('sort=name')
    expect(url).toContain('order=desc')
    expect(url).toContain('search=size')
    expect(url).toContain('pageSize=25')
  })
})

describe('properties.retrieve', () => {
  it('returns the unwrapped property', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/properties/prop_123`, () =>
        HttpResponse.json({ data: sampleProperty }),
      ),
    )
    const client = buildClient()
    const property = await client.properties.retrieve('prop_123')
    expect(property.id).toBe('prop_123')
    expect(property.name).toBe('T-shirt size')
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.properties.retrieve('')).rejects.toThrow(TypeError)
  })

  it('throws ThreeCommonNotFoundError on 404', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/properties/prop_missing`, () =>
        HttpResponse.json({ error: { code: 'not_found', message: 'gone' } }, { status: 404 }),
      ),
    )
    const client = buildClient()
    await expect(client.properties.retrieve('prop_missing')).rejects.toBeInstanceOf(
      ThreeCommonNotFoundError,
    )
  })
})

describe('properties.create', () => {
  it('POSTs the body and returns the unwrapped property (201)', async () => {
    let body: unknown
    server.use(
      http.post(`${TEST_BASE_URL}/v1/properties`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: sampleProperty }, { status: 201 })
      }),
    )
    const client = buildClient()
    const property = await client.properties.create({
      type: 'Select One',
      name: 'T-shirt size',
      status: 'active',
      objectType: 'contact',
      options: [
        { value: 's', label: 'Small' },
        { value: 'm', label: 'Medium' },
      ],
    })
    expect(property.id).toBe('prop_123')
    expect(body).toEqual({
      type: 'Select One',
      name: 'T-shirt size',
      status: 'active',
      objectType: 'contact',
      options: [
        { value: 's', label: 'Small' },
        { value: 'm', label: 'Medium' },
      ],
    })
  })

  it('surfaces 400 validation errors', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/properties`, () =>
        HttpResponse.json(
          { error: { code: 'validation_error', message: 'options required' } },
          { status: 400 },
        ),
      ),
    )
    const client = buildClient()
    await expect(
      client.properties.create({
        type: 'Select One',
        name: 'T-shirt size',
        status: 'active',
        objectType: 'contact',
      }),
    ).rejects.toBeInstanceOf(ThreeCommonValidationError)
  })
})

describe('properties.update', () => {
  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.properties.update('', { name: 'x' })).rejects.toThrow(TypeError)
  })

  it('PATCHes and returns the unwrapped property', async () => {
    let body: unknown
    server.use(
      http.patch(`${TEST_BASE_URL}/v1/properties/prop_123`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({
          data: { ...sampleProperty, name: 'Shirt size', description: undefined },
        })
      }),
    )
    const client = buildClient()
    const property = await client.properties.update('prop_123', {
      name: 'Shirt size',
      description: null,
    })
    expect(property.name).toBe('Shirt size')
    expect(body).toEqual({ name: 'Shirt size', description: null })
  })
})

describe('properties.list - paramsToQuery edge cases', () => {
  it('skips explicit undefined params', async () => {
    let captured = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/properties`, ({ request }) => {
        captured = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.properties.list({
      objectType: 'contact',
      status: undefined,
      pageSize: undefined,
    } as unknown as Parameters<typeof client.properties.list>[0])
    expect(captured).toContain('objectType=contact')
    expect(captured).not.toContain('status')
    expect(captured).not.toContain('pageSize')
  })

  it('drops non-primitive values silently rather than passing them on the wire', async () => {
    let captured = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/properties`, ({ request }) => {
        captured = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.properties.list({
      objectType: 'contact',
      forbidden: ['x', 'y'],
    } as unknown as Parameters<typeof client.properties.list>[0])
    expect(captured).toContain('objectType=contact')
    expect(captured).not.toContain('forbidden')
  })
})

describe('properties.listAutoPaginate', () => {
  it('iterates across pages', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/properties`, ({ request }) => {
        const url = new URL(request.url)
        const page = Number(url.searchParams.get('page') ?? '0')
        if (page === 0) {
          return HttpResponse.json({
            data: [
              { ...sampleProperty, id: 'prop_a' },
              { ...sampleProperty, id: 'prop_b' },
            ],
            hasMore: true,
          })
        }
        if (page === 1) {
          return HttpResponse.json({ data: [{ ...sampleProperty, id: 'prop_c' }], hasMore: false })
        }
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    const ids: string[] = []
    for await (const property of client.properties.listAutoPaginate({ objectType: 'contact' })) {
      ids.push(property.id)
    }
    expect(ids).toEqual(['prop_a', 'prop_b', 'prop_c'])
  })
})
