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
  idempotencyKey: `pmt-${new Date().toISOString()}`,
  note: 'Wire transfer, ref ABCD-1234',
})

console.log(`invoice ${updated.id ?? '?'} now ${updated.status ?? '?'}`)
console.log(`  paid: ${String(updated.amountPaid ?? 0)}, due: ${String(updated.amountDue ?? 0)}`)
