/**
 * Schedule cancellation at the end of the current period. The customer
 * retains access until `currentPeriodEnd`; the next renewal transitions
 * the subscription to `canceled` instead of advancing.
 *
 * Run:
 *   npx tsx examples/subscriptions/cancel.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const sub = await client.subscriptions.cancel('sub_replace_with_real_id', {
  reason: 'Customer requested via support ticket #4821',
})

console.log(`subscription ${sub.id ?? '?'} [${sub.status ?? '?'}]`)
console.log(`  cancelAtPeriodEnd: ${String(sub.cancelAtPeriodEnd ?? false)}`)
console.log(`  access continues until ${sub.currentPeriodEnd ?? '?'}`)
