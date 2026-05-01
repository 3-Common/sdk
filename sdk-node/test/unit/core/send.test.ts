import { describe, expect, it } from 'vitest'

import { send } from '@/core/send'

describe('send', () => {
  const url = 'https://api.test.example.com/v1/events'
  const headers = new Headers({ Authorization: 'Bearer 3co_test' })

  function makeFetch(
    handler: (init: RequestInit | undefined) => Response | Promise<Response>,
  ): typeof fetch {
    return (_input, init) => Promise.resolve(handler(init))
  }

  it('returns a normalized response on a successful fetch', async () => {
    const fetchImpl = makeFetch(
      () => new Response('{"data":[]}', { status: 200, headers: { 'X-Request-ID': 'req-x' } }),
    )
    const response = await send({
      fetch: fetchImpl,
      url,
      method: 'GET',
      headers,
      body: undefined,
      timeoutMs: 5_000,
      signal: undefined,
    })
    expect(response.status).toBe(200)
    expect(response.requestId).toBe('req-x')
    expect(response.bodyText).toBe('{"data":[]}')
  })

  it('serializes body to JSON when supplied', async () => {
    let captured: unknown
    const fetchImpl = makeFetch((init) => {
      captured = init?.body
      return new Response('{}', { status: 200 })
    })
    await send({
      fetch: fetchImpl,
      url,
      method: 'PATCH',
      headers,
      body: { name: 'x' },
      timeoutMs: 5_000,
      signal: undefined,
    })
    expect(captured).toBe('{"name":"x"}')
  })

  it('aborts immediately when the caller-provided signal is already aborted', async () => {
    const controller = new AbortController()
    controller.abort()

    const fetchImpl: typeof fetch = (_input, init) => {
      // The internal AbortController should already be aborted at this point.
      if ((init?.signal as AbortSignal | undefined)?.aborted === true) {
        return Promise.reject(new Error('aborted'))
      }
      return Promise.resolve(new Response('{}', { status: 200 }))
    }

    await expect(
      send({
        fetch: fetchImpl,
        url,
        method: 'GET',
        headers,
        body: undefined,
        timeoutMs: 5_000,
        signal: controller.signal,
      }),
    ).rejects.toThrow()
  })

  it('aborts on caller-provided signal aborting after the request starts', async () => {
    const controller = new AbortController()
    const fetchImpl: typeof fetch = (_input, init) => {
      return new Promise((_resolve, reject) => {
        const sig = init?.signal as AbortSignal | undefined
        sig?.addEventListener('abort', () => {
          reject(new Error('aborted'))
        })
      })
    }

    setTimeout(() => {
      controller.abort()
    }, 5)

    await expect(
      send({
        fetch: fetchImpl,
        url,
        method: 'GET',
        headers,
        body: undefined,
        timeoutMs: 5_000,
        signal: controller.signal,
      }),
    ).rejects.toThrow()
  })

  it('aborts after the configured timeout', async () => {
    const fetchImpl: typeof fetch = (_input, init) => {
      return new Promise((_resolve, reject) => {
        const sig = init?.signal as AbortSignal | undefined
        sig?.addEventListener('abort', () => {
          reject(new Error('aborted by timeout'))
        })
      })
    }

    await expect(
      send({
        fetch: fetchImpl,
        url,
        method: 'GET',
        headers,
        body: undefined,
        timeoutMs: 5,
        signal: undefined,
      }),
    ).rejects.toThrow()
  })
})
