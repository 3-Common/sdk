/**
 * Iterate every active price across all products, transparently fetching each
 * page as the previous one drains.
 *
 * Run:
 *   npx tsx examples/prices/auto-paginate.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

let count = 0
let recurring = 0
for await (const price of client.prices.listAutoPaginate({ active: true })) {
  count += 1
  if (price.type === 'recurring') recurring += 1
}

console.log(`iterated ${String(count)} active prices`)
console.log(`${String(recurring)} are recurring`)
