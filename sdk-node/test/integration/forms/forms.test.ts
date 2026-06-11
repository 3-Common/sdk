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

const sampleForm = {
  id: 'frm_123',
  name: 'Customer survey',
  nameHidden: false,
  ownerId: 'hst_1',
  status: 'active' as const,
  rows: [],
  submitButtonText: 'Submit',
  submitButtonWidth: 'auto' as const,
  type: 'standalone' as const,
  elements: [],
}

const sampleElement = {
  id: 'elm_123',
  prompt: 'What is your name?',
  type: 'Text' as const,
  required: true,
}

describe('forms.list', () => {
  it('returns data + hasMore', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/forms`, () =>
        HttpResponse.json({
          data: [
            { id: 'frm_a', name: 'Survey', numElements: 4, type: 'standalone', status: 'active' },
          ],
          hasMore: false,
        }),
      ),
    )
    const client = buildClient()
    const result = await client.forms.list({ type: 'standalone' })
    expect(result.hasMore).toBe(false)
    expect(result.data).toHaveLength(1)
    expect(result.data[0]?.id).toBe('frm_a')
  })

  it('defaults params to an empty object', async () => {
    let captured = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/forms`, ({ request }) => {
        captured = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    const result = await client.forms.list()
    expect(result.data).toHaveLength(0)
    expect(captured).toContain('/v1/forms')
  })

  it('forwards params as query and skips undefined / non-primitive values', async () => {
    let captured = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/forms`, ({ request }) => {
        captured = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.forms.list({
      type: 'standalone',
      page: 2,
      pageSize: undefined,
      forbidden: ['x', 'y'],
    } as unknown as Parameters<typeof client.forms.list>[0])
    expect(captured).toContain('type=standalone')
    expect(captured).toContain('page=2')
    expect(captured).not.toContain('pageSize')
    expect(captured).not.toContain('forbidden')
  })
})

describe('forms.create', () => {
  it('POSTs the body and returns the unwrapped form', async () => {
    let body: unknown
    server.use(
      http.post(`${TEST_BASE_URL}/v1/forms`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: sampleForm })
      }),
    )
    const client = buildClient()
    const created = await client.forms.create({ name: 'Customer survey', type: 'standalone' })
    expect(created.id).toBe('frm_123')
    expect(body).toEqual({ name: 'Customer survey', type: 'standalone' })
  })

  it('throws ThreeCommonValidationError on 400', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/forms`, () =>
        HttpResponse.json(
          { error: { code: 'validation_error', message: 'name is required' } },
          { status: 400 },
        ),
      ),
    )
    const client = buildClient()
    await expect(
      client.forms.create({ type: 'standalone' } as unknown as Parameters<
        typeof client.forms.create
      >[0]),
    ).rejects.toBeInstanceOf(ThreeCommonValidationError)
  })
})

describe('forms.retrieve', () => {
  it('returns the unwrapped form', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/forms/frm_123`, () => HttpResponse.json({ data: sampleForm })),
    )
    const client = buildClient()
    const form = await client.forms.retrieve('frm_123')
    expect(form.id).toBe('frm_123')
    expect(form.name).toBe('Customer survey')
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.forms.retrieve('')).rejects.toThrow(TypeError)
  })

  it('throws ThreeCommonNotFoundError on 404', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/forms/frm_missing`, () =>
        HttpResponse.json({ error: { code: 'not_found', message: 'gone' } }, { status: 404 }),
      ),
    )
    const client = buildClient()
    await expect(client.forms.retrieve('frm_missing')).rejects.toBeInstanceOf(
      ThreeCommonNotFoundError,
    )
  })
})

describe('forms.update', () => {
  it('PATCHes the body and returns the updated form', async () => {
    let body: unknown
    server.use(
      http.patch(`${TEST_BASE_URL}/v1/forms/frm_123`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: { ...sampleForm, name: 'Renamed survey' } })
      }),
    )
    const client = buildClient()
    const updated = await client.forms.update('frm_123', {
      name: 'Renamed survey',
      status: 'active',
    })
    expect(updated.name).toBe('Renamed survey')
    expect(body).toEqual({ name: 'Renamed survey', status: 'active' })
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.forms.update('', { name: 'x' })).rejects.toThrow(TypeError)
  })
})

describe('forms.duplicate', () => {
  it('POSTs to /duplicate and returns the copy', async () => {
    let body: unknown
    server.use(
      http.post(`${TEST_BASE_URL}/v1/forms/frm_123/duplicate`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: { ...sampleForm, id: 'frm_copy', name: 'Survey (copy)' } })
      }),
    )
    const client = buildClient()
    const copy = await client.forms.duplicate('frm_123', { name: 'Survey (copy)' })
    expect(copy.id).toBe('frm_copy')
    expect(body).toEqual({ name: 'Survey (copy)' })
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.forms.duplicate('', { name: 'x' })).rejects.toThrow(TypeError)
  })
})

