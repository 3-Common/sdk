/**
 * Iterate every contact property, transparently fetching each page as the
 * previous one drains.
 *
 * Run:
 *   npx tsx examples/properties/auto-paginate.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

let count = 0
let archived = 0
for await (const property of client.properties.listAutoPaginate({ objectType: 'contact' })) {
  count += 1
  if (property.status === 'archived') archived += 1
}

console.log(`iterated ${String(count)} contact properties`)
console.log(`${String(archived)} are archived`)
