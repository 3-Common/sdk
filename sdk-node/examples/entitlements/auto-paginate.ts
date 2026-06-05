/**
 * Iterate every entitlement for a feature, transparently fetching each page
 * as the previous one drains. Handy for usage reports or sweeping for
 * low-balance customers.
 *
 * Run:
 *   npx tsx examples/entitlements/auto-paginate.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

let count = 0
let lowBalance = 0
for await (const ent of client.entitlements.listAutoPaginate({ featureKey: 'api_calls' })) {
  count += 1
  if ((ent.balance ?? 0) < 10) lowBalance += 1
}

console.log(`iterated ${String(count)} entitlements`)
console.log(`${String(lowBalance)} are running low (balance < 10)`)
