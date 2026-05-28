/**
 * Create a new subscription with a 14-day trial. The subscription starts in
 * `trialing` and transitions to `active` once the first payment succeeds.
 *
 * Run:
 *   npx tsx examples/subscriptions/create.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const sub = await client.subscriptions.create({
  contactId: 'cnt_replace_with_real_id',
  priceId: 'price_replace_with_real_id',
  quantity: 1,
  trialDays: 14,
  autoCharge: true,
  notes: 'Pro plan — annual billing',
  metadata: { source: 'website-checkout' },
})

console.log(`created ${sub.id ?? '?'} [${sub.status ?? '?'}]`)
console.log(`  trial ends   ${sub.trialEnd ?? 'n/a'}`)
console.log(`  first bill   ${sub.currentPeriodEnd ?? '?'}`)
