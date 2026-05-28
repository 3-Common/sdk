/**
 * Iterate every active subscription, transparently fetching each page as
 * the previous one drains.
 *
 * Run:
 *   npx tsx examples/subscriptions/auto-paginate.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

let count = 0
let mrrMinor = 0
for await (const sub of client.subscriptions.listAutoPaginate({ status: 'active' })) {
  count += 1
  // Approximate MRR contribution as price × quantity. Replace with your own
  // pricing lookup if you need accurate currency-aware totals.
  mrrMinor += sub.quantity ?? 0
}

console.log(`iterated ${String(count)} active subscriptions`)
console.log(`approximate units in flight: ${String(mrrMinor)}`)
