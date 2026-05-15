/**
 * Inject a custom `fetch` implementation. Useful for proxies, custom DNS,
 * request logging, or testing.
 *
 * Run:
 *   npx tsx examples/events/custom-fetch.ts
 */

import { ThreeCommon } from '@3common/sdk'

const loggingFetch: typeof fetch = async (input, init) => {
  const url = typeof input === 'string' || input instanceof URL ? String(input) : input.url
  const start = performance.now()
  const response = await globalThis.fetch(input, init)
  const ms = (performance.now() - start).toFixed(1)
  console.log(`[fetch] ${init?.method ?? 'GET'} ${url} → ${String(response.status)} (${ms}ms)`)
  return response
}

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
  fetch: loggingFetch,
})

const events = await client.events.list({ pageSize: 5 })

console.log(events)
