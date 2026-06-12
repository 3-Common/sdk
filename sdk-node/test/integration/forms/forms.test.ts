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
  name: 'Registration',
  ownerId: 'hst_1',
  type: 'standalone' as const,
  status: 'active' as const,
  submitButtonText: 'Submit',
  submitButtonWidth: 'auto' as const,
  rows: [],
  elements: [],
}

const sampleElement = {
  id: 'elm_1',
  prompt: 'What is your name?',
  type: 'Text' as const,
  required: true,
}

const sampleSummary = {
  id: 'frm_a',
  name: 'Newsletter Signup',
  numElements: 3,
  type: 'standalone' as const,
  status: 'active' as const,
}

describe('forms.list', () => {
  it('returns data + hasMore', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/forms`, () =>
        HttpResponse.json({ data: [sampleSummary], hasMore: false }),
      ),
    )
    const client = buildClient()
    const result = await client.forms.list({ type: 'standalone', pageSize: 10 })
    expect(result.hasMore).toBe(false)
    expect(result.data).toHaveLength(1)
    expect(result.data[0]?.id).toBe('frm_a')
  })

  it('forwards query params', async () => {
    let url = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/forms`, ({ request }) => {
        url = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.forms.list({ type: 'order', page: 2, pageSize: 25 })
    expect(url).toContain('type=order')
    expect(url).toContain('page=2')
    expect(url).toContain('pageSize=25')
  })

  it('defaults to no params', async () => {
    let url = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/forms`, ({ request }) => {
        url = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.forms.list()
    expect(new URL(url).search).toBe('')
  })

  it('skips undefined params and drops non-primitive values', async () => {
    let url = ''
    server.use(
      http.get(`${TEST_BASE_URL}/v1/forms`, ({ request }) => {
        url = request.url
        return HttpResponse.json({ data: [], hasMore: false })
      }),
    )
    const client = buildClient()
    await client.forms.list({
      type: 'standalone',
      pageSize: undefined,
      forbidden: ['x', 'y'],
    } as unknown as Parameters<typeof client.forms.list>[0])
    expect(url).toContain('type=standalone')
    expect(url).not.toContain('pageSize')
    expect(url).not.toContain('forbidden')
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
    expect(form.name).toBe('Registration')
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
    const form = await client.forms.create({ name: 'Registration', type: 'standalone' })
    expect(form.id).toBe('frm_123')
    expect(body).toEqual({ name: 'Registration', type: 'standalone' })
  })

  it('surfaces 400 validation errors', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/forms`, () =>
        HttpResponse.json(
          { error: { code: 'validation_failed', message: 'bad type' } },
          { status: 400 },
        ),
      ),
    )
    const client = buildClient()
    await expect(
      client.forms.create({
        name: 'Bad',
        type: 'invalid',
      } as unknown as Parameters<typeof client.forms.create>[0]),
    ).rejects.toBeInstanceOf(ThreeCommonValidationError)
  })
})

describe('forms.update', () => {
  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.forms.update('', { name: 'x' })).rejects.toThrow(TypeError)
  })

  it('PATCHes and returns the unwrapped form', async () => {
    let body: unknown
    server.use(
      http.patch(`${TEST_BASE_URL}/v1/forms/frm_123`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: { ...sampleForm, name: 'Updated Registration' } })
      }),
    )
    const client = buildClient()
    const form = await client.forms.update('frm_123', { name: 'Updated Registration' })
    expect(form.name).toBe('Updated Registration')
    expect(body).toEqual({ name: 'Updated Registration' })
  })
})

describe('forms.duplicate', () => {
  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(client.forms.duplicate('', { name: 'x' })).rejects.toThrow(TypeError)
  })

  it('POSTs to /duplicate and returns the new form', async () => {
    server.use(
      http.post(`${TEST_BASE_URL}/v1/forms/frm_123/duplicate`, () =>
        HttpResponse.json({ data: { ...sampleForm, id: 'frm_copy', name: 'Registration (Copy)' } }),
      ),
    )
    const client = buildClient()
    const copy = await client.forms.duplicate('frm_123', { name: 'Registration (Copy)' })
    expect(copy.id).toBe('frm_copy')
  })
})

