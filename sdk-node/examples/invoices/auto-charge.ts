/**
 * Off-session auto-charge an open invoice against the customer's saved card.
 * A decline is not an error — the call resolves with `outcome: 'failed'` and a
 * `failureCode`, leaving the invoice in `payment_failed`. Only network /
 * processor (5xx) errors throw.
 *
 * Run:
 *   npx tsx examples/invoices/auto-charge.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const { invoice, outcome, failureCode } = await client.invoices.autoCharge(
  'inv_replace_with_real_id',
)

if (outcome === 'paid') {
  console.log(`invoice ${invoice.id ?? '?'} charged, now ${invoice.status ?? '?'}`)
} else {
  console.warn(`charge failed (${failureCode ?? 'unknown'}); invoice is ${invoice.status ?? '?'}`)
}
