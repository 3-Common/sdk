/**
 * Pre-release smoke test against the live API.
 *
 * Runs the happy path + the common error paths across the events, invoices,
 * and subscriptions resources. Used by `.github/workflows/live-smoke.yml`
 * (maintainer-only).
 *
 * Required env:
 *   THREECOMMON_API_KEY        — a real API key
 *
 * Optional env:
 *   THREECOMMON_BASE_URL       — defaults to https://api.3common.com
 *   SMOKE_EVENT_ID             — an event ID owned by the API-key host; if set,
 *                                exercises the events.retrieve happy path
 *   SMOKE_INVOICE_ID           — an invoice ID owned by the API-key host; if set,
 *                                exercises the invoices.retrieve happy path
 *   SMOKE_SUBSCRIPTION_ID      — a subscription ID owned by the API-key host;
 *                                if set, exercises subscriptions.retrieve and
 *                                subscriptions.previewUpcomingInvoice
 */

import process from 'node:process'

import {
  ThreeCommon,
  ThreeCommonAuthError,
  ThreeCommonError,
  ThreeCommonNotFoundError,
} from '@/index'

// Syntactically valid 24-hex ObjectId that will not match any real record.
// The API rejects non-ObjectId strings with a 400 before reaching the
// not-found path, so this must look well-formed to test 404s.
const MISSING_OBJECT_ID = '000000000000000000000000'

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
  const knownInvoiceId = process.env['SMOKE_INVOICE_ID']
  const knownSubscriptionId = process.env['SMOKE_SUBSCRIPTION_ID']

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

  // 4. 404 path on events — well-formed ID that does not exist.
  try {
    await client.events.retrieve(MISSING_OBJECT_ID)
    results.push({
      check: 'events 404 path',
      status: 'fail',
      detail: 'expected ThreeCommonNotFoundError but call succeeded',
    })
  } catch (err) {
    if (err instanceof ThreeCommonNotFoundError) {
      results.push({
        check: 'events 404 path',
        status: 'pass',
        detail: `code=${err.code}, requestId=${err.requestId ?? '?'}`,
      })
    } else {
      results.push({
        check: 'events 404 path',
        status: 'fail',
        detail: `unexpected error: ${errMsg(err)}`,
      })
    }
  }

  // 5. List invoices.
  try {
    const result = await client.invoices.list({ pageSize: 1 })
    results.push({
      check: 'invoices.list',
      status: 'pass',
      detail: `data.length=${String(result.data.length)}, hasMore=${String(result.hasMore)}`,
    })
  } catch (err) {
    results.push({ check: 'invoices.list', status: 'fail', detail: errMsg(err) })
  }

  // 6. Retrieve a known invoice (if configured).
  if (knownInvoiceId !== undefined && knownInvoiceId.length > 0) {
    try {
      const invoice = await client.invoices.retrieve(knownInvoiceId)
      results.push({
        check: 'invoices.retrieve',
        status: 'pass',
        detail: `id=${invoice.id ?? '?'}`,
      })
    } catch (err) {
      results.push({ check: 'invoices.retrieve', status: 'fail', detail: errMsg(err) })
    }
  } else {
    results.push({
      check: 'invoices.retrieve',
      status: 'skip',
      detail: 'SMOKE_INVOICE_ID not set',
    })
  }

  // 7. 404 path on invoices.
  try {
    await client.invoices.retrieve(MISSING_OBJECT_ID)
    results.push({
      check: 'invoices 404 path',
      status: 'fail',
      detail: 'expected ThreeCommonNotFoundError but call succeeded',
    })
  } catch (err) {
    if (err instanceof ThreeCommonNotFoundError) {
      results.push({
        check: 'invoices 404 path',
        status: 'pass',
        detail: `code=${err.code}, requestId=${err.requestId ?? '?'}`,
      })
    } else {
      results.push({
        check: 'invoices 404 path',
        status: 'fail',
        detail: `unexpected error: ${errMsg(err)}`,
      })
    }
  }

  // 7b. New invoice write methods — only their not-found paths are smoke-tested.
  // The happy paths move real money (auto_charge, refundPayment) or delete a
  // record (deleteDraft), so running them against the live host on every
  // maintainer run isn't safe. A well-formed-but-missing id 404s when the
  // handler loads the invoice, before any side effect, so these exercise the
  // wire path + error mapping without touching real data.
  try {
    await client.invoices.autoCharge(MISSING_OBJECT_ID)
    results.push({
      check: 'invoices.autoCharge 404 path',
      status: 'fail',
      detail: 'expected ThreeCommonNotFoundError but call succeeded',
    })
  } catch (err) {
    results.push(notFoundCheck('invoices.autoCharge 404 path', err))
  }

  try {
    await client.invoices.refundPayment(MISSING_OBJECT_ID, MISSING_OBJECT_ID, { amount: 1 })
    results.push({
      check: 'invoices.refundPayment 404 path',
      status: 'fail',
      detail: 'expected ThreeCommonNotFoundError but call succeeded',
    })
  } catch (err) {
    results.push(notFoundCheck('invoices.refundPayment 404 path', err))
  }

  try {
    await client.invoices.deleteDraft(MISSING_OBJECT_ID)
    results.push({
      check: 'invoices.deleteDraft 404 path',
      status: 'fail',
      detail: 'expected ThreeCommonNotFoundError but call succeeded',
    })
  } catch (err) {
    results.push(notFoundCheck('invoices.deleteDraft 404 path', err))
  }

  // 8. List subscriptions.
  try {
    const result = await client.subscriptions.list({ pageSize: 1 })
    results.push({
      check: 'subscriptions.list',
      status: 'pass',
      detail: `data.length=${String(result.data.length)}, hasMore=${String(result.hasMore)}`,
    })
  } catch (err) {
    results.push({ check: 'subscriptions.list', status: 'fail', detail: errMsg(err) })
  }

  // 9. Retrieve a known subscription (if configured).
  if (knownSubscriptionId !== undefined && knownSubscriptionId.length > 0) {
    try {
      const sub = await client.subscriptions.retrieve(knownSubscriptionId)
      results.push({
        check: 'subscriptions.retrieve',
        status: 'pass',
        detail: `id=${sub.id ?? '?'}, status=${sub.status ?? '?'}`,
      })
    } catch (err) {
      results.push({ check: 'subscriptions.retrieve', status: 'fail', detail: errMsg(err) })
    }

    // 10. Preview upcoming invoice on the known subscription.
    try {
      const preview = await client.subscriptions.previewUpcomingInvoice(knownSubscriptionId)
      results.push({
        check: 'subscriptions.previewUpcomingInvoice',
        status: 'pass',
        detail: preview === null ? 'null (cancel-at-period-end)' : `total=${String(preview.total)}`,
      })
    } catch (err) {
      results.push({
        check: 'subscriptions.previewUpcomingInvoice',
        status: 'fail',
        detail: errMsg(err),
      })
    }
  } else {
    results.push({
      check: 'subscriptions.retrieve',
      status: 'skip',
      detail: 'SMOKE_SUBSCRIPTION_ID not set',
    })
    results.push({
      check: 'subscriptions.previewUpcomingInvoice',
      status: 'skip',
      detail: 'SMOKE_SUBSCRIPTION_ID not set',
    })
  }

  // 11. 404 path on subscriptions.
  try {
    await client.subscriptions.retrieve(MISSING_OBJECT_ID)
    results.push({
      check: 'subscriptions 404 path',
      status: 'fail',
      detail: 'expected ThreeCommonNotFoundError but call succeeded',
    })
  } catch (err) {
    if (err instanceof ThreeCommonNotFoundError) {
      results.push({
        check: 'subscriptions 404 path',
        status: 'pass',
        detail: `code=${err.code}, requestId=${err.requestId ?? '?'}`,
      })
    } else {
      results.push({
        check: 'subscriptions 404 path',
        status: 'fail',
        detail: `unexpected error: ${errMsg(err)}`,
      })
    }
  }

  // 11b. New subscription write methods — only their not-found paths are
  // smoke-tested. `compNextCycle` stages a 100%-off next renewal and
  // `uncompNextCycle` clears it; both mutate real subscription state, so
  // running the happy paths against the live host on every maintainer run
  // isn't safe. A well-formed-but-missing id 404s when the handler loads the
  // subscription, before any state change, so these exercise the wire path +
  // error mapping without touching real data.
  try {
    await client.subscriptions.compNextCycle(MISSING_OBJECT_ID)
    results.push({
      check: 'subscriptions.compNextCycle 404 path',
      status: 'fail',
      detail: 'expected ThreeCommonNotFoundError but call succeeded',
    })
  } catch (err) {
    results.push(notFoundCheck('subscriptions.compNextCycle 404 path', err))
  }

  try {
    await client.subscriptions.uncompNextCycle(MISSING_OBJECT_ID)
    results.push({
      check: 'subscriptions.uncompNextCycle 404 path',
      status: 'fail',
      detail: 'expected ThreeCommonNotFoundError but call succeeded',
    })
  } catch (err) {
    results.push(notFoundCheck('subscriptions.uncompNextCycle 404 path', err))
  }

  // 12. 401 path — wrong API key.
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

// Shared assertion for the new invoice methods' not-found checks: a pass when
// the call rejected with a 404, a fail otherwise.
function notFoundCheck(check: string, err: unknown): SmokeResult {
  if (err instanceof ThreeCommonNotFoundError) {
    return { check, status: 'pass', detail: `code=${err.code}, requestId=${err.requestId ?? '?'}` }
  }
  return { check, status: 'fail', detail: `unexpected error: ${errMsg(err)}` }
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