describe('forms.addElement', () => {
  it('rejects empty id', async () => {
    const client = buildClient()
    await expect(
      client.forms.addElement('', { prompt: 'q', type: 'Text', required: true }),
    ).rejects.toThrow(TypeError)
  })

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
    expect(element.id).toBe('elm_1')
    expect(body).toEqual({ prompt: 'What is your name?', type: 'Text', required: true })
  })
})

describe('forms.updateElement', () => {
  it('rejects empty id and empty elementId', async () => {
    const client = buildClient()
    await expect(client.forms.updateElement('', 'elm_1', {})).rejects.toThrow(TypeError)
    await expect(client.forms.updateElement('frm_123', '', {})).rejects.toThrow(TypeError)
  })

  it('PATCHes the element and returns it', async () => {
    server.use(
      http.patch(`${TEST_BASE_URL}/v1/forms/frm_123/elements/elm_1`, () =>
        HttpResponse.json({ data: { ...sampleElement, prompt: 'What is your full name?' } }),
      ),
    )
    const client = buildClient()
    const element = await client.forms.updateElement('frm_123', 'elm_1', {
      prompt: 'What is your full name?',
      required: false,
    })
    expect(element.id).toBe('elm_1')
  })

  it('throws ThreeCommonNotFoundError when the element is missing', async () => {
    server.use(
      http.patch(`${TEST_BASE_URL}/v1/forms/frm_123/elements/elm_missing`, () =>
        HttpResponse.json({ error: { code: 'not_found', message: 'gone' } }, { status: 404 }),
      ),
    )
    const client = buildClient()
    await expect(
      client.forms.updateElement('frm_123', 'elm_missing', { prompt: 'x' }),
    ).rejects.toBeInstanceOf(ThreeCommonNotFoundError)
  })
})

describe('forms.deleteElement', () => {
  it('rejects empty id and empty elementId', async () => {
    const client = buildClient()
    await expect(client.forms.deleteElement('', 'elm_1')).rejects.toThrow(TypeError)
    await expect(client.forms.deleteElement('frm_123', '')).rejects.toThrow(TypeError)
  })

  it('DELETEs and returns the deleted element id', async () => {
    server.use(
      http.delete(`${TEST_BASE_URL}/v1/forms/frm_123/elements/elm_1`, () =>
        HttpResponse.json({ data: { deletedElementId: 'elm_1' } }),
      ),
    )
    const client = buildClient()
    const result = await client.forms.deleteElement('frm_123', 'elm_1')
    expect(result.deletedElementId).toBe('elm_1')
  })
})

describe('forms.moveElement', () => {
  it('rejects empty id and empty elementId', async () => {
    const client = buildClient()
    await expect(client.forms.moveElement('', 'elm_1', { position: 1 })).rejects.toThrow(TypeError)
    await expect(client.forms.moveElement('frm_123', '', { position: 1 })).rejects.toThrow(
      TypeError,
    )
  })

  it('PUTs to /position and returns the updated form', async () => {
    let body: unknown
    server.use(
      http.put(`${TEST_BASE_URL}/v1/forms/frm_123/elements/elm_1/position`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({ data: sampleForm })
      }),
    )
    const client = buildClient()
    const form = await client.forms.moveElement('frm_123', 'elm_1', { position: 2 })
    expect(form.id).toBe('frm_123')
    expect(body).toEqual({ position: 2 })
  })
})

