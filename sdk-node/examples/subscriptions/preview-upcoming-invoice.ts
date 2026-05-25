/**
 * Preview the invoice the next renewal will generate (Stripe-style
 * `invoice.upcoming`). Returns `null` when the subscription is set to
 * cancel at period end.
 *
 * Run:
 *   npx tsx examples/subscriptions/preview-upcoming-invoice.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const preview = await client.subscriptions.previewUpcomingInvoice('sub_replace_with_real_id')

if (preview === null) {
  console.log('subscription is set to cancel at period end — no upcoming invoice')
} else {
  console.log(`next invoice — ${String(preview.total)} ${preview.currency}`)
  console.log(`  period ${preview.periodStart} → ${preview.periodEnd}`)
  for (const line of preview.lineItems) {
    console.log(`  • ${line.description} — ${String(line.quantity)} × ${String(line.unitAmount)}`)
  }
}
