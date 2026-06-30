/**
 * Stage a one-time fully-free (100% off) next renewal cycle. The next renewal
 * consumes the comp exactly once, then billing resumes at full price. Rejected
 * on a `canceled` or `unpaid` subscription.
 *
 * Run:
 *   npx tsx examples/subscriptions/comp-next-cycle.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const sub = await client.subscriptions.compNextCycle('sub_replace_with_real_id')

console.log(`subscription ${sub.id ?? '?'} [${sub.status ?? '?'}]`)
console.log(`  next renewal (${sub.currentPeriodEnd ?? '?'}) will be comped`)
