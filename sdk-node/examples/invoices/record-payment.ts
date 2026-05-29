/**
 * Record a manual payment against an open invoice. The `idempotencyKey`
 * makes the request safe to replay — recording the same payment twice with
 * the same key is a no-op.
 *
 * Run:
 *   npx tsx examples/invoices/record-payment.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const updated = await client.invoices.recordPayment('inv_replace_with_real_id', {
  payment: 50_000, // $500.00 in cents
  // Derive the idempotency key from a stable business event id (e.g. the
  // payment id in your own system) — never the wall clock. A fresh value on
  // each run is a new key, so a retry after a crash records a second payment.
  idempotencyKey: 'pmt-4310',
  note: 'Wire transfer, ref ABCD-1234',
})

console.log(`invoice ${updated.id ?? '?'} now ${updated.status ?? '?'}`)
console.log(`  paid: ${String(updated.amountPaid ?? 0)}, due: ${String(updated.amountDue ?? 0)}`)