describe('forms.addElement', () => {
  it('POSTs the element and returns it', async () => {
    let body: unknown
    server.use(
      http.post(`${TEST_BASE_URL}/v1/forms/frm_123/elements`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: sampleElement })
      }),
    )
    const client = buildClient()
    const element = await client.forms.addElement('frm_123', {
      prompt: 'What is your name?',
      type: 'Text',
      required: true,
    })
    expect(element.id).toBe('elm_123')
    expect(element.type).toBe('Text')
    expect(body).toEqual({ prompt: 'What is your name?', type: 'Text', required: true })
  })

  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(
      client.forms.addElement('', { prompt: 'x', type: 'Text', required: true }),
    ).rejects.toThrow(TypeError)
  })

  it('throws ThreeCommonValidationError on 400', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/forms/frm_123/elements`, () =>
        HttpResponse.json(
          { error: { code: 'validation_error', message: 'prompt is required' } },
          { status: 400 },
        ),
      ),
    )
    const client = buildClient()
    await expect(
      client.forms.addElement('frm_123', { type: 'Text', required: true } as unknown as Parameters<
        typeof client.forms.addElement
      >[1]),
    ).rejects.toBeInstanceOf(ThreeCommonValidationError)
  })
})

describe('forms.updateElement', () => {
  it('PATCHes the element and returns it', async () => {
    let body: unknown
    server.use(
      http.patch(`${TEST_BASE_URL}/v1/forms/frm_123/elements/elm_123`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: { ...sampleElement, prompt: 'What is your full name?' } })
      }),
    )
    const client = buildClient()
    const element = await client.forms.updateElement('frm_123', 'elm_123', {
      prompt: 'What is your full name?',
    })
    expect(element.prompt).toBe('What is your full name?')
    expect(body).toEqual({ prompt: 'What is your full name?' })
  })

  it('rejects empty id and empty elementId', async () => {
    const client = buildClient()
    await expect(client.forms.updateElement('', 'elm_123', {})).rejects.toThrow(TypeError)
    await expect(client.forms.updateElement('frm_123', '', {})).rejects.toThrow(TypeError)
  })
})

describe('forms.deleteElement', () => {
  it('DELETEs and returns the deleted element id', async () => {
    server.use(
      http.delete(`${TEST_BASE_URL}/v1/forms/frm_123/elements/elm_123`, () =>
        HttpResponse.json({ data: { deletedElementId: 'elm_123' } }),
      ),
    )
    const client = buildClient()
    const result = await client.forms.deleteElement('frm_123', 'elm_123')
    expect(result.deletedElementId).toBe('elm_123')
  })

  it('rejects empty id and empty elementId', async () => {
    const client = buildClient()
    await expect(client.forms.deleteElement('', 'elm_123')).rejects.toThrow(TypeError)
    await expect(client.forms.deleteElement('frm_123', '')).rejects.toThrow(TypeError)
  })

  it('surfaces 404 on missing element', async () => {
    server.use(
      http.delete(`${TEST_BASE_URL}/v1/forms/frm_123/elements/elm_missing`, () =>
        HttpResponse.json({ error: { code: 'not_found', message: 'gone' } }, { status: 404 }),
      ),
    )
    const client = buildClient()
    await expect(client.forms.deleteElement('frm_123', 'elm_missing')).rejects.toBeInstanceOf(
      ThreeCommonNotFoundError,
    )
  })
})

describe('forms.moveElement', () => {
  it('PUTs the position and returns the form', async () => {
    let body: unknown
    server.use(
      http.put(
        `${TEST_BASE_URL}/v1/forms/frm_123/elements/elm_123/position`,
        async ({ request }) => {
          body = await request.json()
          return HttpResponse.json({ data: sampleForm })
        },
      ),
    )
    const client = buildClient()
    const form = await client.forms.moveElement('frm_123', 'elm_123', { position: 2 })
    expect(form.id).toBe('frm_123')
    expect(body).toEqual({ position: 2 })
  })

  it('rejects empty id and empty elementId', async () => {
    const client = buildClient()
    await expect(client.forms.moveElement('', 'elm_123', { position: 1 })).rejects.toThrow(
      TypeError,
    )
    await expect(client.forms.moveElement('frm_123', '', { position: 1 })).rejects.toThrow(
      TypeError,
    )
  })
})

describe('forms.enableOtherOption', () => {
  it('PUTs the other-option body and returns the element', async () => {
    let body: unknown
    server.use(
      http.put(
        `${TEST_BASE_URL}/v1/forms/frm_123/elements/elm_select/other-option`,
        async ({ request }) => {
          body = await request.json()
          return HttpResponse.json({
            data: { ...sampleElement, id: 'elm_select', type: 'Select One or "Other"' },
          })
        },
      ),
    )
    const client = buildClient()
    const element = await client.forms.enableOtherOption('frm_123', 'elm_select', {
      otherPrompt: 'Other (please specify)',
    })
    expect(element.id).toBe('elm_select')
    expect(body).toEqual({ otherPrompt: 'Other (please specify)' })
  })

  it('rejects empty id and empty elementId', async () => {
    const client = buildClient()
    await expect(
      client.forms.enableOtherOption('', 'elm_select', { otherPrompt: 'x' }),
    ).rejects.toThrow(TypeError)
    await expect(
      client.forms.enableOtherOption('frm_123', '', { otherPrompt: 'x' }),
    ).rejects.toThrow(TypeError)
  })
})

describe('forms.disableOtherOption', () => {
  it('DELETEs the other-option and returns the element', async () => {
    server.use(
      http.delete(`${TEST_BASE_URL}/v1/forms/frm_123/elements/elm_select/other-option`, () =>
        HttpResponse.json({ data: { ...sampleElement, id: 'elm_select', type: 'Select One' } }),
      ),
    )
    const client = buildClient()
    const element = await client.forms.disableOtherOption('frm_123', 'elm_select')
    expect(element.id).toBe('elm_select')
  })

  it('rejects empty id and empty elementId', async () => {
    const client = buildClient()
    await expect(client.forms.disableOtherOption('', 'elm_select')).rejects.toThrow(TypeError)
    await expect(client.forms.disableOtherOption('frm_123', '')).rejects.toThrow(TypeError)
  })
})

describe('forms.addLogicRule', () => {
  it('POSTs the rule and returns the source element', async () => {
    let body: unknown
    server.use(
      http.post(
        `${TEST_BASE_URL}/v1/forms/frm_123/elements/elm_select/logic-rules`,
        async ({ request }) => {
          body = await request.json()
          return HttpResponse.json({
            data: { ...sampleElement, id: 'elm_select', type: 'Select One' },
          })
        },
      ),
    )
    const client = buildClient()
    const element = await client.forms.addLogicRule('frm_123', 'elm_select', {
      revealedElementId: 'elm_followup',
      condition: { optionIndices: [0], operator: 'any_of' },
    })
    expect(element.id).toBe('elm_select')
    expect(body).toEqual({
      revealedElementId: 'elm_followup',
      condition: { optionIndices: [0], operator: 'any_of' },
    })
  })

  it('rejects empty id and empty elementId', async () => {
    const client = buildClient()
    await expect(
      client.forms.addLogicRule('', 'elm_select', {
        revealedElementId: 'elm_followup',
        condition: { optionIndices: [0], operator: 'any_of' },
      }),
    ).rejects.toThrow(TypeError)
    await expect(
      client.forms.addLogicRule('frm_123', '', {
        revealedElementId: 'elm_followup',
        condition: { optionIndices: [0], operator: 'any_of' },
      }),
    ).rejects.toThrow(TypeError)
  })
})

describe('forms.removeLogicRule', () => {
  it('DELETEs the rule and returns the source element', async () => {
    server.use(
      http.delete(
        `${TEST_BASE_URL}/v1/forms/frm_123/elements/elm_select/logic-rules/elm_followup`,
        () =>
          HttpResponse.json({ data: { ...sampleElement, id: 'elm_select', type: 'Select One' } }),
      ),
    )
    const client = buildClient()
    const element = await client.forms.removeLogicRule('frm_123', 'elm_select', 'elm_followup')
    expect(element.id).toBe('elm_select')
  })

  it('rejects empty id, empty elementId, and empty targetElementId', async () => {
    const client = buildClient()
    await expect(client.forms.removeLogicRule('', 'elm_select', 'elm_followup')).rejects.toThrow(
      TypeError,
    )
    await expect(client.forms.removeLogicRule('frm_123', '', 'elm_followup')).rejects.toThrow(
      TypeError,
    )
    await expect(client.forms.removeLogicRule('frm_123', 'elm_select', '')).rejects.toThrow(
      TypeError,
    )
  })
})

describe('forms.listAutoPaginate', () => {
  it('walks page -> page+1 until hasMore is false', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/forms`, ({ request }) => {
        const url = new URL(request.url)
        const page = Number(url.searchParams.get('page') ?? '0')
        if (page === 0) {
          return HttpResponse.json({
            data: [
              { id: 'frm_a', name: 'A', numElements: 1, type: 'standalone', status: 'active' },
              { id: 'frm_b', name: 'B', numElements: 1, type: 'standalone', status: 'active' },
            ],
            hasMore: true,
          })
        }
        return HttpResponse.json({
          data: [{ id: 'frm_c', name: 'C', numElements: 1, type: 'standalone', status: 'draft' }],
          hasMore: false,
        })
      }),
    )
    const client = buildClient()
    const ids: string[] = []
    for await (const form of client.forms.listAutoPaginate()) {
      ids.push(form.id)
    }
    expect(ids).toEqual(['frm_a', 'frm_b', 'frm_c'])
  })

  it('honors a non-zero starting page', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/forms`, ({ request }) => {
        const url = new URL(request.url)
        expect(url.searchParams.get('page')).toBe('3')
        return HttpResponse.json({
          data: [
            { id: 'frm_p3', name: 'P3', numElements: 1, type: 'standalone', status: 'active' },
          ],
          hasMore: false,
        })
      }),
    )
    const client = buildClient()
    const ids: string[] = []
    for await (const form of client.forms.listAutoPaginate({ page: 3 })) {
      ids.push(form.id)
    }
    expect(ids).toEqual(['frm_p3'])
  })
})