describe('forms.enableOtherOption / disableOtherOption', () => {
  it('enable rejects empty id and empty elementId', async () => {
    const client = buildClient()
    await expect(
      client.forms.enableOtherOption('', 'elm_1', { otherPrompt: 'Other' }),
    ).rejects.toThrow(TypeError)
    await expect(
      client.forms.enableOtherOption('frm_123', '', { otherPrompt: 'Other' }),
    ).rejects.toThrow(TypeError)
  })

  it('PUTs other-option to enable and returns the element', async () => {
    let body: unknown
    server.use(
      http.put(
        `${TEST_BASE_URL}/v1/forms/frm_123/elements/elm_1/other-option`,
        async ({ request }) => {
          body = await request.json()
          return HttpResponse.json({
            data: { ...sampleElement, type: 'Select One or "Other"', otherPrompt: 'Other' },
          })
        },
      ),
    )
    const client = buildClient()
    const element = await client.forms.enableOtherOption('frm_123', 'elm_1', {
      otherPrompt: 'Other',
    })
    expect(element.id).toBe('elm_1')
    expect(body).toEqual({ otherPrompt: 'Other' })
  })

  it('disable rejects empty id and empty elementId', async () => {
    const client = buildClient()
    await expect(client.forms.disableOtherOption('', 'elm_1')).rejects.toThrow(TypeError)
    await expect(client.forms.disableOtherOption('frm_123', '')).rejects.toThrow(TypeError)
  })

  it('DELETEs other-option to disable and returns the element', async () => {
    server.use(
      http.delete(`${TEST_BASE_URL}/v1/forms/frm_123/elements/elm_1/other-option`, () =>
        HttpResponse.json({ data: { ...sampleElement, type: 'Select One' } }),
      ),
    )
    const client = buildClient()
    const element = await client.forms.disableOtherOption('frm_123', 'elm_1')
    expect(element.id).toBe('elm_1')
  })
})

describe('forms.addLogicRule / removeLogicRule', () => {
  it('add rejects empty id and empty elementId', async () => {
    const client = buildClient()
    await expect(
      client.forms.addLogicRule('', 'elm_1', {
        revealedElementId: 'elm_2',
        condition: { optionIndices: [0], operator: 'any_of' },
      }),
    ).rejects.toThrow(TypeError)
    await expect(
      client.forms.addLogicRule('frm_123', '', {
        revealedElementId: 'elm_2',
        condition: { optionIndices: [0], operator: 'any_of' },
      }),
    ).rejects.toThrow(TypeError)
  })

  it('POSTs a logic rule and returns the source element', async () => {
    let body: unknown
    server.use(
      http.post(
        `${TEST_BASE_URL}/v1/forms/frm_123/elements/elm_1/logic-rules`,
        async ({ request }) => {
          body = await request.json()
          return HttpResponse.json({ data: { ...sampleElement, type: 'Select One' } })
        },
      ),
    )
    const client = buildClient()
    const element = await client.forms.addLogicRule('frm_123', 'elm_1', {
      revealedElementId: 'elm_2',
      condition: { optionIndices: [0], operator: 'any_of' },
    })
    expect(element.id).toBe('elm_1')
    expect(body).toEqual({
      revealedElementId: 'elm_2',
      condition: { optionIndices: [0], operator: 'any_of' },
    })
  })

  it('remove rejects empty id, elementId, and targetElementId', async () => {
    const client = buildClient()
    await expect(client.forms.removeLogicRule('', 'elm_1', 'elm_2')).rejects.toThrow(TypeError)
    await expect(client.forms.removeLogicRule('frm_123', '', 'elm_2')).rejects.toThrow(TypeError)
    await expect(client.forms.removeLogicRule('frm_123', 'elm_1', '')).rejects.toThrow(TypeError)
  })

  it('DELETEs a logic rule and returns the source element', async () => {
    server.use(
      http.delete(`${TEST_BASE_URL}/v1/forms/frm_123/elements/elm_1/logic-rules/elm_2`, () =>
        HttpResponse.json({ data: { ...sampleElement, type: 'Select One' } }),
      ),
    )
    const client = buildClient()
    const element = await client.forms.removeLogicRule('frm_123', 'elm_1', 'elm_2')
    expect(element.id).toBe('elm_1')
  })
})

describe('forms.listAutoPaginate', () => {
  it('iterates across pages', async () => {
    server.use(
      http.get(`${TEST_BASE_URL}/v1/forms`, ({ request }) => {
        const page = Number(new URL(request.url).searchParams.get('page') ?? '0')
        if (page === 0) {
          return HttpResponse.json({
            data: [
              { ...sampleSummary, id: 'frm_1' },
              { ...sampleSummary, id: 'frm_2' },
            ],
            hasMore: true,
          })
        }
        return HttpResponse.json({ data: [{ ...sampleSummary, id: 'frm_3' }], hasMore: false })
      }),
    )
    const client = buildClient()
    const ids: string[] = []
    for await (const form of client.forms.listAutoPaginate({ type: 'standalone' })) {
      ids.push(form.id)
    }
    expect(ids).toEqual(['frm_1', 'frm_2', 'frm_3'])
  })
})
