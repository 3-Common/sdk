/**
 * Refund all or part of a recorded payment on a paid invoice. The
 * `idempotencyKey` makes the request safe to replay — refunding twice with the
 * same key returns the existing refund instead of issuing a second one.
 *
 * Run:
 *   npx tsx examples/invoices/refund-payment.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const refunded = await client.invoices.refundPayment(
  'inv_replace_with_real_id',
  'pay_replace_with_real_id',
  {
    amount: 25_000, // $250.00 in cents; capped at the payment's refundable balance
    reason: 'requested_by_customer',
    // Derive the idempotency key from a stable business event id (e.g. the
    // refund-request id in your own system) — never the wall clock. A fresh
    // value on each run is a new key, so a retry after a crash refunds twice.
    idempotencyKey: 'rfnd-8842',
  },
)

console.log(`invoice ${refunded.id ?? '?'} now ${refunded.status ?? '?'}`)
console.log(`  paid: ${String(refunded.amountPaid ?? 0)}, due: ${String(refunded.amountDue ?? 0)}`)
