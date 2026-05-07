/**
 * Response parsing helpers — pure functions over a typed response wrapper.
 *
 * @internal
 */

/**
 * Typed wrapper around a `fetch` `Response`. Header keys are normalized to
 * lowercase; the body is read once as text and cached.
 *
 * @internal
 */
export interface HttpClientResponse {
  readonly status: number
  readonly headers: ReadonlyMap<string, string>
  readonly requestId: string | undefined
  readonly bodyText: string
}

/**
 * Build an {@link HttpClientResponse} from a standard `Response`.
 *
 * @internal
 */
export async function fromFetchResponse(response: Response): Promise<HttpClientResponse> {
  const headers = new Map<string, string>()
  response.headers.forEach((value, key) => {
    headers.set(key.toLowerCase(), value)
  })

  const bodyText = await response.text()
  const requestId = headers.get('x-request-id')

  return { status: response.status, headers, requestId, bodyText }
}

/**
 * Parse a 2xx response body. Empty or non-JSON bodies resolve to `undefined`.
 *
 * @internal
 */
export function parseSuccessBody(response: HttpClientResponse): unknown {
  if (response.bodyText.length === 0) return undefined
  try {
    return JSON.parse(response.bodyText)
  } catch {
    return undefined
  }
}

/**
 * Best-effort JSON parse for error response bodies.
 *
 * @internal
 */
export function tryParseJson(text: string): unknown {
  if (text.length === 0) return undefined
  try {
    return JSON.parse(text)
  } catch {
    return undefined
  }
}

/**
 * Parse a `Retry-After` header value into seconds.
 *
 * Accepts either a delta-seconds integer or an HTTP-date. Returns `undefined`
 * if the header is missing or unparseable; returns `0` for past HTTP-dates.
 *
 * @internal
 */
export function parseRetryAfter(header: string | undefined): number | undefined {
  if (header === undefined || header.length === 0) return undefined

  const seconds = Number(header)
  if (Number.isFinite(seconds) && seconds >= 0) return seconds

  const date = Date.parse(header)
  if (Number.isNaN(date)) return undefined
  const delta = (date - Date.now()) / 1000
  return delta > 0 ? delta : 0
}
