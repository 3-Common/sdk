/**
 * Iterate every active feature in the catalog, transparently fetching each
 * page as the previous one drains.
 *
 * Run:
 *   npx tsx examples/features/auto-paginate.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

let count = 0
let quantity = 0
for await (const feature of client.features.listAutoPaginate({ active: true })) {
  count += 1
  if (feature.type === 'quantity') quantity += 1
}

console.log(`iterated ${String(count)} active features`)
console.log(`${String(quantity)} are quantity-typed`)
