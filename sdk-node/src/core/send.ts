import { fromFetchResponse, type HttpClientResponse } from './parse'

import type { HttpMethod } from './retry'

/**
 * One-shot HTTP call: builds an `AbortController` that combines the SDK's
 * timeout with any caller-provided signal, dispatches `fetch`, and returns a
 * normalized response. Pure I/O — does not implement retry.
 *
 * @internal
 */
export async function send(args: {
  readonly fetch: typeof fetch
  readonly url: string
  readonly method: HttpMethod
  readonly headers: Headers
  readonly body: Record<string, unknown> | undefined
  readonly timeoutMs: number
  readonly signal: AbortSignal | undefined
}): Promise<HttpClientResponse> {
  const controller = new AbortController()
  const timeout = setTimeout(() => {
    controller.abort(new Error(`Request timed out after ${String(args.timeoutMs)}ms`))
  }, args.timeoutMs)

  const userSignal = args.signal
  const onUserAbort = (): void => {
    controller.abort(userSignal?.reason)
  }
  if (userSignal !== undefined) {
    if (userSignal.aborted) {
      controller.abort(userSignal.reason)
    } else {
      userSignal.addEventListener('abort', onUserAbort, { once: true })
    }
  }

  try {
    const response = await args.fetch(args.url, {
      method: args.method,
      headers: args.headers,
      body: args.body === undefined ? null : JSON.stringify(args.body),
      signal: controller.signal,
    })
    return await fromFetchResponse(response)
  } finally {
    clearTimeout(timeout)
    if (userSignal !== undefined) {
      userSignal.removeEventListener('abort', onUserAbort)
    }
  }
}
