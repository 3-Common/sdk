/**
 * Create a quantity feature in the catalog. The `key` is the stable
 * identifier that prices and entitlements reference; `type` decides how the
 * feature resolves.
 *
 * Run:
 *   npx tsx examples/features/create.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const feature = await client.features.create({
  key: 'api_calls',
  name: 'API calls',
  type: 'quantity',
  description: 'Monthly API call quota',
  metadata: { category: 'usage' },
})

console.log(`created ${feature.id ?? '?'} — ${feature.key ?? '?'} [${feature.type ?? '?'}]`)
