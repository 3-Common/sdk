/**
 * Pre-release smoke test against the live API.
 *
 * Runs ≤ 10 calls and verifies the happy path + the four common error paths.
 * Used by `.github/workflows/live-smoke.yml` (maintainer-only).
 *
 * Required env:
 *   THREECOMMON_API_KEY   — a real API key
 *
 * Optional env:
 *   THREECOMMON_BASE_URL  — defaults to https://api.3common.com
 *   SMOKE_EVENT_ID        — an event ID known to belong to the API-key host;
 *                           required for retrieve / 403 / 422 checks
 */

import process from 'node:process'

import {
  ThreeCommon,
  ThreeCommonAuthError,
  ThreeCommonError,
  ThreeCommonNotFoundError,
} from '@/index'

interface SmokeResult {
  readonly check: string
  readonly status: 'pass' | 'fail' | 'skip'
  readonly detail?: string
}

async function run(): Promise<SmokeResult[]> {
  const apiKey = process.env['THREECOMMON_API_KEY']
  if (apiKey === undefined || apiKey.length === 0) {
    throw new Error('THREECOMMON_API_KEY env var is required for live-smoke runs')
  }

  const baseUrl = process.env['THREECOMMON_BASE_URL'] ?? 'https://api.3common.com'
  const knownEventId = process.env['SMOKE_EVENT_ID']

  const results: SmokeResult[] = []
  const client = new ThreeCommon({ apiKey, baseUrl, telemetry: false })

  // 1. List events.
  try {
    const result = await client.events.list({ pageSize: 1 })
    results.push({
      check: 'events.list',
      status: 'pass',
      detail: `data.length=${String(result.data.length)}, hasMore=${String(result.hasMore)}`,
    })
  } catch (err) {
    results.push({ check: 'events.list', status: 'fail', detail: errMsg(err) })
  }

  // 2. Auto-paginate (one round of next()).
  try {
    const iter = client.events.listAutoPaginate({ pageSize: 1 })
    const first = await iter.next()
    results.push({
      check: 'events.listAutoPaginate',
      status: 'pass',
      detail: `done=${String(first.done)}`,
    })
  } catch (err) {
    results.push({ check: 'events.listAutoPaginate', status: 'fail', detail: errMsg(err) })
  }

  // 3. Retrieve a known event (if configured).
  if (knownEventId !== undefined && knownEventId.length > 0) {
    try {
      const event = await client.events.retrieve(knownEventId)
      results.push({
        check: 'events.retrieve',
        status: 'pass',
        detail: `id=${event.id ?? '?'}`,
      })
    } catch (err) {
      results.push({ check: 'events.retrieve', status: 'fail', detail: errMsg(err) })
    }
  } else {
    results.push({
      check: 'events.retrieve',
      status: 'skip',
      detail: 'SMOKE_EVENT_ID not set',
    })
  }

  // 4. 404 path — random ID that should not exist.
  try {
    await client.events.retrieve('000000000000000000000000')
    results.push({
      check: '404 path',
      status: 'fail',
      detail: 'expected ThreeCommonNotFoundError but call succeeded',
    })
  } catch (err) {
    if (err instanceof ThreeCommonNotFoundError) {
      results.push({
        check: '404 path',
        status: 'pass',
        detail: `code=${err.code}, requestId=${err.requestId ?? '?'}`,
      })
    } else {
      results.push({
        check: '404 path',
        status: 'fail',
        detail: `unexpected error: ${errMsg(err)}`,
      })
    }
  }

  // 5. 401 path — wrong API key.
  try {
    const badClient = new ThreeCommon({
      apiKey: '3co_smoke_test_invalid_key', // gitleaks:allow — deliberate fake to test the 401 path
      baseUrl,
      telemetry: false,
      maxRetries: 0,
    })
    await badClient.events.list({ pageSize: 1 })
    results.push({
      check: '401 path',
      status: 'fail',
      detail: 'expected ThreeCommonAuthError but call succeeded',
    })
  } catch (err) {
    if (err instanceof ThreeCommonAuthError) {
      results.push({ check: '401 path', status: 'pass', detail: `code=${err.code}` })
    } else {
      results.push({
        check: '401 path',
        status: 'fail',
        detail: `unexpected error: ${errMsg(err)}`,
      })
    }
  }

  return results
}

function errMsg(err: unknown): string {
  if (err instanceof ThreeCommonError) return err.toString()
  if (err instanceof Error) return err.message
  return String(err)
}

const results = await run()

let failed = 0
for (const r of results) {
  const icon = r.status === 'pass' ? '✓' : r.status === 'skip' ? '○' : '✗'
  process.stdout.write(`${icon} ${r.check}${r.detail !== undefined ? ` — ${r.detail}` : ''}\n`)
  if (r.status === 'fail') failed += 1
}

if (failed > 0) {
  process.stderr.write(`\n${String(failed)} check(s) failed.\n`)
  process.exit(1)
}
