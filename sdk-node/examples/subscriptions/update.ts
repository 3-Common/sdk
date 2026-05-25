/**
 * Apply a mid-cycle upgrade. The SDK returns the updated subscription, a
 * proration summary, and (when the rate difference is positive) a slim
 * reference to the proration invoice.
 *
 * Run:
 *   npx tsx examples/subscriptions/update.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const { subscription, invoice, proration } = await client.subscriptions.update(
  'sub_replace_with_real_id',
  {
    priceId: 'price_upgrade_replace_with_real_id',
    quantity: 2,
  },
)

console.log(
  `updated ${subscription.id ?? '?'} → ${subscription.priceId ?? '?'} × ${String(subscription.quantity ?? 0)}`,
)
console.log(
  `proration: ${String(proration.netAmountMinor)} minor units (${String(proration.daysRemaining)}/${String(proration.daysInCycle)} days)`,
)
if (invoice !== undefined) {
  console.log(
    `proration invoice ${invoice.id} [${invoice.status}] — total ${String(invoice.total)} ${invoice.currency}`,
  )
} else {
  console.log('downgrade or no-op — no proration invoice issued')
}
