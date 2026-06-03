/**
 * Retrieve a single feature by ID.
 *
 * Run:
 *   npx tsx examples/features/retrieve.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const feature = await client.features.retrieve('feat_replace_with_real_id')

console.log(`feature ${feature.id ?? '?'} [${feature.type ?? '?'}]`)
console.log(`  key   ${feature.key ?? '?'}`)
console.log(`  name  ${feature.name ?? '?'}`)
console.log(`  active ${String(feature.active ?? false)}`)
if (feature.enumValues !== undefined) {
  console.log(`  values ${feature.enumValues.join(', ')}`)
}
