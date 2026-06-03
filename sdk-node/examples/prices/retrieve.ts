/**
 * Retrieve a single price by ID, including its recurring cadence and feature
 * grants.
 *
 * Run:
 *   npx tsx examples/prices/retrieve.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const price = await client.prices.retrieve('price_replace_with_real_id')

console.log(`price ${price.id ?? '?'} [${price.type ?? '?'}]`)
console.log(`  product  ${price.productId ?? '?'}`)
console.log(`  amount   ${String(price.unitAmount ?? 0)} ${price.currency ?? ''}`)
if (price.recurring) {
  console.log(
    `  cadence  every ${String(price.recurring.intervalCount)} ${price.recurring.interval}`,
  )
}
for (const feature of price.features ?? []) {
  console.log(`  feature  ${feature.featureKey} [${feature.type}]`)
}
