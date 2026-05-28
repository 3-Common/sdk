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
    idempotencyKey: `rfnd-${new Date().toISOString()}`,
  },
)

console.log(`invoice ${refunded.id ?? '?'} now ${refunded.status ?? '?'}`)
console.log(`  paid: ${String(refunded.amountPaid ?? 0)}, due: ${String(refunded.amountDue ?? 0)}`)
