/**
 * Retrieve a single subscription by ID.
 *
 * Run:
 *   npx tsx examples/subscriptions/retrieve.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const sub = await client.subscriptions.retrieve('sub_replace_with_real_id')

console.log(`subscription ${sub.id ?? '?'} [${sub.status ?? '?'}]`)
console.log(`  price          ${sub.priceId ?? '?'} × ${String(sub.quantity ?? 0)}`)
console.log(`  current period ${sub.currentPeriodStart ?? '?'} → ${sub.currentPeriodEnd ?? '?'}`)
console.log(`  cancelAtPeriodEnd: ${String(sub.cancelAtPeriodEnd ?? false)}`)
console.log(`  autoCharge: ${String(sub.autoCharge ?? false)}`)
